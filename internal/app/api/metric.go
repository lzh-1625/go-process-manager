package api

import (
	"github.com/gin-gonic/gin"
	"github.com/lzh-1625/go_process_manager/internal/app/logic"
)

type metricApi struct{}

var MetricApi = new(metricApi)

func (m *metricApi) GetPerformceUsage(ctx *gin.Context, _ any) any {
	result, err := logic.MetricLogic.GetPerformceUsage()
	if err != nil {
		return err
	}
	return result
}

func (m *metricApi) GetLogicStatsticMetric(ctx *gin.Context, req struct {
	DateType int `form:"dateType"`
}) any {
	return logic.MetricLogic.GetLogMetric(req.DateType)
}
