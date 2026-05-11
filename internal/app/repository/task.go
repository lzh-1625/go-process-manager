package repository

import (
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
)

func NewTaskRepository(query *query.Query) *TaskRepository {
	return &TaskRepository{
		query: query,
	}
}

type TaskRepository struct {
	query *query.Query
}

func (t *TaskRepository) GetAllTask() (result []*model.Task) {
	result, _ = t.query.Task.Find()
	return
}

func (t *TaskRepository) GetTaskByID(id int) (result *model.Task, err error) {
	result, err = t.query.Task.Where(t.query.Task.ID.Eq(id)).First()
	return
}

func (t *TaskRepository) GetTaskByKey(key string) (result *model.Task, err error) {
	result, err = t.query.Task.Where(t.query.Task.Key.Eq(key), t.query.Task.ApiEnable.Is(true)).First()
	return
}

func (t *TaskRepository) AddTask(data model.Task) (task int, err error) {
	err = t.query.Task.Create(&data)
	task = data.ID
	return
}

func (t *TaskRepository) DeleteTask(id int) (err error) {
	_, err = t.query.Task.Where(t.query.Task.ID.Eq(id)).Delete()
	return
}

func (t *TaskRepository) EditTask(data *model.Task) (err error) {
	err = t.query.Task.Save(data)
	return
}

func (t *TaskRepository) GetAllTaskWithProcessName() (result []model.TaskVo) {
	p := t.query.Process.As("p")
	p2 := t.query.Process.As("p2")
	p3 := t.query.Process.As("p3")
	task := t.query.Task
	task.Select(
		task.ALL,
		p.Name.As("process_name"),
		p2.Name.As("target_name"),
		p3.Name.As("trigger_name"),
	).
		LeftJoin(p, p.UUID.EqCol(task.ProcessID)).
		LeftJoin(p2, p2.UUID.EqCol(task.OperationTarget)).
		LeftJoin(p3, p3.UUID.EqCol(task.TriggerTarget)).
		Scan(&result)
	return
}

func (t *TaskRepository) GetTriggerTask(processName string, event eum.ProcessState) []model.Task {
	result := []model.Task{}
	t.query.Task.Select(t.query.Task.ALL).
		LeftJoin(t.query.Process, t.query.Process.UUID.EqCol(t.query.Task.TriggerTarget)).
		Where(t.query.Process.Name.Eq(processName)).
		Where(t.query.Task.TriggerEvent.Eq(int32(event))).
		Scan(&result)
	return result
}
