package logic

import (
	"runtime"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type MetricLogic struct {
	processCtlLogic *ProcessCtlLogic
	logHandler      *LogHandler
	logLogic        search.LogLogic
}

func NewMetricLogic(processCtlLogic *ProcessCtlLogic, logHandler *LogHandler, logLogic search.LogLogic) *MetricLogic {
	return &MetricLogic{
		processCtlLogic: processCtlLogic,
		logHandler:      logHandler,
		logLogic:        logLogic,
	}
}

func (m *MetricLogic) GetPerformceUsage() (*model.PerformceUsage, error) {
	pl := m.processCtlLogic.GetProcessList()
	items := make([]model.PerformceUsageItem, 0, len(pl))

	for _, v := range pl {
		if v.State.State != eum.ProcessStateRunning {
			continue
		}
		items = append(items, model.PerformceUsageItem{
			Name: v.Name,
			CPU:  v.Usage.Cpu[len(v.Usage.Cpu)-1] / float64(runtime.NumCPU()),
			Mem:  v.Usage.Mem[len(v.Usage.Mem)-1],
		})
	}

	percentages, err := cpu.Percent(time.Millisecond*200, false)
	if err != nil {
		return nil, err
	}
	cpuUsed := percentages[0]
	cpuIdle := 100 - cpuUsed
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	memFree := vmStat.Available >> 10
	return &model.PerformceUsage{
		CPUFree: cpuIdle,
		MemFree: float64(memFree),
		Items:   items,
	}, nil

}

func (m *MetricLogic) GetLogMetric(dateType int) (result model.LogStatsticMetric) {
	t := time.Now()
	switch dateType {
	case 1:
		for range 7 {
			start := datetime.BeginOfDay(t)
			end := datetime.EndOfDay(t)

			resp := m.logLogic.Search(model.GetLogReq{
				TimeRange: struct {
					StartTime int64 `json:"startTime"`
					EndTime   int64 `json:"endTime"`
				}{
					StartTime: start.UnixMilli(),
					EndTime:   end.UnixMilli(),
				},
			})
			result.Items = append(result.Items, model.LogStatsticMetricItem{
				Date:  t.Format(time.DateOnly),
				Count: resp.Total,
			})
			t = datetime.AddDay(t, -1)
		}
	case 2:
		for range 6 {
			start := datetime.BeginOfWeek(t, time.Monday)
			end := datetime.EndOfWeek(t, time.Monday)

			resp := m.logLogic.Search(model.GetLogReq{
				TimeRange: struct {
					StartTime int64 `json:"startTime"`
					EndTime   int64 `json:"endTime"`
				}{
					StartTime: start.UnixMilli(),
					EndTime:   end.UnixMilli(),
				},
			})
			result.Items = append(result.Items, model.LogStatsticMetricItem{
				Date:  t.Format(time.DateOnly),
				Count: resp.Total,
			})
			t = datetime.AddWeek(t, -1)
		}
	case 3:
		for range 6 {
			start := datetime.BeginOfMonth(t)
			end := datetime.EndOfMonth(t)

			resp := m.logLogic.Search(model.GetLogReq{
				TimeRange: struct {
					StartTime int64 `json:"startTime"`
					EndTime   int64 `json:"endTime"`
				}{
					StartTime: start.UnixMilli(),
					EndTime:   end.UnixMilli(),
				},
			})
			result.Items = append(result.Items, model.LogStatsticMetricItem{
				Date:  t.Format("2006-01"),
				Count: resp.Total,
			})
			t = datetime.AddMonth(t, -1)
		}
	}
	result.Executing = m.logHandler.GetRunning()
	return
}
