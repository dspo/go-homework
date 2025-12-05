package model

import (
	"gorm.io/gorm"
)

type Team struct {
	gorm.Model

	Name     string `json:"name" gorm:"column:name;unique"`
	Desc     string `json:"desc" gorm:"column:description"`
	LeaderId uint   `json:"-" gorm:"column:leader_id"`

	Leader *User `json:"leader" gorm:"foreignkey:LeaderId"`
}

func (Team) TableName() string {
	return "team"
}
