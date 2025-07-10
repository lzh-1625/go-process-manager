package api

import (
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"

	"github.com/gin-gonic/gin"
)

type configApi struct{}

var ConfigApi = new(configApi)

func (c *configApi) GetSystemConfiguration(ctx *gin.Context, _ any) []model.SystemConfigurationVo {
	return logic.ConfigLogic.GetSystemConfiguration()
}

func (c *configApi) SetSystemConfiguration(ctx *gin.Context, _ any) (err error) {
	req := map[string]string{}
	if err = ctx.BindJSON(&req); err != nil {
		return err
	}
	err = logic.ConfigLogic.SetSystemConfiguration(req)
	return
}

func (c *configApi) LogConfigReload(ctx *gin.Context, _ any) (err error) {
	return logic.LogLogicImpl.Init()
}
