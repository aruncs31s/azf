package persistence

import (
	"time"

	"gorm.io/gorm"
)

// DBPoolConfig holds database connection pool configuration
type DBPoolConfig struct {
	// MaxIdleConns is the maximum number of idle connections
	MaxIdleConns int
	// MaxOpenConns is the maximum number of open connections
	MaxOpenConns int
	// ConnMaxLifetime is the maximum lifetime of a connection
	ConnMaxLifetime time.Duration
	// ConnMaxIdleTime is the maximum idle time for a connection
	ConnMaxIdleTime time.Duration
}

// DefaultDBPoolConfig returns sensible defaults for connection pooling
func DefaultDBPoolConfig() *DBPoolConfig {
	return &DBPoolConfig{
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	}
}

// ConfigureConnectionPool configures the connection pool for a GORM database
func ConfigureConnectionPool(db *gorm.DB, cfg *DBPoolConfig) error {
	if cfg == nil {
		cfg = DefaultDBPoolConfig()
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	return nil
}

// ConfigureConnectionPoolFromConfig configures the pool using the application config
func ConfigureConnectionPoolFromConfig(db *gorm.DB, maxIdle, maxOpen int, maxLifetime, maxIdleTime time.Duration) error {
	return ConfigureConnectionPool(db, &DBPoolConfig{
		MaxIdleConns:    maxIdle,
		MaxOpenConns:    maxOpen,
		ConnMaxLifetime: maxLifetime,
		ConnMaxIdleTime: maxIdleTime,
	})
}

// GetDBStats returns connection pool statistics
func GetDBStats(db *gorm.DB) (map[string]interface{}, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":               stats.InUse,
		"idle":                 stats.Idle,
		"wait_count":           stats.WaitCount,
		"wait_duration":        stats.WaitDuration.String(),
		"max_idle_closed":      stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}, nil
}

// HealthCheck performs a simple health check on the database
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
