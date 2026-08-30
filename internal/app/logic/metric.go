package logic

import (
	"fmt"
	"maps"
	"runtime"
	"slices"
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/lzh-1625/go_process_manager/internal/app/repository/search"
	"github.com/lzh-1625/go_process_manager/internal/app/types"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type MetricLogic struct {
	processCtlLogic *ProcessCtlLogic
	logHandler      *LogHandler
	ILogLogic       search.ILogLogic
}

func NewMetricLogic(processCtlLogic *ProcessCtlLogic, logHandler *LogHandler, ILogLogic search.ILogLogic) *MetricLogic {
	return &MetricLogic{
		processCtlLogic: processCtlLogic,
		logHandler:      logHandler,
		ILogLogic:       ILogLogic,
	}
}

// GetPerformceUsage returns performance metrics for all current processes.
func (m *MetricLogic) GetPerformceUsage() (*model.PerformceUsage, error) {
	pl := m.processCtlLogic.GetProcessList()
	items := make([]model.PerformceUsageItem, 0, len(pl))

	for _, v := range pl {
		if v.State.State != types.ProcessStateRunning {
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

// GetLogMetric returns recent log statistics.
// dateType selects the time range: 0 for 12 hours, 1 for 7 days, 2 for 6 weeks, and 3 for 6 months.
func (m *MetricLogic) GetLogMetric(dateType int) *model.LogStatsticMetric {
	pl := m.processCtlLogic.GetProcessList()
	result := &model.LogStatsticMetric{
		Items: make(map[string][]model.LogStatsticMetricItem, len(pl)),
	}
	for _, v := range pl {
		t := time.Now()
		result.Items[v.Name] = []model.LogStatsticMetricItem{}
		switch dateType {
		case 0:
			for range 12 {
				start := datetime.BeginOfHour(t)
				end := datetime.EndOfHour(t)
				req := model.GetLogReq{}
				req.FilterName = []string{v.Name}
				req.TimeRange.StartTime = start.UnixMilli()
				req.TimeRange.EndTime = end.UnixMilli()
				resp, err := m.ILogLogic.Search(req)
				if err != nil {
					return nil
				}
				result.Items[v.Name] = append(result.Items[v.Name], model.LogStatsticMetricItem{
					Date:  fmt.Sprintf("%d:00", t.Hour()),
					Count: resp.Total,
				})
				t = datetime.AddHour(t, -1)
			}
		case 1:
			for range 7 {
				start := datetime.BeginOfDay(t)
				end := datetime.EndOfDay(t)
				req := model.GetLogReq{}
				req.FilterName = []string{v.Name}
				req.TimeRange.StartTime = start.UnixMilli()
				req.TimeRange.EndTime = end.UnixMilli()
				resp, err := m.ILogLogic.Search(req)
				if err != nil {
					return nil
				}
				result.Items[v.Name] = append(result.Items[v.Name], model.LogStatsticMetricItem{
					Date:  t.Format(time.DateOnly),
					Count: resp.Total,
				})
				t = datetime.AddDay(t, -1)
			}
		case 2:
			for range 6 {
				start := datetime.BeginOfWeek(t, time.Monday)
				end := datetime.EndOfWeek(t, time.Monday)
				req := model.GetLogReq{}
				req.FilterName = []string{v.Name}
				req.TimeRange.StartTime = start.UnixMilli()
				req.TimeRange.EndTime = end.UnixMilli()
				resp, err := m.ILogLogic.Search(req)
				if err != nil {
					return nil
				}
				result.Items[v.Name] = append(result.Items[v.Name], model.LogStatsticMetricItem{
					Date:  t.Format(time.DateOnly),
					Count: resp.Total,
				})
				t = datetime.AddWeek(t, -1)
			}
		case 3:
			for range 6 {
				start := datetime.BeginOfMonth(t)
				end := datetime.EndOfMonth(t)
				req := model.GetLogReq{}
				req.FilterName = []string{v.Name}
				req.TimeRange.StartTime = start.UnixMilli()
				req.TimeRange.EndTime = end.UnixMilli()
				resp, err := m.ILogLogic.Search(req)

				if err != nil {
					return nil
				}
				result.Items[v.Name] = append(result.Items[v.Name], model.LogStatsticMetricItem{
					Date:  t.Format("2006-01"),
					Count: resp.Total,
				})
				t = datetime.AddMonth(t, -1)
			}
		}
	}
	maps.DeleteFunc(result.Items, func(s string, v []model.LogStatsticMetricItem) bool {
		return !slices.ContainsFunc(v, func(item model.LogStatsticMetricItem) bool {
			return item.Count != 0
		})
	})
	result.Executing = m.logHandler.GetRunning()
	return result
}
