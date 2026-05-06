package api

import (
	"github.com/labstack/echo"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
)

type configApi struct{}

var ConfigApi = new(configApi)

func (c *configApi) GetSystemConfiguration(ctx echo.Context) error {
	return ctx.JSON(200, logic.ConfigLogic.GetSystemConfiguration())
}

func (c *configApi) SetSystemConfiguration(ctx echo.Context) error {
	req := map[string]string{}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return logic.ConfigLogic.SetSystemConfiguration(req)
}

func (c *configApi) LogConfigReload(ctx echo.Context) error {
	return logic.LogLogicImpl.Init()
}
