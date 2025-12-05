package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Username string `json:"username" gorm:"column:username;unique;not null"`
	Password string `json:"password" gorm:"column:password;not null"`
	Email    string `json:"email" gorm:"column:email;unique;not null"`
	Nickname string `json:"nickname" gorm:"column:nickname;not null"`
	Logo     string `json:"logo" gorm:"column:logo;not null"`
}

func (User) TableName() string {
	return "user"
}
