package api

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
)

type PermissionApi struct {
	permissionLogic *logic.PermissionLogic
}

func NewPermissionApi(permissionLogic *logic.PermissionLogic) *PermissionApi {
	return &PermissionApi{
		permissionLogic: permissionLogic,
	}
}

func (p *PermissionApi) EditPermssion(ctx *echo.Context) error {
	var req model.Permission
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return p.permissionLogic.EditPermssion(req)
}

func (p *PermissionApi) GetPermissionList(ctx *echo.Context) error {
	var req struct {
		Account string `query:"account"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, model.Response[[]model.PermissionPo]{
		Data:    p.permissionLogic.GetPermssionList(req.Account),
		Message: "success",
		Code:    0,
	})
}

func (p *PermissionApi) hasOprPermission(c *echo.Context, uuid int, op types.OprPermission) bool {
	if isAdmin(c) {
		return true
	}
	per := p.permissionLogic.GetPermission(
		getUserName(c),
		uuid,
	)
	if per == nil {
		return false
	}

	switch op {
	case types.OperationLog:
		return per.Log
	case types.OperationTerminal:
		return per.Terminal
	case types.OperationStart:
		return per.Start
	case types.OperationStop:
		return per.Stop
	case types.OperationTerminalWrite:
		return per.Write
	default:
		panic("unknown operation")
	}
}
