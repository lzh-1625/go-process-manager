package api

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
)

type MetricApi struct {
	metricLogic *logic.MetricLogic
}

func NewMetricApi(metricLogic *logic.MetricLogic) *MetricApi {
	return &MetricApi{
		metricLogic: metricLogic,
	}
}

func (m *MetricApi) GetPerformceUsage(ctx *echo.Context) error {
	result, err := m.metricLogic.GetPerformceUsage()
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, model.Response[*model.PerformceUsage]{
		Data:    result,
		Message: "success",
		Code:    0,
	})
}

func (m *MetricApi) GetLogicStatsticMetric(ctx *echo.Context) error {
	var req struct {
		DateType int `query:"dateType"`
	}
	if err := ctx.Bind(&req); err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, model.Response[model.LogStatsticMetric]{
		Data:    m.metricLogic.GetLogMetric(req.DateType),
		Message: "success",
		Code:    0,
	})
}
