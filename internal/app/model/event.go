package model

import (
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/eum"
)

type Event struct {
	Id          uint64        `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string        `gorm:"column:name" json:"name"`
	Type        eum.EventType `gorm:"column:type" json:"type"`
	Additional  string        `gorm:"column:additional" json:"additional"`
	CreatedTime time.Time     `gorm:"column:created_time" json:"createdTime"`
}

func (*Event) TableName() string {
	return "event"
}
