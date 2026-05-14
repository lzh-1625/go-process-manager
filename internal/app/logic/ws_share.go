package logic

import (
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
)

type WsShareLogic struct {
	wsShareRepository *repository.WsShareRepository
}

func NewWsShareLogic(wsShareRepository *repository.WsShareRepository) *WsShareLogic {
	return &WsShareLogic{
		wsShareRepository: wsShareRepository,
	}
}

func (w *WsShareLogic) GetWsShareDataByToken(token string) (*model.WsShare, error) {
	return w.wsShareRepository.GetWsShareDataByToken(token)
}

func (w *WsShareLogic) GetWsShareList() []*model.WsShare {
	return w.wsShareRepository.GetWsShareList()
}

func (w *WsShareLogic) DeleteByID(id int) error {
	return w.wsShareRepository.Delete(id)
}

func (w *WsShareLogic) Edit(data *model.WsShare) error {
	return w.wsShareRepository.Edit(data)
}

func (w *WsShareLogic) AddShareData(data model.WsShare) error {
	return w.wsShareRepository.AddShareData(data)
}
