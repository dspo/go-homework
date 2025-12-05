package model

import (
	"time"

	"gorm.io/gorm"
)

type Audit struct {
	gorm.Model

	UserId     uint      `json:"user_id" gorm:"column:user_id"`
	OperatedAt time.Time `json:"operated_at" gorm:"column:operated_at"`
	Method     string    `json:"method" gorm:"column:method"`
	URL        string    `json:"url" gorm:"column:url"`
	Result     string    `json:"result" gorm:"column:result"`
}

func (Audit) TableName() string {
	return "audit"
}
