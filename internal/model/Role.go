package model

import (
	"gorm.io/gorm"
)

type Role struct {
	gorm.Model

	Name string `json:"name" gorm:"column:name;type:varchar(255);not null"`
}

func (Role) TableName() string {
	return "role"
}
