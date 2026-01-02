package model

import (
	"time"
)

type WsShare struct {
	ID         int       `gorm:"column:id;primaryKey" json:"id"`
	Pid        int       `gorm:"column:pid" json:"pid"`
	Write      bool      `gorm:"column:write" json:"write"`
	ExpireTime time.Time `gorm:"column:expire_time" json:"expireTime"`
	CreateBy   string    `gorm:"column:create_by" json:"createBy"`
	Token      string    `gorm:"column:token" json:"token"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"createdAt"`
	LastLink   time.Time `gorm:"column:last_link" json:"lastLink"`
}

func (WsShare) TableName() string {
	return "ws_share"
}

type WebsocketHandleReq struct {
	UUID  int    `form:"uuid"`
	Cols  int    `form:"cols"`
	Rows  int    `form:"rows"`
	Token string `form:"token"`
}
