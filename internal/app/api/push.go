package api

import (
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type pushApi struct{}

var PushApi = new(pushApi)

func (p *pushApi) GetPushList(ctx *gin.Context, __ any) any {
	return repository.PushRepository.GetPushList()
}

func (p *pushApi) GetPushById(ctx *gin.Context, req struct {
	Id int `form:"id" binding:"required"`
}) any {
	return repository.PushRepository.GetPushConfigById(req.Id)
}

func (p *pushApi) AddPushConfig(ctx *gin.Context, req model.Push) (err error) {
	return repository.PushRepository.AddPushConfig(req)
}

func (p *pushApi) UpdatePushConfig(ctx *gin.Context, req model.Push) (err error) {
	return repository.PushRepository.UpdatePushConfig(req)
}

func (p *pushApi) DeletePushConfig(ctx *gin.Context, req struct {
	Id int `form:"id" binding:"required"`
}) (err error) {
	return repository.PushRepository.DeletePushConfig(req.Id)
}
