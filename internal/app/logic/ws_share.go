package logic

import (
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
)

type wsShareLogic struct{}

var WsShareLogic = &wsShareLogic{}

func (w *wsShareLogic) GetWsShareDataByToken(token string) (*model.WsShare, error) {
	return repository.WsShare.GetWsShareDataByToken(token)
}

func (w *wsShareLogic) GetWsShareList() []*model.WsShare {
	return repository.WsShare.GetWsShareList()
}

func (w *wsShareLogic) DeleteByID(id int) error {
	return repository.WsShare.Delete(id)
}

func (w *wsShareLogic) Edit(data *model.WsShare) error {
	return repository.WsShare.Edit(data)
}

func (w *wsShareLogic) AddShareData(data model.WsShare) error {
	return repository.WsShare.AddShareData(data)
}
