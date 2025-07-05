package api

import (
	"errors"

	"github.com/lzh-1625/go_process_manager/internal/app/logic"

	"github.com/gin-gonic/gin"
)

type configApi struct{}

var ConfigApi = new(configApi)

func (c *configApi) GetSystemConfiguration(ctx *gin.Context, _ any) error {
	result := logic.ConfigLogic.GetSystemConfiguration()
	rOk(ctx, "Operation successful!", result)
	return nil
}

func (c *configApi) SetSystemConfiguration(ctx *gin.Context, _ any) (err error) {
	req := map[string]string{}
	if err = ctx.BindJSON(&req); err != nil {
		return
	}
	if err = logic.ConfigLogic.SetSystemConfiguration(req); err != nil {
		return
	}
	return
}

func (c *configApi) EsConfigReload(ctx *gin.Context, _ any) (err error) {
	if !logic.EsLogic.InitEs() {
		return errors.New("es init fail")
	}
	return
}
