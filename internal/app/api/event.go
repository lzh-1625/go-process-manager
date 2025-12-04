package api

import (
	"github.com/gin-gonic/gin"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
)

type eventApi struct{}

var EventApi = new(eventApi)

func (e *eventApi) GetEventList(ctx *gin.Context, req model.EventListReq) any {
	data, total, err := logic.EventLogic.Get(req)
	if err != nil {
		return err
	}
	return model.EventListResp{
		Total: total,
		Data:  data,
	}
}
