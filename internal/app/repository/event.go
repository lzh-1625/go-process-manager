package repository

import (
	"context"
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
)

func NewEventRepository(query *query.Query) *EventRepository {
	return &EventRepository{
		query: query,
	}
}

type EventRepository struct {
	query *query.Query
}

func (e *EventRepository) Create(event model.Event) error {
	return e.query.Event.Create(&event)
}

func (e *EventRepository) GetList(req model.EventListReq) ([]*model.Event, int64, error) {
	tx := e.query.Event.WithContext(context.TODO())

	if req.StartTime != 0 {
		tx = tx.Where(e.query.Event.CreatedTime.Gte(time.Unix(req.StartTime, 0)))
	}
	if req.EndTime != 0 {
		tx = tx.Where(e.query.Event.CreatedTime.Lte(time.Unix(req.EndTime, 0)))
	}
	if req.Type != "" {
		tx = tx.Where(e.query.Event.Type.Eq(string(req.Type)))
	}
	if req.Name != "" {
		tx = tx.Where(e.query.Event.Name.Like("%" + req.Name + "%"))
	}
	return tx.Order(e.query.Event.CreatedTime.Desc()).FindByPage((req.Page-1)*req.Size, req.Size)
}

func (e *EventRepository) Clean(t time.Duration) error {
	_, err := e.query.Event.Where(e.query.Event.CreatedTime.Lt(time.Now().Add(-t))).Delete()
	return err
}
