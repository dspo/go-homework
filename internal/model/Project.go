package model

import (
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model

	Name   string `json:"name" gorm:"column:name; not null"`
	Desc   string `json:"desc" gorm:"column:description; not null"`
	Status string `json:"status" gorm:"column:status; not null"`
}
