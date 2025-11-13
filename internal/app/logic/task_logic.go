package logic

import (
	"context"
	"errors"
	"sync"

	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"
)

type taskLogic struct {
	taskJobMap sync.Map
}

var TaskLogic = new(taskLogic)

func (t *taskLogic) getTaskJob(id int) (*TaskJob, error) {
	c, ok := t.taskJobMap.Load(id)
	if !ok {
		return nil, errors.New("don't exist this task id")
	}
	return c.(*TaskJob), nil
}

func (t *taskLogic) InitTaskJob() {
	for _, v := range repository.TaskRepository.GetAllTask() {
		tj, err := NewTaskJob(v)
		if err != nil {
			log.Logger.Warnw("任务初始化失败", "err", err)
			continue
		}
		t.taskJobMap.Store(v.Id, tj)
	}
}

func (t *taskLogic) StopTaskJob(id int) error {
	taskJob, err := t.getTaskJob(id)
	if err != nil {
		return errors.New("don't exist this task id")
	}
	if taskJob.Running {
		taskJob.Cancel()
	}
	return nil
}

func (t *taskLogic) StartTaskJob(id int) error {
	taskJob, err := t.getTaskJob(id)
	if err != nil {
		return errors.New("don't exist this task id")
	}
	taskJob.Cron.Run()
	return nil
}

func (t *taskLogic) GetAllTaskJob() []model.TaskVo {
	result := repository.TaskRepository.GetAllTaskWithProcessName()
	for i, v := range result {
		task, err := t.getTaskJob(v.Id)
		if err != nil {
			continue
		}
		result[i].Id = task.TaskConfig.Id
		result[i].Running = task.Running
		result[i].Enable = task.TaskConfig.Enable
	}
	return result
}

func (t *taskLogic) DeleteTask(id int) (err error) {
	t.StopTaskJob(id)
	t.EditTaskEnable(id, false)
	t.taskJobMap.Delete(id)
	err = repository.TaskRepository.DeleteTask(id)
	if err != nil {
		return
	}
	return
}

func (t *taskLogic) CreateTask(data model.Task) error {
	tj, err := NewTaskJob(data)
	if err != nil {
		return err
	}
	taskId, err := repository.TaskRepository.AddTask(data)
	if err != nil {
		return err
	}
	tj.TaskConfig.Id = taskId
	t.taskJobMap.Store(taskId, tj)
	return nil
}

func (t *taskLogic) EditTask(data model.Task) error {
	tj, err := t.getTaskJob(data.Id)
	if err != nil {
		return errors.New("task id not exist")
	}

	if tj.Running {
		return errors.New("can't edit running task")
	}

	if tj.Cron != nil {
		tj.Cron.Stop()
		tj.Cron = nil
	}

	tj.TaskConfig = &data
	t.EditTaskEnable(data.Id, tj.TaskConfig.Enable)
	return repository.TaskRepository.EditTask(data)
}

func (t *taskLogic) EditTaskEnable(id int, status bool) error {
	tj, err := t.getTaskJob(id)
	if err != nil {
		return errors.New("don't exist this task id")
	}
	if tj.TaskConfig.CronExpression == "" {
		return errors.New("task cron expression is empty")
	}
	if err := tj.EditStatus(status); err != nil {
		return err
	}
	if err := repository.TaskRepository.EditTaskEnable(id, status); err != nil {
		return err
	}
	return nil
}

func (t *taskLogic) CreateApiKey(id int) error {
	data, err := repository.TaskRepository.GetTaskById(id)
	if err != nil {
		return err
	}
	key := utils.RandString(10)
	data.Key = &key
	repository.TaskRepository.EditTask(data)
	return nil
}

func (t *taskLogic) RunTaskByKey(key string) error {
	data, err := repository.TaskRepository.GetTaskByKey(key)
	if err != nil {
		return errors.New("don't exist key")
	}
	go t.RunTaskById(data.Id)
	return nil
}

func (t *taskLogic) RunTaskByTriggerEvent(processName string, event eum.ProcessState) {
	taskList := repository.TaskRepository.GetTriggerTask(processName, event)
	if len(taskList) == 0 {
		return
	}
	log.Logger.Infow("获取触发任务", "count", len(taskList), "prcess", processName, "触发事件", event)
	for _, v := range taskList {
		log.Logger.Infow("执行触发任务", "taskId", v.Id)
		t.RunTaskById(v.Id)
	}
}

func (t *taskLogic) RunTaskById(id int) error {
	task, err := t.getTaskJob(id)
	if err != nil {
		return errors.New("id不存在")
	}
	if task.Running {
		return errors.New("task is running")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	task.Cancel = cancel
	task.Run(ctx)
	return nil
}
