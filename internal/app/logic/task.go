package logic

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
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
}

func NewTaskJob(data *model.Task) (*TaskJob, error) {
	tj := &TaskJob{
		TaskConfig: data,
		StartTime:  time.Now(),
	}
	if data.Enable {
		if err := tj.InitCronHandle(); err != nil {
			log.Logger.Warnw("cron task start failed", "err", err, "task", data.ID)
		}
	}
	return tj, nil
}

func (t *TaskJob) Run(ctx context.Context) (err error) {
	if ctx.Value(eum.CtxTaskTraceId{}) == nil {
		ctx = context.WithValue(ctx, eum.CtxTaskTraceId{}, uuid.NewString())
	}
	EventLogic.Create(t.TaskConfig.Name, eum.EventTaskStart, "traceId", ctx.Value(eum.CtxTaskTraceId{}).(string))
	defer func() {
		EventLogic.Create(t.TaskConfig.Name, eum.EventTaskStop, "traceId", ctx.Value(eum.CtxTaskTraceId{}).(string), "success", strconv.FormatBool(err == nil), "time", time.Since(t.StartTime).String())
	}()
	t.Running = true
	t.StartTime = time.Now()
	TaskWaitCond.Trigger()
	defer func() {
		t.Running = false
		TaskWaitCond.Trigger()
	}()

	proc, err := ProcessCtlLogic.GetProcess(t.TaskConfig.OperationTarget)
	if err != nil {
		log.Logger.Debugw("process not found, end task")
		return err
	}

	var ok bool
	// check if the condition is satisfied
	if t.TaskConfig.Condition == eum.TaskCondPass || t.TaskConfig.ProcessId == 0 {
		ok = true
	} else {
		ok = conditionHandle[t.TaskConfig.Condition](t.TaskConfig, proc)
	}
	log.Logger.Debugw("task condition check", "pass", ok)
	if !ok {
		return
	}
	// execute operation
	log.Logger.Infow("task execute started")
	if !GetOperationHandle()[t.TaskConfig.Operation](t.TaskConfig, proc) {
		log.Logger.Warnw("task execution failed")
		return errors.New("task execution failed")
	}
	log.Logger.Infow("task execution success", "target", t.TaskConfig.OperationTarget)

	if t.TaskConfig.NextId != nil {
		var nextTask *TaskJob
		nextTask, err = TaskLogic.getTaskJob(*t.TaskConfig.NextId)
		if err != nil {
			log.Logger.Errorw("cannot get next node, end task", "nextId", t.TaskConfig.NextId)
			return err
		}
		select {
		case <-ctx.Done():
			log.Logger.Infow("task flow manually ended")
		default:
			log.Logger.Debugw("execute next node", "nextId", *t.TaskConfig.NextId)
			if nextTask.Running {
				log.Logger.Errorw("next node is running, end task", "nextId", t.TaskConfig.NextId)
				return
			}
			return nextTask.Run(ctx)
		}
	} else {
		log.Logger.Infow("task flow ended")
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
	t.TaskConfig.Enable = status
	return nil
}
