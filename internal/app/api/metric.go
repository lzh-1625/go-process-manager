package api

import (
	"github.com/labstack/echo"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
)

type metricApi struct{}

var MetricApi = new(metricApi)

func (m *metricApi) GetPerformceUsage(ctx echo.Context) error {
	result, err := logic.MetricLogic.GetPerformceUsage()
	if err != nil {
		return err
	}
	return ctx.JSON(200, model.Response[*model.PerformceUsage]{
		Data:    result,
		Message: "success",
		Code:    0,
	})
}

func (m *metricApi) GetLogicStatsticMetric(ctx echo.Context) error {
	var req struct {
		DateType int `query:"dateType"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return ctx.JSON(200, model.Response[model.LogStatsticMetric]{
		Data:    logic.MetricLogic.GetLogMetric(req.DateType),
		Message: "success",
		Code:    0,
	})
}
