package model

import (
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/eum"
)

type Task struct {
	Id              int               `gorm:"column:id;NOT NULL;primaryKey;autoIncrement;" json:"id" `
	Name            string            `gorm:"column:name" json:"name" `
	ProcessId       int               `gorm:"column:process_id;NOT NULL" json:"processId" `
	Condition       eum.Condition     `gorm:"column:condition;NOT NULL" json:"condition" `
	NextId          *int              `gorm:"column:next_id;" json:"nextId" `
	Operation       eum.TaskOperation `gorm:"column:operation;NOT NULL" json:"operation" `
	TriggerEvent    *eum.ProcessState `gorm:"column:trigger_event;" json:"triggerEvent" `
	TriggerTarget   *int              `gorm:"column:trigger_target;" json:"triggerTarget" `
	OperationTarget int               `gorm:"column:operation_target;NOT NULL" json:"operationTarget" `
	CronExpression  string            `gorm:"column:cron;" json:"cron" `
	Enable          bool              `gorm:"column:enable;" json:"enable" `
	ApiEnable       bool              `gorm:"column:api_enable;" json:"apiEnable" `
	Key             *string           `gorm:"column:key;" json:"key" `
}

func (*Task) TableName() string {
	return "task"
}

type TaskVo struct {
	Task
	ProcessName string    `gorm:"column:process_name;" json:"processName"`
	TargetName  string    `gorm:"column:target_name;" json:"targetName"`
	TriggerName string    `gorm:"column:trigger_name;" json:"triggerName"`
	StartTime   time.Time `json:"startTime"`
	Running     bool      `json:"running"`
}
