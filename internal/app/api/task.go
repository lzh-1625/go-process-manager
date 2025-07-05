package api

import (
	"errors"

	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type taskApi struct{}

var TaskApi = new(taskApi)

func (t *taskApi) CreateTask(ctx *gin.Context, req model.Task) (err error) {
	return logic.TaskLogic.CreateTask(req)
}

func (t *taskApi) GetTaskById(ctx *gin.Context, req model.TaskIdReq) (err error) {
	result, err := repository.TaskRepository.GetTaskById(req.Id)
	if err != nil {
		return err
	}
	rOk(ctx, "Operation successful!", result)
	return
}

func (t *taskApi) GetTaskList(ctx *gin.Context, _ any) (err error) {
	result := logic.TaskLogic.GetAllTaskJob()
	rOk(ctx, "Operation successful!", result)
	return
}

func (t *taskApi) DeleteTaskById(ctx *gin.Context, req model.TaskIdReq) (err error) {
	return logic.TaskLogic.DeleteTask(req.Id)

}

func (t *taskApi) StartTask(ctx *gin.Context, req model.TaskIdReq) (err error) {
	go logic.TaskLogic.RunTaskById(req.Id)
	return
}

func (t *taskApi) StopTask(ctx *gin.Context, req model.TaskIdReq) (err error) {
	if logic.TaskLogic.StopTaskJob(req.Id) != nil {
		return errors.New("operation failed")
	}
	return
}

func (t *taskApi) EditTask(ctx *gin.Context, req model.Task) (err error) {
	return logic.TaskLogic.EditTask(req)
}

func (t *taskApi) EditTaskEnable(ctx *gin.Context, req model.Task) (err error) {
	return logic.TaskLogic.EditTaskEnable(req.Id, req.Enable)
}

func (t *taskApi) RunTaskByKey(ctx *gin.Context, _ any) (err error) {
	return logic.TaskLogic.RunTaskByKey(ctx.Param("key"))

}

func (t *taskApi) CreateTaskApiKey(ctx *gin.Context, req model.TaskIdReq) (err error) {
	return logic.TaskLogic.CreateApiKey(req.Id)
}
