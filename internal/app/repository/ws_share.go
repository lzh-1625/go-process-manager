package repository

import "github.com/lzh-1625/go_process_manager/internal/app/model"

type wsShare struct{}

var WsShare = new(wsShare)

func (p *wsShare) GetWsShareDataByToken(token string) (data model.WsShare, err error) {
	err = db.Model(&model.WsShare{}).Where("token = ?", token).First(&data).Error
	return
}

func (p *wsShare) AddShareData(data model.WsShare) error {
	return db.Save(data).Error
}
