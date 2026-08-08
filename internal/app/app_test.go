package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/process"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
	"github.com/lzh-1625/go_process_manager/utils"
	"go.uber.org/fx"
)

func TestProcess(t *testing.T) {
	config.CF.ConfigDir = t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	app := NewApp(ctx, fx.Invoke(func(processCtlLogic *logic.ProcessCtlLogic,
		lc fx.Lifecycle) {
		lc.Append(fx.StartHook(func() {
			defer cancel()
			proc, err := processCtlLogic.CreateProcess(model.Process{
				Name: "test1",
				Cmd:  "sleep 100",
				Cwd:  "./",
			})
			if err != nil {
				t.Errorf("failed to create process %q: %v", "test1", err)
				return
			}
			_, err = processCtlLogic.CreateProcess(model.Process{
				Name: "test1",
				Cmd:  "sleep 100",
				Cwd:  "./",
			})
			if err == nil {
				t.Errorf("expected duplicate process creation for %q to fail, but it succeeded", "test1")
				return
			}
			if err := proc.Start(); err != nil {
				t.Errorf("failed to start process %q: %v", proc.Name, err)
				return
			}
			if err := processCtlLogic.DeleteProcess(proc.UUID); err == nil {
				t.Errorf("expected deleting running process %q to fail, but it succeeded", proc.Name)
				return
			}
			if err := proc.Kill(); err != nil {
				t.Errorf("failed to stop process %q: %v", proc.Name, err)
				return
			}
			if err := proc.Kill(); err == nil {
				t.Errorf("expected stopping already stopped process %q to fail, but it succeeded", proc.Name)
				return
			}
			if err := processCtlLogic.DeleteProcess(proc.UUID); err != nil {
				t.Errorf("failed to delete stopped process %q: %v", proc.Name, err)
				return
			}
			proc1, err := processCtlLogic.CreateProcess(model.Process{
				Name: "test1",
				Cmd:  "sleep 10",
				Cwd:  "./",
			})
			if err != nil {
				t.Errorf("failed to create process %q: %v", "test1", err)
				return
			}
			proc2, err := processCtlLogic.CreateProcess(model.Process{
				Name: "test2",
				Cmd:  "sleep 10",
				Cwd:  "./",
			})
			if err != nil {
				t.Errorf("failed to create process %q: %v", "test2", err)
				return
			}

			processCtlLogic.ForEach(func(proc *process.Process) {
				if err := proc.Start(); err != nil {
					t.Errorf("failed to start process %q during bulk start: %v", proc.Name, err)
					return
				}
			})

			if proc1.State.State != types.ProcessStateRunning || proc2.State.State != types.ProcessStateRunning {
				t.Errorf("expected both processes to be running after bulk start; got %q=%v, %q=%v", proc1.Name, proc1.State.State, proc2.Name, proc2.State.State)
				return
			}
			processCtlLogic.ForEach(func(proc *process.Process) {
				if err := proc.Kill(); err != nil {
					t.Errorf("failed to stop process %q during bulk stop: %v", proc.Name, err)
					return
				}
			})

			if proc1.State.State != types.ProcessStateStopped || proc2.State.State != types.ProcessStateStopped {
				t.Errorf("expected both processes to be stopped after bulk stop; got %q=%v, %q=%v", proc1.Name, proc1.State.State, proc2.Name, proc2.State.State)
				return
			}
		}))
	}))
	app.Start(ctx)
}

func TestEvent(t *testing.T) {
	config.CF.ConfigDir = t.TempDir()
	config.CF.KillWaitTime = 1
	ctx, cancel := context.WithCancel(context.Background())
	app := NewApp(ctx, fx.Invoke(func(
		processCtlLogic *logic.ProcessCtlLogic,
		eventLogic *logic.EventLogic,
		pushLogic *logic.PushLogic,
		lc fx.Lifecycle) {
		lc.Append(fx.StartHook(func() {
			defer cancel()
			proc, err := processCtlLogic.CreateProcess(model.Process{
				Name: "test1",
				Cmd:  "sleep 100",
				Cwd:  "./",
			})
			if err != nil {
				t.Errorf("failed to create process %q: %v", "test1", err)
				return
			}
			if err := proc.Operate("test-user", func() error {
				return proc.Start()
			}); err != nil {
				t.Errorf("failed to start process %q as %q: %v", proc.Name, "test-user", err)
				return
			}
			data, count, err := eventLogic.Get(model.EventListReq{Page: 1, Size: 1000})
			if err != nil {
				t.Errorf("failed to list process events after start: %v", err)
				return
			}
			if count != 1 || len(data) != 1 {
				t.Errorf("expected 1 event after starting the process; got count=%d, records=%d", count, len(data))
				return
			}
			if data[0].Name != "test1" || data[0].User != "test-user" || data[0].Type != types.EventProcessStart {
				t.Errorf("unexpected process-start event: got %+v", data[0])
				return
			}
			if err := proc.Operate("test-user-1", func() error {
				return proc.Kill()
			}); err != nil {
				t.Errorf("failed to stop process %q as %q: %v", proc.Name, "test-user-1", err)
				return
			}
			data, count, err = eventLogic.Get(model.EventListReq{Page: 1, Size: 1000})
			if err != nil {
				t.Errorf("failed to list process events after stop: %v", err)
				return
			}
			if count != 2 || len(data) != 2 {
				t.Errorf("expected 2 events after stopping the process; got count=%d, records=%d", count, len(data))
				return
			}
			if data[0].Name != "test1" || data[0].User != "test-user-1" || data[0].Type != types.EventProcessStop {
				t.Errorf("unexpected process-stop event: got %+v", data[0])
				return
			}
			var rwc io.ReadWriteCloser
			proc.AddWriter("test-ws", rwc)
			proc.DeleteWriter("test-ws")
			data, count, err = eventLogic.Get(model.EventListReq{Page: 1, Size: 1000})
			if err != nil {
				t.Errorf("failed to list process events after WebSocket connection: %v", err)
				return
			}
			if count != 3 || len(data) != 3 {
				t.Errorf("expected 3 events after the WebSocket connection; got count=%d, records=%d", count, len(data))
				return
			}
			if data[0].Name != "test1" || data[0].User != "test-ws" || data[0].Type != types.EventProcessConnect {
				t.Errorf("unexpected process-connect event: got %+v", data[0])
				return
			}
			if err := pushLogic.AddPushConfig(model.Push{Method: "GET", Url: "http://" + config.CF.Listen, Enable: true, Remark: "push-test-1"}); err != nil {
				t.Errorf("failed to add enabled push configuration: %v", err)
				return
			}
			if err := pushLogic.AddPushConfig(model.Push{Method: "GET", Url: "http://" + config.CF.Listen, Enable: false, Remark: "push-test-2"}); err != nil {
				t.Errorf("failed to add disabled push configuration: %v", err)
				return
			}
			pushItems := pushLogic.GetPushList()
			if len(pushItems) != 2 {
				t.Errorf("expected 2 push configurations; got %d: %+v", len(pushItems), pushItems)
				return
			}
			pushIDs := []int{}
			for _, v := range pushItems {
				pushIDs = append(pushIDs, int(v.ID))
			}
			pushIDsByte, _ := json.Marshal(pushIDs)
			if err := processCtlLogic.UpdateProcessConfig(model.Process{
				UUID:    proc.UUID,
				Name:    "test1",
				Cmd:     "sleep 100",
				Cwd:     "./",
				PushIDs: string(pushIDsByte),
			}); err != nil {
				t.Errorf("failed to update process %q with push configuration IDs %s: %v", proc.Name, pushIDsByte, err)
				return
			}
			proc.Start()
			time.Sleep(time.Second)
			data, count, err = eventLogic.Get(model.EventListReq{Page: 1, Size: 1000, Type: types.EventPush})
			if err != nil {
				t.Errorf("failed to list push events: %v", err)
				return
			}
			if count != 1 || len(data) != 1 {
				t.Errorf("expected 1 push event; got count=%d, records=%d", count, len(data))
				return
			}
			var enableID int
			for _, v := range pushItems {
				if v.Enable {
					enableID = int(v.ID)
					break
				}
			}
			if data[0].Name != fmt.Sprintf("ID:%d", enableID) || data[0].User != "" || data[0].Type != types.EventPush {
				t.Errorf("unexpected push event for enabled push ID %d: got %+v", enableID, data[0])
				return
			}

		}))
	}))
	app.Start(ctx)
}

func TestTask(t *testing.T) {
	config.CF.ConfigDir = t.TempDir()
	config.CF.KillWaitTime = 1
	ctx, cancel := context.WithCancel(context.Background())
	app := NewApp(ctx, fx.Invoke(func(
		processCtlLogic *logic.ProcessCtlLogic,
		eventLogic *logic.EventLogic,
		taskLogic *logic.TaskLogic,
		lc fx.Lifecycle) {
		lc.Append(fx.StartHook(func() {
			defer cancel()
			proc1, err := processCtlLogic.CreateProcess(model.Process{
				Name: "test1",
				Cmd:  "sleep 10",
				Cwd:  "./",
			})
			if err != nil {
				t.Errorf("failed to create process %q: %v", "test1", err)
				return
			}
			proc2, err := processCtlLogic.CreateProcess(model.Process{
				Name: "test2",
				Cmd:  "sleep 10",
				Cwd:  "./",
			})
			if err != nil {
				t.Errorf("failed to create process %q: %v", "test2", err)
				return
			}

			err = taskLogic.CreateTask(model.Task{
				Name:            "task2",
				Condition:       types.TaskCondPass,
				NextID:          nil,
				Operation:       types.TaskStart,
				OperationTarget: proc2.UUID,
			})
			if err != nil {
				t.Errorf("failed to create task %q for process %q: %v", "task2", proc2.Name, err)
				return
			}
			tasks := taskLogic.GetAllTaskJob()
			if len(tasks) == 0 || tasks[0].Name != "task2" {
				t.Errorf("expected first task to be %q; got %+v", "task2", tasks)
				return
			}
			taskLogic.CreateTask(model.Task{
				Name:            "task1",
				Condition:       types.TaskCondPass,
				NextID:          &tasks[0].ID,
				Operation:       types.TaskStart,
				OperationTarget: proc1.UUID,
			})
			tasks = taskLogic.GetAllTaskJob()
			if len(tasks) != 2 || tasks[1].Name != "task1" {
				t.Errorf("expected tasks [task2 task1]; got %+v", tasks)
				return
			}
			err = taskLogic.RunTaskByID(context.TODO(), tasks[1].ID)
			if err != nil {
				t.Errorf("failed to run task %q (ID %d): %v", tasks[1].Name, tasks[1].ID, err)
				return
			}
			if proc1.State.State != types.ProcessStateRunning || proc2.State.State != types.ProcessStateRunning {
				t.Errorf("expected task chain to start both processes; got %q=%v, %q=%v", proc1.Name, proc1.State.State, proc2.Name, proc2.State.State)
				return
			}
			processCtlLogic.ForEach(func(proc *process.Process) {
				proc.Kill()
			})
			taskLogic.CreateTask(model.Task{
				Name:            "task3",
				Operation:       types.TaskStart,
				OperationTarget: proc1.UUID,
				TriggerEvent:    new(types.ProcessStateRunning),
				TriggerTarget:   &proc2.UUID,
			})
			proc2.Start()
			time.Sleep(time.Second)
			if proc1.State.State != types.ProcessStateRunning {
				t.Errorf("expected process %q to be running after its trigger event; got %v", proc1.Name, proc1.State.State)
				return
			}

			const (
				cronTaskName     = "cron-test"
				cronExpression   = "* * * * * *"
				expectedOutput   = "123qwe"
				cronWait         = 1500 * time.Millisecond
				disabledCronWait = 2500 * time.Millisecond
			)
			tempFilePath := path.Join(t.TempDir(), "tempFile")
			cronProc, err := processCtlLogic.CreateProcess(model.Process{
				Name: "echo",
				Cmd:  fmt.Sprintf(`sh -c "echo %s > %s"`, expectedOutput, tempFilePath),
				Cwd:  "./",
			})
			if err != nil {
				t.Errorf("failed to create cron process %q for task %q: %v", "echo", cronTaskName, err)
				return
			}

			err = taskLogic.CreateTask(model.Task{
				ID:              10000,
				Operation:       types.TaskStart,
				OperationTarget: cronProc.UUID,
				Name:            cronTaskName,
				CronExpression:  cronExpression,
				Enable:          true,
			})
			if err != nil {
				t.Errorf("failed to create enabled cron task %q with expression %q: %v", cronTaskName, cronExpression, err)
				return
			}
			time.Sleep(cronWait)
			b, err := os.ReadFile(tempFilePath)
			if err != nil {
				t.Errorf("failed to read cron output %q after waiting %s for task %q: %v", tempFilePath, cronWait, cronTaskName, err)
				return
			}
			if actualOutput := strings.TrimSuffix(string(b), "\n"); actualOutput != expectedOutput {
				t.Errorf("unexpected cron output in %q for task %q: got %q, want %q", tempFilePath, cronTaskName, actualOutput, expectedOutput)
				return
			}
			if err := os.Remove(tempFilePath); err != nil {
				t.Errorf("failed to remove cron output %q before disabling task %q: %v", tempFilePath, cronTaskName, err)
				return
			}
			if err := taskLogic.EditTask(&model.Task{
				ID:              10000,
				Operation:       types.TaskStart,
				OperationTarget: cronProc.UUID,
				Name:            cronTaskName,
				CronExpression:  cronExpression,
				Enable:          false,
			}); err != nil {
				t.Errorf("failed to disable cron task %q: %v", cronTaskName, err)
				return
			}
			time.Sleep(disabledCronWait)
			_, err = os.Stat(tempFilePath)
			if !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("expected disabled cron task %q not to recreate output %q after waiting %s; stat error: %v", cronTaskName, tempFilePath, disabledCronWait, err)
				return
			}
			err = taskLogic.EditTask(&model.Task{
				ID:              10000,
				Operation:       types.TaskStart,
				OperationTarget: cronProc.UUID,
				Name:            cronTaskName,
				CronExpression:  cronExpression,
				Enable:          true,
			})
			if err != nil {
				t.Errorf("failed to re-enable cron task %q: %v", cronTaskName, err)
				return
			}
			time.Sleep(cronWait)
			b, err = os.ReadFile(tempFilePath)
			if err != nil {
				t.Errorf("failed to read cron output %q after re-enabling task %q and waiting %s: %v", tempFilePath, cronTaskName, cronWait, err)
				return
			}
			if actualOutput := strings.TrimSuffix(string(b), "\n"); actualOutput != expectedOutput {
				t.Errorf("unexpected cron output in %q after re-enabling task %q: got %q, want %q", tempFilePath, cronTaskName, actualOutput, expectedOutput)
				return
			}
		}))
	}))
	app.Start(ctx)
}

func TestApi(t *testing.T) {
	config.CF.ConfigDir = t.TempDir()
	config.CF.KillWaitTime = 1
	ctx, cancel := context.WithCancel(context.Background())
	app := NewApp(ctx, fx.Invoke(func(
		e *echo.Echo,
		lc fx.Lifecycle) {
		lc.Append(fx.StartHook(func() {
			defer cancel()
			resp := req[map[string]any](http.MethodPost, e, "/api/user/login", "", model.LoginHandlerReq{
				Account:  logic.DefaultRootAccount,
				Password: logic.DefaultRootPassword,
			})
			token := resp.Data["token"].(string)

			req[model.User](http.MethodPost, e, "/api/user", token, model.User{
				Account:  "admin",
				Password: "admin",
				Role:     types.RoleAdmin,
			})
			resp1 := req[map[string]any](http.MethodPost, e, "/api/user/login", token, model.LoginHandlerReq{
				Account:  "admin",
				Password: "admin",
			})
			adminToken := resp1.Data["token"].(string)

			req[model.User](http.MethodPost, e, "/api/user", token, model.User{
				Account:  "user",
				Password: "user",
				Role:     types.RoleUser,
			})
			resp2 := req[map[string]any](http.MethodPost, e, "/api/user/login", "", model.LoginHandlerReq{
				Account:  "user",
				Password: "user",
			})

			userToken := resp2.Data["token"].(string)

			pid1 := req[map[string]int32](http.MethodPost, e, "/api/process/config", token, model.Process{
				Name:      "test1",
				Cmd:       `cmd /c "echo 111\n111\n"`,
				LogReport: true,
			}).Data["uuid"]
			_ = req[map[string]int32](http.MethodPost, e, "/api/process/config", token, model.Process{
				Name: "test2",
				Cmd:  "echo 222",
			}).Data["uuid"]

			processinfo := req[[]model.ProcessInfo](http.MethodGet, e, "/api/process", adminToken, struct{}{})
			if len(processinfo.Data) != 2 {
				t.Error("")
			}

			processinfo = req[[]model.ProcessInfo](http.MethodGet, e, "/api/process", userToken, struct{}{})
			if len(processinfo.Data) != 0 {
				t.Error("")
			}

			respStart := req[any](http.MethodPut, e, "/api/process", userToken, map[string]any{"uuid": pid1})
			if respStart.Code == 0 {
				t.Error("")
			}

			req[model.User](http.MethodPut, e, "/api/permission", token, model.Permission{
				Account: "user",
				Pid:     pid1,
				Owned:   true,
				Start:   true,
			})

			processinfo = req[[]model.ProcessInfo](http.MethodGet, e, "/api/process", userToken, struct{}{})
			if len(processinfo.Data) != 1 || processinfo.Data[0].UUID != int(pid1) {
				t.Error("")
			}

			respStart = req[any](http.MethodPut, e, "/api/process", userToken, map[string]any{"uuid": pid1})
			if respStart.Code != 0 {
				t.Error("")
			}
			time.Sleep(time.Second * 3)

			logReq := model.GetLogReq{}
			logReq.Page.Size = 100
			logResp := req[model.LogResp](http.MethodPost, e, "/api/log", userToken, logReq)
			if len(logResp.Data.Data) != 0 {
				t.Error("")
			}

			req[model.User](http.MethodPut, e, "/api/permission", token, model.Permission{
				Account: "user",
				Pid:     pid1,
				Owned:   true,
				Start:   true,
				Log:     true,
			})

			logResp = req[model.LogResp](http.MethodPost, e, "/api/log", userToken, logReq)
			if len(logResp.Data.Data) == 0 {
				t.Error("")
			}
		}))
	}))
	app.Start(ctx)
}

func req[T any](method string, e *echo.Echo, uri string, token string, body any) model.Response[T] {
	req := httptest.NewRequest(
		method,
		uri,
		bytes.NewBufferString(utils.StructToJsonStr(body)),
	)
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", token)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	resp := model.Response[T]{}
	json.NewDecoder(rec.Body).Decode(&resp)
	return resp
}
