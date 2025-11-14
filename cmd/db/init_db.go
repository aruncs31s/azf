package db

import (
	"github.com/aruncs31s/azf/cmd/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic("No DB")
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		panic("hi")
	}
	return db
}
