package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username  string `gorm:"unique"`
	Password  string
	Email     string `gorm:"unique"`
	LastLogin *time.Time
}

func (User) TableName() string {
	return "user"
}
