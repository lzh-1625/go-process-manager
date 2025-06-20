package repository

import (
	"errors"

	"github.com/lzh-1625/go_process_manager/internal/app/constants"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"

	"gorm.io/gorm"
)

type taskRepository struct{}

var TaskRepository = new(taskRepository)

func (t *taskRepository) GetAllTask() (result []model.Task) {
	db.Find(&result)
	return
}

func (t *taskRepository) GetTaskById(id int) (result model.Task, err error) {
	err = db.Model(&model.Task{}).Where(&model.Task{Id: int(id)}).First(&result).Error
	return
}

func (t *taskRepository) GetTaskByKey(key string) (result model.Task, err error) {
	err = db.Model(&model.Task{}).Where(&model.Task{Key: &key, ApiEnable: true}).First(&result).Error
	return
}

func (t *taskRepository) AddTask(data model.Task) (taskId int, err error) {
	err = db.Create(&data).Error
	taskId = data.Id
	return
}

func (t *taskRepository) DeleteTask(id int) (err error) {
	err = db.Delete(&model.Task{Id: id}).Error
	return
}

func (t *taskRepository) EditTask(data model.Task) (err error) {
	err = db.Model(&model.Task{}).Where(&model.Task{Id: data.Id}).First(&model.Task{}).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	err = db.Model(&model.Task{}).Where(&model.Task{Id: data.Id}).Save(data).Error
	return
}

func (t *taskRepository) EditTaskEnable(id int, enable bool) (err error) {
	_, err = query.Task.Where(query.Task.Id.Eq(id)).Update(query.Task.Enable, enable)
	return
}

func (t *taskRepository) GetAllTaskWithProcessName() (result []model.TaskVo) {
	process := query.Process
	task := query.Task
	task.Select(
		task.ALL,
		process.As("p").Name.As("process_name"),
		process.As("p2").Name.As("target_name"),
		process.As("p3").Name.As("trigger_name"),
	).
		LeftJoin(process, process.As("p").Uuid.EqCol(task.ProcessId)).
		LeftJoin(process, process.As("p2").Uuid.EqCol(task.OperationTarget)).
		LeftJoin(process, process.As("p3").Uuid.EqCol(task.TriggerTarget)).
		Scan(&result)
	return
}

func (t *taskRepository) GetTriggerTask(processName string, event constants.ProcessState) []model.Task {
	result := []model.Task{}
	query.Task.Select(query.Task.ALL).
		LeftJoin(query.Process, query.Process.Uuid.EqCol(query.Task.TriggerTarget)).
		Where(query.Process.Name.Eq(processName)).
		Where(query.Task.TriggerEvent.Eq(int32(event))).
		Scan(result)
	return result
}
