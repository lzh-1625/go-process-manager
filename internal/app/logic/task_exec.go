package logic

import (
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/log"
)

type conditionFunc func(data *model.Task, proc *ProcessBase) bool

var conditionHandle = map[eum.Condition]conditionFunc{
	eum.TaskCondRunning: func(data *model.Task, proc *ProcessBase) bool {
		return proc.State.State == eum.ProcessStateRunning
	},
	eum.TaskCondNotRunning: func(data *model.Task, proc *ProcessBase) bool {
		return proc.State.State != eum.ProcessStateRunning && proc.State.State != eum.ProcessStateStart
	},
	eum.TaskCondException: func(data *model.Task, proc *ProcessBase) bool {
		return proc.State.State == eum.ProcessStateWarnning
	},
}

// 执行操作，返回结果是否成功
type operationFunc func(data *model.Task, proc *ProcessBase) bool

var OperationHandle = map[eum.TaskOperation]operationFunc{
	eum.TaskStart: func(data *model.Task, proc *ProcessBase) bool {
		if proc.State.State == eum.ProcessStateRunning || proc.State.State == eum.ProcessStateStart {
			log.Logger.Debugw("进程已在运行", "proc", proc.Name)
			return false
		}
		proc.Start()
		return true
	},

	eum.TaskStartWaitDone: func(data *model.Task, proc *ProcessBase) bool {
		if proc.State.State == eum.ProcessStateRunning || proc.State.State == eum.ProcessStateStart {
			log.Logger.Debugw("进程已在运行", "proc", proc.Name)
			return false
		}
		if err := proc.Start(); err != nil {
			log.Logger.Debugw("进程启动失败", "proc", proc.Name)
			return false
		}
		select {
		case <-proc.StopChan:
			log.Logger.Debugw("进程停止，任务完成", "proc", proc.Name)
			return true
		case <-time.After(time.Second * time.Duration(config.CF.TaskTimeout)):
			log.Logger.Errorw("任务超时")
			return false
		}
	},

	eum.TaskStop: func(data *model.Task, proc *ProcessBase) bool {
		if proc.State.State != eum.ProcessStateRunning {
			log.Logger.Debugw("进程未在运行", "proc", proc.Name)
			return false
		}
		log.Logger.Debugw("异步停止任务", "proc", proc.Name)
		proc.State.manualStopFlag = true
		go proc.Kill()
		return true
	},

	eum.TaskStopWaitDone: func(data *model.Task, proc *ProcessBase) bool {
		if proc.State.State != eum.ProcessStateRunning {
			log.Logger.Debugw("进程未在运行", "proc", proc.Name)
			return false
		}
		log.Logger.Debugw("停止任务并等待结束", "proc", proc.Name)
		proc.State.manualStopFlag = true
		return proc.Kill() == nil
	},
}
