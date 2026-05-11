package logic

import (
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/process"
	"github.com/lzh-1625/go_process_manager/log"
)

type Event struct {
	p     *process.ProcessBase
	state eum.ProcessState
}

// EventBus is a channel that publishes and subscribes to events
type EventBus struct {
	events chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		events: make(chan Event, 32),
	}
}

func (e *EventBus) Publish(event Event) {
	select {
	case e.events <- event:
	default:
		log.Logger.Warnw("event bus is full", "event", event)
	}
}

func (e *EventBus) Subscribe() <-chan Event {
	return e.events
}
