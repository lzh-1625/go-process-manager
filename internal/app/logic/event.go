package logic

import (
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository"
	"github.com/lzh-1625/go_process_manager/log"
	"github.com/lzh-1625/go_process_manager/utils"
)

type eventLogic struct{}

var EventLogic = new(eventLogic)

func (e *eventLogic) Create(name string, eventType eum.EventType, additionalKv ...string) {
	if len(additionalKv)%2 != 0 {
		log.Logger.Errorw("参数长度错误", "args", additionalKv)
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
	if err := repository.EventRepository.Create(data); err != nil {
		log.Logger.Errorw("事件创建失败", "err", err)
	}
}

func (e *eventLogic) Get(req model.EventListReq) ([]*model.Event, int64, error) {
	return repository.EventRepository.GetList(req)
}
