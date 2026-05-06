package api

import (
	"github.com/labstack/echo"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
)

type eventApi struct{}

var EventApi = new(eventApi)

func (e *eventApi) GetEventList(ctx echo.Context) error {
	var req model.EventListReq
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	data, total, err := logic.EventLogic.Get(req)
	if err != nil {
		return err
	}
	return ctx.JSON(200, model.EventListResp{
		Total: total,
		Data:  data,
	})
}
