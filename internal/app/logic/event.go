package logic

import (
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"
)

type EventLogic struct {
	eventRepository *repository.EventRepository
}

func NewEventLogic(eventRepository *repository.EventRepository) *EventLogic {
	return &EventLogic{
		eventRepository: eventRepository,
	}
}

func (e *EventLogic) Create(name string, eventType eum.EventType, additionalKv ...string) {
	if len(additionalKv)%2 != 0 {
		log.Logger.Errorw("parameters length error", "args", additionalKv)
		return
	}
	data := model.Event{
		Name:        name,
		CreatedTime: time.Now(),
		Type:        eventType,
	}
	m := map[string]string{}
	for i := range len(additionalKv) / 2 {
		m[additionalKv[2*i]] = additionalKv[2*i+1]
	}
	data.Additional = utils.StructToJsonStr(m)
	if err := e.eventRepository.Create(data); err != nil {
		log.Logger.Errorw("event creation failed", "err", err)
	}
}

func (e *EventLogic) Get(req model.EventListReq) ([]*model.Event, int64, error) {
	return e.eventRepository.GetList(req)
}

func (e *EventLogic) Clean(t time.Duration) error {
	return e.eventRepository.Clean(t)
}
