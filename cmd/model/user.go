package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username      string `gorm:"unique"`
	Password      string
	Email         string `gorm:"unique"`
	LastLogin     *time.Time
	OAuthProvider string `gorm:"size:50"`
	OAuthID       string `gorm:"size:255"`
}

func (User) TableName() string {
	return "user"
}
