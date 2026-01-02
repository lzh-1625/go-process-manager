package api

import (
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

func (t *taskApi) GetTaskById(ctx *gin.Context, req struct {
	ID int `form:"id" binding:"required"`
}) any {
	result, err := repository.TaskRepository.GetTaskById(req.ID)
	if err != nil {
		return err
	}
	return result
}

func (t *taskApi) GetTaskList(ctx *gin.Context, _ any) any {
	return logic.TaskLogic.GetAllTaskJob()
}

func (t *taskApi) DeleteTaskById(ctx *gin.Context, req struct {
	ID int `form:"id" binding:"required"`
}) (err error) {
	return logic.TaskLogic.DeleteTask(req.ID)

}

func (t *taskApi) StartTask(ctx *gin.Context, req struct {
	ID int `form:"id" binding:"required"`
}) (err error) {
	go logic.TaskLogic.RunTaskById(req.ID)
	return
}

func (t *taskApi) StopTask(ctx *gin.Context, req struct {
	ID int `form:"id" binding:"required"`
}) (err error) {
	return logic.TaskLogic.StopTaskJob(req.ID)
}

func (t *taskApi) EditTask(ctx *gin.Context, req model.Task) (err error) {
	return logic.TaskLogic.EditTask(req)
}

func (t *taskApi) EditTaskEnable(ctx *gin.Context, req model.Task) (err error) {
	return logic.TaskLogic.EditTaskEnable(req.ID, req.Enable)
}

func (t *taskApi) RunTaskByKey(ctx *gin.Context, _ any) (err error) {
	return logic.TaskLogic.RunTaskByKey(ctx.Param("key"))
}

func (t *taskApi) CreateTaskApiKey(ctx *gin.Context, req struct {
	ID int `form:"id" binding:"required"`
}) (err error) {
	return logic.TaskLogic.CreateApiKey(req.ID)
}
