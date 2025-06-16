package repository

import (
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
)

type wsShare struct{}

var WsShare = new(wsShare)

func (p *wsShare) GetWsShareDataByToken(token string) (data *model.WsShare, err error) {
	data, err = query.WsShare.Where(query.WsShare.Token.Eq(token)).First()
	return
}

func (p *wsShare) AddShareData(data model.WsShare) error {
	return db.Save(&data).Error
}
