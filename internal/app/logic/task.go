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
	if data.Enable && data.CronExpression != "" {
		err := tj.InitCronHandle()
		if err != nil {
			log.Logger.Warnw("定时任务启动失败", "err", err, "task", data.Id)
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
	var ok bool
	// 判断条件是否满足
	if t.TaskConfig.Condition == eum.TaskCondPass || t.TaskConfig.ProcessId == 0 {
		ok = true
	} else {
		proc, err := ProcessCtlLogic.GetProcess(t.TaskConfig.OperationTarget)
		if err != nil {
			return err
		}
		ok = conditionHandle[t.TaskConfig.Condition](t.TaskConfig, proc)
	}
	log.Logger.Debugw("任务条件判断", "pass", ok)
	if !ok {
		return
	}

	proc, err := ProcessCtlLogic.GetProcess(t.TaskConfig.OperationTarget)
	if err != nil {
		log.Logger.Debugw("不存在该进程，结束任务")
		return err
	}

	// 执行操作
	log.Logger.Infow("任务开始执行")
	if !OperationHandle[t.TaskConfig.Operation](t.TaskConfig, proc) {
		log.Logger.Warnw("任务执行失败")
		return errors.New("task execute failed")
	}
	log.Logger.Infow("任务执行成功", "target", t.TaskConfig.OperationTarget)

	if t.TaskConfig.NextId != nil {
		var nextTask *TaskJob
		nextTask, err = TaskLogic.getTaskJob(*t.TaskConfig.NextId)
		if err != nil {
			log.Logger.Errorw("无法获取到下一个节点,结束任务", "nextId", t.TaskConfig.NextId)
			return err
		}
		select {
		case <-ctx.Done():
			log.Logger.Infow("任务流被手动结束")
		default:
			log.Logger.Debugw("执行下一个节点", "nextId", *t.TaskConfig.NextId)
			if nextTask.Running {
				log.Logger.Errorw("下一个节点已在运行，结束任务", "nextId", t.TaskConfig.NextId)
				return
			}
			return nextTask.Run(ctx)
		}
	} else {
		log.Logger.Infow("任务流结束")
	}
	return
}

func (t *TaskJob) InitCronHandle() error {
	if _, err := cron.ParseStandard(t.TaskConfig.CronExpression); err != nil { // cron表达式校验
		log.Logger.Errorw("cron解析失败", "cron", t.TaskConfig.CronExpression, "err", err)
		return err
	}
	c := cron.New()
	_, err := c.AddFunc(t.TaskConfig.CronExpression, func() {
		log.Logger.Infow("定时任务启动")
		if t.Running {
			log.Logger.Infow("任务已在运行，跳过当前任务")
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		t.Cancel = cancel
		t.Run(ctx)
		log.Logger.Infow("定时任务结束")
	})
	if err != nil {
		return err
	}
	c.Start()
	t.Cron = c
	return nil
}

func (t *TaskJob) EditStatus(status bool) error {
	if t.Cron != nil && !status {
		t.Cron.Stop()
	} else if err := t.InitCronHandle(); err != nil {
		return err
	}
	t.TaskConfig.Enable = status
	return nil
}
