package logic

import (
	"time"

	"github.com/lzh-1625/go_process_manager/config"
	"github.com/lzh-1625/go_process_manager/internal/app/constants"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/log"
)

type conditionFunc func(data *model.Task, proc *ProcessBase) bool

var conditionHandle = map[constants.Condition]conditionFunc{
	constants.RUNNING: func(data *model.Task, proc *ProcessBase) bool {
		return proc.State.State == 1
	},
	constants.NOT_RUNNING: func(data *model.Task, proc *ProcessBase) bool {
		return proc.State.State != 1
	},
	constants.EXCEPTION: func(data *model.Task, proc *ProcessBase) bool {
		return proc.State.State == 2
	},
}

// 执行操作，返回结果是否成功
type operationFunc func(data *model.Task, proc *ProcessBase) bool

var OperationHandle = map[constants.TaskOperation]operationFunc{
	constants.TASK_START: func(data *model.Task, proc *ProcessBase) bool {
		if proc.State.State == 1 {
			log.Logger.Debugw("进程已在运行")
			return false
		}
		go proc.Start()
		return true
	},

	constants.TASK_START_WAIT_DONE: func(data *model.Task, proc *ProcessBase) bool {
		if proc.State.State == 1 {
			log.Logger.Debugw("进程已在运行")
			return false
		}
		if err := proc.Start(); err != nil {
			log.Logger.Debugw("进程启动失败")
			return false
		}
		select {
		case <-proc.StopChan:
			log.Logger.Debugw("进程停止，任务完成")
			return true
		case <-time.After(time.Second * time.Duration(config.CF.TaskTimeout)):
			log.Logger.Errorw("任务超时")
			return false
		}
	},

	constants.TASK_STOP: func(data *model.Task, proc *ProcessBase) bool {
		if proc.State.State != 1 {
			log.Logger.Debugw("进程未在运行")
			return false
		}
		log.Logger.Debugw("异步停止任务")
		proc.State.manualStopFlag = true
		go proc.Kill()
		return true
	},

	constants.TASK_STOP_WAIT_DONE: func(data *model.Task, proc *ProcessBase) bool {
		if proc.State.State != 1 {
			log.Logger.Debugw("进程未在运行")
			return false
		}
		log.Logger.Debugw("停止任务并等待结束")
		proc.State.manualStopFlag = true
		return proc.Kill() == nil
	},
}
