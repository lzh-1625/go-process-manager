package api

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/robfig/cron/v3"
)

type TaskApi struct {
	taskLogic *logic.TaskLogic
}

func NewTaskApi(taskLogic *logic.TaskLogic) *TaskApi {
	return &TaskApi{
		taskLogic: taskLogic,
	}
}

func (t *TaskApi) CreateTask(ctx *echo.Context) error {
	var req model.Task
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return t.taskLogic.CreateTask(req)
}

func (t *TaskApi) GetTaskByID(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	result, err := t.taskLogic.GetTaskByID(req.ID)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, model.Response[*model.Task]{
		Data:    result,
		Message: "success",
		Code:    0,
	})
}

func (t *TaskApi) GetTaskList(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, model.Response[[]model.TaskVo]{
		Data:    t.taskLogic.GetAllTaskJob(),
		Message: "success",
		Code:    0,
	})
}

func (t *TaskApi) DeleteTaskByID(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return t.taskLogic.DeleteTask(req.ID)
}

func (t *TaskApi) StartTask(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	go t.taskLogic.RunTaskByID(req.ID)
	return nil
}

func (t *TaskApi) StopTask(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return t.taskLogic.StopTaskJob(req.ID)
}

func (t *TaskApi) EditTask(ctx *echo.Context) error {
	var req model.Task
	if err := ctx.Bind(&req); err != nil {
		return err
	}

	if _, err := cron.ParseStandard(req.CronExpression); err != nil && req.CronExpression != "" { // cron expression validation
		return err
	}
	return t.taskLogic.EditTask(&req)
}

func (t *TaskApi) RunTaskByKey(ctx *echo.Context) error {
	return t.taskLogic.RunTaskByKey(ctx.Param("key"))
}

func (t *TaskApi) CreateTaskApiKey(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return t.taskLogic.CreateApiKey(req.ID)
}
