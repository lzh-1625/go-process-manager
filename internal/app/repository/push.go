package repository

import (
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
)

type pushRepository struct{}

var PushRepository = new(pushRepository)

func (p *pushRepository) GetPushList() (result []*model.Push) {
	result, _ = query.Push.Find()
	return
}

func (p *pushRepository) GetPushConfigById(id int) (result *model.Push) {
	result, _ = query.Push.Where(query.Push.Id.Eq(int64(id))).First()
	return
}

func (p *pushRepository) UpdatePushConfig(data model.Push) error {
	return query.Push.Save(&data)
}

func (p *pushRepository) AddPushConfig(data model.Push) error {
	return query.Push.Create(&data)
}

func (p *pushRepository) DeletePushConfig(id int) error {
	_, err := query.Push.Where(query.Push.Id.Eq(int64(id))).Delete()
	return err
}

func (p *pushRepository) GetPushConfigByIds(ids []int64) (result []*model.Push) {
	result, _ = query.Push.Where(query.Push.Id.In(ids...)).Find()
	return
}
