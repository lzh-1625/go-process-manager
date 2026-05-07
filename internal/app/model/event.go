package model

import (
	"time"

	"github.com/lzh-1625/go_process_manager/internal/app/eum"
)

type Event struct {
	ID          uint64        `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string        `gorm:"column:name" json:"name"`
	Type        eum.EventType `gorm:"column:type" json:"type"`
	Additional  string        `gorm:"column:additional" json:"additional"`
	CreatedTime time.Time     `gorm:"column:created_time" json:"createdTime"`
}

func (*Event) TableName() string {
	return "event"
}

type EventListReq struct {
	Page      int           `query:"page"`
	Size      int           `query:"size"`
	StartTime int64         `query:"startTime"`
	EndTime   int64         `query:"endTime"`
	Type      eum.EventType `query:"type"`
	Name      string        `query:"name"`
}

type EventListResp struct {
	Total int64    `json:"total"`
	Data  []*Event `json:"data"`
}
