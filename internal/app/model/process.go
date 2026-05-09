package model

type Process struct {
	UUID              int      `gorm:"primaryKey;autoIncrement;column:uuid" json:"uuid"`
	Name              string   `gorm:"column:name;uniqueIndex;type:text" json:"name" binding:"required"`
	Cmd               string   `gorm:"column:args" json:"cmd"`
	Cwd               string   `gorm:"column:cwd" json:"cwd"`
	AutoRestart       bool     `gorm:"column:auto_restart" json:"autoRestart"`
	CompulsoryRestart bool     `gorm:"column:compulsory_restart" json:"compulsoryRestart"`
	PushIDs           string   `gorm:"column:push_ids;type:json" json:"pushIds"`
	LogReport         bool     `gorm:"column:log_report" json:"logReport"`
	CgroupEnable      bool     `gorm:"column:cgroup_enable" json:"cgroupEnable"`
	MemoryLimit       *float32 `gorm:"column:memory_limit" json:"memoryLimit"`
	CpuLimit          *float32 `gorm:"column:cpu_limit" json:"cpuLimit"`
	Env               string   `gorm:"column:env" json:"env"`
}

func (*Process) TableName() string {
	return "process"
}

type ProcessShare struct {
	Minutes int  `json:"minutes"`
	Pid     int  `json:"pid"`
	Write   bool `json:"write"`
}
