package repository

import (
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/query"
)

type eventRepository struct{}

var EventRepository = new(eventRepository)

func (e *eventRepository) Create(event model.Event) error {
	return query.Event.Create(&event)
}
