package api

import (
	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
)

var PermissionApi = new(permissionApi)

type permissionApi struct{}

func (p *permissionApi) EditPermssion(ctx *echo.Context) error {
	var req model.Permission
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return repository.PermissionRepository.EditPermssion(req)
}

func (p *permissionApi) GetPermissionList(ctx *echo.Context) error {
	var req struct {
		Account string `query:"account" binding:"required"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return ctx.JSON(200, model.Response[[]model.PermissionPo]{
		Data:    repository.PermissionRepository.GetPermssionList(req.Account),
		Message: "success",
		Code:    0,
	})
}
