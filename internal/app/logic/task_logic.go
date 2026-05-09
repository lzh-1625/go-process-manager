package logic

import (
	"context"
	"errors"
	"sync"
	"time"

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
			log.Logger.Warnw("task initialization failed", "err", err)
			continue
		}
		t.taskJobMap.Store(v.ID, tj)
	}
}

func (t *taskLogic) StopTaskJob(id int) error {
	taskJob, err := t.getTaskJob(id)
	if err != nil {
		return err
	}
	if taskJob.Running {
		taskJob.Cancel()
	}
	if taskJob.Cron != nil {
		taskJob.Cron.Stop()
	}
	return nil
}

func (t *taskLogic) StartTaskJob(id int) error {
	taskJob, err := t.getTaskJob(id)
	if err != nil {
		return err
	}
	taskJob.Cron.Run()
	return nil
}

func (t *taskLogic) GetAllTaskJob() []model.TaskVo {
	result := repository.TaskRepository.GetAllTaskWithProcessName()
	for i, v := range result {
		task, err := t.getTaskJob(v.ID)
		if err != nil {
			continue
		}
		result[i].ID = task.TaskConfig.ID
		result[i].Running = task.Running
		result[i].Enable = task.TaskConfig.Enable
		result[i].StartTime = task.StartTime.Format(time.DateTime)
	}
	return result
}

func (t *taskLogic) GetTaskByID(id int) (*model.Task, error) {
	return repository.TaskRepository.GetTaskByID(id)
}

func (t *taskLogic) DeleteTask(id int) (err error) {
	t.StopTaskJob(id)
	t.taskJobMap.Delete(id)
	err = repository.TaskRepository.DeleteTask(id)
	if err != nil {
		return
	}
	return
}

func (t *taskLogic) CreateTask(data model.Task) error {
	tj, err := NewTaskJob(&data)
	if err != nil {
		return err
	}
	taskID, err := repository.TaskRepository.AddTask(data)
	if err != nil {
		return err
	}
	tj.TaskConfig.ID = taskID
	t.taskJobMap.Store(taskID, tj)
	return nil
}

func (t *taskLogic) EditTask(data model.Task) error {
	tj, err := t.getTaskJob(data.ID)
	if err != nil {
		return err
	}

	if tj.Running {
		return errors.New("can't edit running task")
	}

	if err := tj.EditStatus(data.Enable); err != nil {
		return err
	}

	tj.TaskConfig = &data
	return repository.TaskRepository.EditTask(&data)
}

func (t *taskLogic) CreateApiKey(id int) error {
	data, err := repository.TaskRepository.GetTaskByID(id)
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
	go t.RunTaskByID(data.ID)
	return nil
}

func (t *taskLogic) RunTaskByTriggerEvent(processName string, event eum.ProcessState) {
	taskList := repository.TaskRepository.GetTriggerTask(processName, event)
	if len(taskList) == 0 {
		return
	}
	log.Logger.Infow("get trigger task", "count", len(taskList), "prcess", processName, "trigger event", event)
	for _, v := range taskList {
		log.Logger.Infow("execute trigger task", "taskID", v.ID)
		t.RunTaskByID(v.ID)
	}
}

func (t *taskLogic) RunTaskByID(id int) error {
	task, err := t.getTaskJob(id)
	if err != nil {
		return err
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
