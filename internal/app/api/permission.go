package api

import (
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"

	"github.com/gin-gonic/gin"
)

var PermissionApi = new(permissionApi)

type permissionApi struct{}

func (p *permissionApi) EditPermssion(ctx *gin.Context, req model.Permission) (err error) {
	return repository.PermissionRepository.EditPermssion(req)
}

func (p *permissionApi) GetPermissionList(ctx *gin.Context, req model.GetPermissionListReq) (err error) {
	result := repository.PermissionRepository.GetPermssionList(req.Account)
	rOk(ctx, "Query successful!", result)
	return
}
