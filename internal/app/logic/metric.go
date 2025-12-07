package logic

import (
	"time"

	"github.com/duke-git/lancet/v2/datetime"
	"github.com/lzh-1625/go_process_manager/internal/app/eum"
	"github.com/lzh-1625/go_process_manager/internal/app/model"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type metricLogic struct{}

var MetricLogic = new(metricLogic)

func (m *metricLogic) GetPerformceUsage() (*model.PerformceUsage, error) {
	pl := ProcessCtlLogic.GetProcessList()
	items := make([]model.PerformceUsageItem, 0, len(pl))

	for _, v := range pl {
		if v.State.State != eum.ProcessStateRunning {
			continue
		}
		items = append(items, model.PerformceUsageItem{
			Name: v.Name,
			CPU:  v.Usage.Cpu[len(v.Usage.Cpu)-1],
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
	memFree := vmStat.Free >> 10
	return &model.PerformceUsage{
		CPUFree: cpuIdle,
		MemFree: float64(memFree),
		Items:   items,
	}, nil

}

func (m *metricLogic) GetLogMetric(dateType int) (result model.LogStatsticMetric) {
	t := time.Now()
	switch dateType {
	case 1:
		for range 7 {
			start := datetime.BeginOfDay(t)
			end := datetime.EndOfDay(t)

			resp := LogLogicImpl.Search(model.GetLogReq{
				TimeRange: struct {
					StartTime int64 `json:"startTime"`
					EndTime   int64 `json:"endTime"`
				}{
					StartTime: start.Unix(),
					EndTime:   end.Unix(),
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

			resp := LogLogicImpl.Search(model.GetLogReq{
				TimeRange: struct {
					StartTime int64 `json:"startTime"`
					EndTime   int64 `json:"endTime"`
				}{
					StartTime: start.Unix(),
					EndTime:   end.Unix(),
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

			resp := LogLogicImpl.Search(model.GetLogReq{
				TimeRange: struct {
					StartTime int64 `json:"startTime"`
					EndTime   int64 `json:"endTime"`
				}{
					StartTime: start.Unix(),
					EndTime:   end.Unix(),
				},
			})
			result.Items = append(result.Items, model.LogStatsticMetricItem{
				Date:  t.Format("2006-01"),
				Count: resp.Total,
			})
			t = datetime.AddMonth(t, -1)
		}
	}
	result.Executing = Loghandler.GetRunning()
	return
}
