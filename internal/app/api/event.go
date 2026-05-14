package api

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
)

type EventApi struct {
	eventLogic *logic.EventLogic
}

func NewEventApi(eventLogic *logic.EventLogic) *EventApi {
	return &EventApi{
		eventLogic: eventLogic,
	}
}

func (e *EventApi) GetEventList(ctx *echo.Context) error {
	var req model.EventListReq
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	data, total, err := e.eventLogic.Get(req)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, model.Response[model.EventListResp]{
		Data: model.EventListResp{
			Total: total,
			Data:  data,
		},
		Message: "success",
	})
}
