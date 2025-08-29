package logic

import (
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/log"
)

type eventLogic struct{}

var EventLogic = new(eventLogic)

func (e *eventLogic) Create(name string, eventType eum.EventType, additionalKv ...any) error {
	if len(additionalKv)%2 != 0 {
		log.Logger.Errorw("参数长度错误", "args", additionalKv)
	}
	data := model.Event{
		Name:        name,
		CreatedTime: time.Now(),
		Type:        eventType,
	}
	m := map[any]any{}
	for i := range len(additionalKv) / 2 {
		m[additionalKv[2*i]] = additionalKv[2*i+1]
	}
	return repository.EventRepository.Create(data)
}
