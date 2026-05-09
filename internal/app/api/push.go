package api

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
)

type pushApi struct{}

var PushApi = new(pushApi)

func (p *pushApi) GetPushList(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, model.Response[[]*model.Push]{
		Data:    logic.PushLogic.GetPushList(),
		Message: "success",
		Code:    0,
	})
}

func (p *pushApi) GetPushByID(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id" binding:"required"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, model.Response[*model.Push]{
		Data:    logic.PushLogic.GetPushConfigByID(req.ID),
		Message: "success",
		Code:    0,
	})
}

func (p *pushApi) AddPushConfig(ctx *echo.Context) error {
	var req model.Push
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return logic.PushLogic.AddPushConfig(req)
}

func (p *pushApi) UpdatePushConfig(ctx *echo.Context) error {
	var req model.Push
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return logic.PushLogic.UpdatePushConfig(req)
}

func (p *pushApi) DeletePushConfig(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id" binding:"required"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return logic.PushLogic.DeletePushConfig(req.ID)
}
