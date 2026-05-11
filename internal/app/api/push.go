package api

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
)

type PushApi struct {
	pushLogic *logic.PushLogic
}

func NewPushApi(pushLogic *logic.PushLogic) *PushApi {
	return &PushApi{
		pushLogic: pushLogic,
	}
}

func (p *PushApi) GetPushList(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, model.Response[[]*model.Push]{
		Data:    p.pushLogic.GetPushList(),
		Message: "success",
		Code:    0,
	})
}

func (p *PushApi) GetPushByID(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, model.Response[*model.Push]{
		Data:    p.pushLogic.GetPushConfigByID(req.ID),
		Message: "success",
		Code:    0,
	})
}

func (p *PushApi) AddPushConfig(ctx *echo.Context) error {
	var req model.Push
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return p.pushLogic.AddPushConfig(req)
}

func (p *PushApi) UpdatePushConfig(ctx *echo.Context) error {
	var req model.Push
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return p.pushLogic.UpdatePushConfig(req)
}

func (p *PushApi) DeletePushConfig(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return p.pushLogic.DeletePushConfig(req.ID)
}
