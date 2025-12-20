package repository

import (
	"context"
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
)

type eventRepository struct{}

var EventRepository = new(eventRepository)

func (e *eventRepository) Create(event model.Event) error {
	return query.Event.Create(&event)
}

func (e *eventRepository) GetList(req model.EventListReq) ([]*model.Event, int64, error) {
	tx := query.Event.WithContext(context.TODO())

	if req.StartTime != 0 {
		tx = tx.Where(query.Event.CreatedTime.Gte(time.Unix(req.StartTime, 0)))
	}
	if req.EndTime != 0 {
		tx = tx.Where(query.Event.CreatedTime.Lte(time.Unix(req.EndTime, 0)))
	}
	if req.Type != "" {
		tx = tx.Where(query.Event.Type.Eq(string(req.Type)))
	}
	if req.Name != "" {
		tx = tx.Where(query.Event.Name.Like("%" + req.Name + "%"))
	}
	return tx.Order(query.Event.CreatedTime.Desc()).FindByPage((req.Page-1)*req.Size, req.Size)
}

func (e *eventRepository) Clean(t time.Duration) error {
	_, err := query.Event.Where(query.Event.CreatedTime.Lt(time.Now().Add(-t))).Delete()
	return err
}
