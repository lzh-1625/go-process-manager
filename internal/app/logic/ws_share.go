package logic

import (
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
)

type wsShareLogic struct{}

var WsSahreLogic = &wsShareLogic{}

func (w *wsShareLogic) GetWsShareList() []*model.WsShare {
	return repository.WsShare.GetWsShareList()
}

func (w *wsShareLogic) DeleteById(id int) error {
	return repository.WsShare.Delete(id)
}
