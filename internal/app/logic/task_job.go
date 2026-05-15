package logic

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/process"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/robfig/cron/v3"
)

type TaskJob struct {
	Cron       *cron.Cron         `json:"-"`
	TaskConfig *model.Task        `json:"task"`
	Running    bool               `json:"running"`
	Cancel     context.CancelFunc `json:"-"`
	StartTime  time.Time          `json:"startTime"`
	EndTime    time.Time          `json:"endTime"`

	eventLogic      *EventLogic
	processCtlLogic *ProcessCtlLogic
	taskLogic       *TaskLogic
}

func NewTaskJob(data *model.Task, eventLogic *EventLogic, processCtlLogic *ProcessCtlLogic, taskLogic *TaskLogic) (*TaskJob, error) {
	tj := &TaskJob{
		TaskConfig:      data,
		StartTime:       time.Now(),
		eventLogic:      eventLogic,
		processCtlLogic: processCtlLogic,
		taskLogic:       taskLogic,
	}
	if data.Enable {
		if err := tj.InitCronHandle(); err != nil {
			log.Logger.Warnw("cron task start failed", "err", err, "task", data.ID)
		}
	}
	return tj, nil
}

func (t *TaskJob) Run(ctx context.Context) (err error) {
	logger := log.Logger.With("taskID", t.TaskConfig.ID)
	if ctx.Value(eum.CtxTaskTraceID{}) == nil {
		ctx = context.WithValue(ctx, eum.CtxTaskTraceID{}, uuid.NewString())
	}
	t.eventLogic.Create(t.TaskConfig.Name, eum.EventTaskStart, "traceID", ctx.Value(eum.CtxTaskTraceID{}).(string))
	defer func() {
		t.eventLogic.Create(t.TaskConfig.Name, eum.EventTaskStop, "traceID", ctx.Value(eum.CtxTaskTraceID{}).(string), "success", strconv.FormatBool(err == nil), "time", time.Since(t.StartTime).String())
	}()
	logger = logger.With("traceID", ctx.Value(eum.CtxTaskTraceID{}).(string))
	t.Running = true
	t.StartTime = time.Now()
	TaskWaitCond.Trigger()
	defer func() {
		t.Running = false
		TaskWaitCond.Trigger()
	}()

	proc, err := t.processCtlLogic.GetProcess(t.TaskConfig.OperationTarget)
	if err != nil {
		logger.Debugw("process not found, task execution failed")
		return err
	}

	var ok bool
	// check if the condition is satisfied
	if t.TaskConfig.Condition == eum.TaskCondPass || t.TaskConfig.ProcessID == 0 {
		ok = true
	} else {
		ok = conditionHandle[t.TaskConfig.Condition](t.TaskConfig, proc)
	}
	logger.Debugw("task condition check", "pass", ok)
	if !ok {
		return
	}
	// execute operation
	logger.Infow("task execute started")
	if !operationHandle[t.TaskConfig.Operation](t.TaskConfig, proc) {
		logger.Warnw("task execution failed")
		return errors.New("task execution failed")
	}
	logger.Infow("task execution success", "target", t.TaskConfig.OperationTarget)

	if t.TaskConfig.NextID != nil {
		var nextTask *TaskJob
		nextTask, err = t.taskLogic.getTaskJob(*t.TaskConfig.NextID)
		if err != nil {
			logger.Warnw("cannot get next node, task execution failed", "nextID", t.TaskConfig.NextID)
			return err
		}
		select {
		case <-ctx.Done():
			logger.Infow("task flow manually ended")
		default:
			logger.Debugw("execute next node", "nextID", t.TaskConfig.NextID)
			if nextTask.Running {
				logger.Warnw("next node is running, task execution failed", "nextID", t.TaskConfig.NextID)
				return
			}
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			nextTask.Cancel = cancel
			return nextTask.Run(ctx)
		}
	} else {
		logger.Infow("task flow ended")
	}
	return
}

func (t *TaskJob) InitCronHandle() error {
	if _, err := cron.ParseStandard(t.TaskConfig.CronExpression); err != nil { // cron expression validation
		log.Logger.Errorw("cron parse failed", "cron", t.TaskConfig.CronExpression, "err", err)
		return err
	}
	c := cron.New()
	_, err := c.AddFunc(t.TaskConfig.CronExpression, func() {
		log.Logger.Infow("cron task start")
		if t.Running {
			log.Logger.Infow("task is running, skip current task")
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		t.Cancel = cancel
		t.Run(ctx)
		log.Logger.Infow("cron task ended")
	})
	if err != nil {
		return err
	}
	c.Start()
	t.Cron = c
	return nil
}

func (t *TaskJob) EditStatus(status bool) error {
	if t.Cron != nil { // stop cron
		t.Cron.Stop()
	}
	if status { // start cron
		if err := t.InitCronHandle(); err != nil {
			return err
		}
	}
	return nil
}

type conditionFunc func(data *model.Task, proc *process.ProcessPty) bool

var conditionHandle = map[eum.Condition]conditionFunc{
	eum.TaskCondRunning: func(data *model.Task, proc *process.ProcessPty) bool {
		return proc.State.State == eum.ProcessStateRunning
	},
	eum.TaskCondNotRunning: func(data *model.Task, proc *process.ProcessPty) bool {
		state := proc.State.State
		return state != eum.ProcessStateRunning && state != eum.ProcessStateStart
	},
	eum.TaskCondException: func(data *model.Task, proc *process.ProcessPty) bool {
		return proc.State.State == eum.ProcessStateWarnning
	},
}

// execute operation, return result whether successfully
type operationFunc func(*model.Task, *process.ProcessPty) bool

var operationHandle = map[eum.TaskOperation]operationFunc{
	eum.TaskStart: func(data *model.Task, proc *process.ProcessPty) bool {
		state := proc.State.State
		if state == eum.ProcessStateRunning || state == eum.ProcessStateStart {
			log.Logger.Debugw("process is running", "proc", proc.Name)
			return false
		}
		proc.Start()
		return true
	},

	eum.TaskStartWaitDone: func(data *model.Task, proc *process.ProcessPty) bool {
		state := proc.State.State
		if state == eum.ProcessStateRunning || state == eum.ProcessStateStart {
			log.Logger.Debugw("process is running", "proc", proc.Name)
			return false
		}
		if err := proc.Start(); err != nil {
			log.Logger.Debugw("process start failed", "proc", proc.Name)
			return false
		}
		select {
		case <-proc.StopChan:
			log.Logger.Debugw("process stopped, task completed", "proc", proc.Name)
			return true
		case <-time.After(time.Second * time.Duration(config.CF.TaskTimeout)):
			log.Logger.Errorw("task timeout")
			return false
		}
	},

	eum.TaskStop: func(data *model.Task, proc *process.ProcessPty) bool {
		if proc.State.State != eum.ProcessStateRunning {
			log.Logger.Debugw("process is not running", "proc", proc.Name)
			return false
		}
		log.Logger.Debugw("async stop task", "proc", proc.Name)
		go proc.Kill()
		return true
	},

	eum.TaskStopWaitDone: func(data *model.Task, proc *process.ProcessPty) bool {
		if proc.State.State != eum.ProcessStateRunning {
			log.Logger.Debugw("process is not running", "proc", proc.Name)
			return false
		}
		log.Logger.Debugw("stop task and wait done", "proc", proc.Name)
		return proc.Kill() == nil
	},
}
