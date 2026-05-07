package api

import (
	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
)

type pushApi struct{}

var PushApi = new(pushApi)

func (p *pushApi) GetPushList(ctx *echo.Context) error {
	return ctx.JSON(200, model.Response[[]*model.Push]{
		Data:    repository.PushRepository.GetPushList(),
		Message: "success",
		Code:    0,
	})
}

func (p *pushApi) GetPushById(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id" binding:"required"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return ctx.JSON(200, model.Response[*model.Push]{
		Data:    repository.PushRepository.GetPushConfigById(req.ID),
		Message: "success",
		Code:    0,
	})
}

func (p *pushApi) AddPushConfig(ctx *echo.Context) error {
	var req model.Push
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return repository.PushRepository.AddPushConfig(req)
}

func (p *pushApi) UpdatePushConfig(ctx *echo.Context) error {
	var req model.Push
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return repository.PushRepository.UpdatePushConfig(req)
}

func (p *pushApi) DeletePushConfig(ctx *echo.Context) error {
	var req struct {
		ID int `query:"id" binding:"required"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return repository.PushRepository.DeletePushConfig(req.ID)
}
