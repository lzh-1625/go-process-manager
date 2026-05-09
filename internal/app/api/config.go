package api

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
)

type configApi struct{}

var ConfigApi = new(configApi)

func (c *configApi) GetSystemConfiguration(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, model.Response[[]model.SystemConfigurationVo]{
		Data:    logic.ConfigLogic.GetSystemConfiguration(),
		Message: "success",
	})
}

func (c *configApi) SetSystemConfiguration(ctx *echo.Context) error {
	req := map[string]string{}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return logic.ConfigLogic.SetSystemConfiguration(req)
}

func (c *configApi) LogConfigReload(ctx *echo.Context) error {
	return logic.LogLogicImpl.Init()
}
