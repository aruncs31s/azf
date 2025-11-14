package initializer

import (
	"fmt"

	"github.com/aruncs31s/azf/shared/logger"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type makeDB struct {
	DB  *gorm.DB
	Err error
}

var DB *gorm.DB

func InitLocalDB(db *gorm.DB) {
	// Check if the DB is nil
	//
	if db != nil {
		return
	}

	chResultCreateDB := make(chan makeDB)
	go func() {
		localDBPath := "tmp/AZF_auth_z.db"
		// db, err := gorm.Open(sqlite.Open(configs.RegApplicationDBFilePath), &gorm.Config{})
		db, err := gorm.Open(sqlite.Open(localDBPath), &gorm.Config{
			SkipDefaultTransaction: true,
		})

		logger.Debug(
			"Initalized Local DB",
			zap.String("db type", "sql lite"),
			zap.String("path", localDBPath),
		)
		chResultCreateDB <- makeDB{DB: db, Err: err}
	}()

	result := <-chResultCreateDB

	if result.Err != nil {
		// Log the error and attempt a fallback to an in-memory SQLite database.
		// This helps tests and environments where the file-based DB path is not writable.
		logger.Error("Error initializing SQLite database, attempting in-memory fallback", zap.Error(result.Err))

		memDB, memErr := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		if memErr != nil {
			// If fallback also fails, log and leave DB nil so callers can handle absence.
			logger.Error("Failed to initialize in-memory SQLite fallback", zap.Error(memErr))
			DB = nil
			return
		}

		DB = memDB
	} else {
		DB = result.DB
	}

	// Run migrations on whichever DB we have (file-based or in-memory)
	migrateTable(DB)
	fmt.Println("Database Initialized")

}
