package model

import (
	"time"
)

type WsShare struct {
	Id         int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Pid        int       `gorm:"column:pid" json:"pid"`
	Write      bool      `gorm:"column:write" json:"write"`
	ExpireTime time.Time `gorm:"column:expire_time" json:"expireTime"`
	Token      string    `gorm:"column:token" json:"token"`
}
