package repository

import (
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
)

func NewWsShareRepository(query *query.Query) *WsShareRepository {
	return &WsShareRepository{
		query: query,
	}
}

type WsShareRepository struct {
	query *query.Query
}

func (p *WsShareRepository) GetWsShareDataByToken(token string) (data *model.WsShare, err error) {
	data, err = p.query.WsShare.Where(p.query.WsShare.Token.Eq(token)).First()
	return
}

func (p *WsShareRepository) AddShareData(data model.WsShare) error {
	return p.query.WsShare.Save(&data)
}

func (p *WsShareRepository) GetWsShareList() (data []*model.WsShare) {
	ws := p.query.WsShare
	data, _ = ws.Order(ws.CreatedAt.Desc()).Find()
	return
}

func (p *WsShareRepository) Delete(id int) error {
	ws := p.query.WsShare
	_, err := ws.Where(ws.ID.Eq(id)).Delete()
	return err
}

func (p *WsShareRepository) Edit(data *model.WsShare) error {
	ws := p.query.WsShare
	_, err := ws.Where(ws.ID.Eq(data.ID)).Updates(data)
	return err
}
