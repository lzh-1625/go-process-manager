package api

import (
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type pushApi struct{}

var PushApi = new(pushApi)

func (p *pushApi) GetPushList(ctx *gin.Context, __ any) (err error) {
	rOk(ctx, "Query successful!", repository.PushRepository.GetPushList())
	return
}

func (p *pushApi) GetPushById(ctx *gin.Context, req model.PushIdReq) (err error) {
	rOk(ctx, "Query successful!", repository.PushRepository.GetPushConfigById(req.Id))
	return
}

func (p *pushApi) AddPushConfig(ctx *gin.Context, req model.Push) (err error) {
	err = repository.PushRepository.AddPushConfig(req)
	return
}

func (p *pushApi) UpdatePushConfig(ctx *gin.Context, req model.Push) (err error) {
	err = repository.PushRepository.UpdatePushConfig(req)
	return
}

func (p *pushApi) DeletePushConfig(ctx *gin.Context, req model.PushIdReq) (err error) {
	err = repository.PushRepository.DeletePushConfig(req.Id)
	return
}
