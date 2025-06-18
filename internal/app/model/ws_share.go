package model

import (
	"time"

	"gorm.io/gorm"
)

type WsShare struct {
	gorm.Model
	Pid        int       `gorm:"column:pid" json:"pid"`
	Write      bool      `gorm:"column:write" json:"write"`
	ExpireTime time.Time `gorm:"column:expire_time" json:"expireTime"`
	CreateBy   string    `gorm:"column:create_by" json:"createBy"`
	Token      string    `gorm:"column:token" json:"token"`
}
