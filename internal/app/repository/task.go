package repository

import (
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
)

type taskRepository struct{}

var TaskRepository = new(taskRepository)

func (t *taskRepository) GetAllTask() (result []*model.Task) {
	result, _ = query.Task.Find()
	return
}

func (t *taskRepository) GetTaskById(id int) (result *model.Task, err error) {
	result, err = query.Task.Where(query.Task.ID.Eq(id)).First()
	return
}

func (t *taskRepository) GetTaskByKey(key string) (result *model.Task, err error) {
	result, err = query.Task.Where(query.Task.Key.Eq(key), query.Task.ApiEnable.Is(true)).First()
	return
}

func (t *taskRepository) AddTask(data model.Task) (taskId int, err error) {
	err = query.Task.Create(&data)
	taskId = data.ID
	return
}

func (t *taskRepository) DeleteTask(id int) (err error) {
	_, err = query.Task.Where(query.Task.ID.Eq(id)).Delete()
	return
}

func (t *taskRepository) EditTask(data *model.Task) (err error) {
	err = query.Task.Save(data)
	return
}

func (t *taskRepository) EditTaskEnable(id int, enable bool) (err error) {
	_, err = query.Task.Where(query.Task.ID.Eq(id)).Update(query.Task.Enable, enable)
	return
}

func (t *taskRepository) GetAllTaskWithProcessName() (result []model.TaskVo) {
	p := query.Process.As("p")
	p2 := query.Process.As("p2")
	p3 := query.Process.As("p3")
	task := query.Task
	task.Select(
		task.ALL,
		p.Name.As("process_name"),
		p2.Name.As("target_name"),
		p3.Name.As("trigger_name"),
	).
		LeftJoin(p, p.UUID.EqCol(task.ProcessId)).
		LeftJoin(p2, p2.UUID.EqCol(task.OperationTarget)).
		LeftJoin(p3, p3.UUID.EqCol(task.TriggerTarget)).
		Scan(&result)
	return
}

func (t *taskRepository) GetTriggerTask(processName string, event eum.ProcessState) []model.Task {
	result := []model.Task{}
	query.Task.Select(query.Task.ALL).
		LeftJoin(query.Process, query.Process.UUID.EqCol(query.Task.TriggerTarget)).
		Where(query.Process.Name.Eq(processName)).
		Where(query.Task.TriggerEvent.Eq(int32(event))).
		Scan(&result)
	return result
}
