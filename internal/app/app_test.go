package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/process"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
	"go.uber.org/fx"
)

func TestProcess(t *testing.T) {
	config.CF.ConfigDir = t.TempDir()

	ctx, cancel := context.WithCancel(context.Background())
	app := NewApp(fx.Invoke(func(processCtlLogic *logic.ProcessCtlLogic,
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
	app := NewApp(fx.Invoke(func(
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
	app := NewApp(fx.Invoke(func(
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

		}))
	}))
	app.Start(ctx)
}
