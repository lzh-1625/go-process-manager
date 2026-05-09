package api

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
)

type taskApi struct{}

var TaskApi = new(taskApi)

func (t *taskApi) CreateTask(ctx *echo.Context) error {
	var req model.Task
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return logic.TaskLogic.CreateTask(req)
}

func (t *taskApi) GetTaskById(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id" binding:"required"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	result, err := repository.TaskRepository.GetTaskById(req.ID)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, model.Response[*model.Task]{
		Data:    result,
		Message: "success",
		Code:    0,
	})
}

func (t *taskApi) GetTaskList(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, model.Response[[]model.TaskVo]{
		Data:    logic.TaskLogic.GetAllTaskJob(),
		Message: "success",
		Code:    0,
	})
}

func (t *taskApi) DeleteTaskById(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id" binding:"required"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return logic.TaskLogic.DeleteTask(req.ID)
}

func (t *taskApi) StartTask(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id" binding:"required"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	go logic.TaskLogic.RunTaskById(req.ID)
	return nil
}

func (t *taskApi) StopTask(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id" binding:"required"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return logic.TaskLogic.StopTaskJob(req.ID)
}

func (t *taskApi) EditTask(ctx *echo.Context) error {
	var req model.Task
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return logic.TaskLogic.EditTask(req)
}

func (t *taskApi) EditTaskEnable(ctx *echo.Context) error {
	var req model.Task
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return logic.TaskLogic.EditTaskEnable(req.ID, req.Enable)
}

func (t *taskApi) RunTaskByKey(ctx *echo.Context) error {
	return logic.TaskLogic.RunTaskByKey(ctx.Param("key"))
}

func (t *taskApi) CreateTaskApiKey(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id" binding:"required"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return logic.TaskLogic.CreateApiKey(req.ID)
}
