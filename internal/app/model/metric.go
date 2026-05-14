package model

type PerformceUsage struct {
	Items   []PerformceUsageItem `json:"items"`
	CPUFree float64              `json:"cpuFree"`
	MemFree float64              `json:"memFree"`
}

type PerformceUsageItem struct {
	Name string  `json:"name"`
	CPU  float64 `json:"cpu"`
	Mem  float64 `json:"mem"`
}

type LogStatsticMetric struct {
	Executing int                     `json:"executing"`
	Items     []LogStatsticMetricItem `json:"items"`
}

type LogStatsticMetricItem struct {
	Count int64  `json:"count"`
	Date  string `json:"date"`
}
