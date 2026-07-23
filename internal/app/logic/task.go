package logic

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"
)

type TaskLogic struct {
	taskRepository  *repository.TaskRepository
	taskJobMap      sync.Map
	eventLogic      *EventLogic
	processCtlLogic *ProcessCtlLogic
}

func NewTaskLogic(
	taskRepository *repository.TaskRepository,
	eventLogic *EventLogic,
	processCtlLogic *ProcessCtlLogic,
) *TaskLogic {
	t := &TaskLogic{
		taskRepository:  taskRepository,
		taskJobMap:      sync.Map{},
		eventLogic:      eventLogic,
		processCtlLogic: processCtlLogic,
	}
	return t
}

func (t *TaskLogic) getTaskJob(id int) (*TaskJob, error) {
	c, ok := t.taskJobMap.Load(id)
	if !ok {
		return nil, errors.New("don't exist this task id")
	}
	return c.(*TaskJob), nil
}

// InitTaskJob initializes all task jobs.
func (t *TaskLogic) InitTaskJob() {
	for _, v := range t.taskRepository.GetAllTask() {
		tj, err := NewTaskJob(v, t.eventLogic, t.processCtlLogic, t)
		if err != nil {
			log.Logger.Warnw("task initialization failed", "err", err)
			continue
		}
		t.taskJobMap.Store(v.ID, tj)
	}
}

// StopTaskJob stops a task job.
func (t *TaskLogic) StopTaskJob(id int) error {
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

// GetAllTaskJob returns all current task jobs.
func (t *TaskLogic) GetAllTaskJob() []model.TaskVo {
	result := t.taskRepository.GetAllTaskWithProcessName()
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

// GetTaskByID returns a task by ID.
func (t *TaskLogic) GetTaskByID(id int) (*model.Task, error) {
	return t.taskRepository.GetTaskByID(id)
}

// DeleteTask deletes a task.
func (t *TaskLogic) DeleteTask(id int) (err error) {

	if tj, err := t.getTaskJob(id); err == nil {
		if tj.Running {
			return errors.New("task is running, can't delete")
		}
	}
	t.StopTaskJob(id)
	t.taskJobMap.Delete(id)
	err = t.taskRepository.DeleteTask(id)
	if err != nil {
		return
	}
	return
}

// CreateTask creates a task.
func (t *TaskLogic) CreateTask(data model.Task) error {
	tj, err := NewTaskJob(&data, t.eventLogic, t.processCtlLogic, t)
	if err != nil {
		return err
	}
	taskID, err := t.taskRepository.AddTask(data)
	if err != nil {
		return err
	}
	tj.TaskConfig.ID = taskID
	t.taskJobMap.Store(taskID, tj)
	return nil
}

// EditTask updates a task.
func (t *TaskLogic) EditTask(data *model.Task) error {
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

	if data.ApiEnable && data.Key == nil {
		data.Key = new(utils.RandString(10))
	}

	tj.TaskConfig = data
	return t.taskRepository.EditTask(data)
}

// CreateApiKey creates an API key that can trigger a task.
func (t *TaskLogic) CreateApiKey(id int) error {
	data, err := t.taskRepository.GetTaskByID(id)
	if err != nil {
		return err
	}
	data.Key = new(utils.RandString(10))
	t.EditTask(data)
	return nil
}

// RunTaskByKey triggers a task using an API key.
func (t *TaskLogic) RunTaskByKey(key string) error {
	data, err := t.taskRepository.GetTaskByKey(key)
	if err != nil {
		return errors.New("don't exist key")
	}
	go t.RunTaskByID(data.ID)
	return nil
}

// RunTaskByTriggerEvent runs tasks triggered by a process state change.
func (t *TaskLogic) RunTaskByTriggerEvent(processName string, event types.ProcessState) {
	taskList := t.taskRepository.GetTriggerTask(processName, event)
	if len(taskList) == 0 {
		return
	}
	log.Logger.Infow("get trigger task", "count", len(taskList), "prcess", processName, "trigger event", event)
	for _, v := range taskList {
		log.Logger.Infow("execute trigger task", "taskID", v.ID)
		t.RunTaskByID(v.ID)
	}
}

// RunTaskByID runs the task with the specified ID.
func (t *TaskLogic) RunTaskByID(id int) error {
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
