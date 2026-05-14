package repository

import (
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
)

func NewPushRepository(query *query.Query) *PushRepository {
	return &PushRepository{
		query: query,
	}
}

type PushRepository struct {
	query *query.Query
}

func (p *PushRepository) GetPushList() (result []*model.Push) {
	result, _ = p.query.Push.Find()
	return
}

func (p *PushRepository) GetPushConfigByID(id int) (result *model.Push) {
	result, _ = p.query.Push.Where(p.query.Push.ID.Eq(int64(id))).First()
	return
}

func (p *PushRepository) UpdatePushConfig(data model.Push) error {
	return p.query.Push.Save(&data)
}

func (p *PushRepository) AddPushConfig(data model.Push) error {
	return p.query.Push.Create(&data)
}

func (p *PushRepository) DeletePushConfig(id int) error {
	_, err := p.query.Push.Where(p.query.Push.ID.Eq(int64(id))).Delete()
	return err
}

func (p *PushRepository) GetPushConfigByIDs(ids []int64) (result []*model.Push) {
	result, _ = p.query.Push.Where(p.query.Push.ID.In(ids...)).Find()
	return
}
