package api

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"
)

type ConfigApi struct {
	configLogic *logic.ConfigLogic
	ILogLogic   search.ILogLogic
}

func NewConfigApi(configLogic *logic.ConfigLogic, ILogLogic search.ILogLogic) *ConfigApi {
	return &ConfigApi{
		configLogic: configLogic,
		ILogLogic:   ILogLogic,
	}
}

func (c *ConfigApi) GetSystemConfiguration(ctx *echo.Context) error {
	return ctx.JSON(http.StatusOK, model.Response[[]model.SystemConfigurationVo]{
		Data:    c.configLogic.GetSystemConfiguration(),
		Message: "success",
	})
}

func (c *ConfigApi) SetSystemConfiguration(ctx *echo.Context) error {
	req := map[string]string{}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return c.configLogic.SetSystemConfiguration(req)
}

func (c *ConfigApi) LogConfigReload(ctx *echo.Context) error {
	return c.ILogLogic.Reload()
}
