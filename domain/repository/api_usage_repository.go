package repository

import "github.com/aruncs31s/azf/domain/api_usage"

// APIUsageLogReader defines read operations for API usage logs
type APIUsageLogReader interface {
	FindByID(id string) (*api_usage.APIUsageLog, error)
	FindAll(limit int, offset int) (*[]api_usage.APIUsageLog, error)
	FindByEndpoint(endpoint string, limit int, offset int) (*[]api_usage.APIUsageLog, error)
	FindByUserID(userID string, limit int, offset int) (*[]api_usage.APIUsageLog, error)
	FindByDateRange(startDate string, endDate string, limit int, offset int) (*[]api_usage.APIUsageLog, error)
	CountByEndpoint(endpoint string) (int64, error)
	CountTotal() (int64, error)
}

// APIUsageLogWriter defines write operations for API usage logs
type APIUsageLogWriter interface {
	Create(
		log *api_usage.APIUsageLog,
	) (*api_usage.APIUsageLog, error)
	BatchCreate(logs *[]api_usage.APIUsageLog) error
	DeleteOlderThan(days int) error
	DeleteAll() error
}

// APIUsageLogRepository combines read and write operations
type APIUsageLogRepository interface {
	APIUsageLogReader
	APIUsageLogWriter
}

// APIUsageStatsReader defines read operations for API usage statistics
type APIUsageStatsReader interface {
	FindByID(id string) (*api_usage.APIUsageStats, error)
	FindAll(limit int, offset int) (*[]api_usage.APIUsageStats, error)
	FindByEndpoint(endpoint string) (*api_usage.APIUsageStats, error)
	GetTopEndpointsByUsage(limit int) (*[]api_usage.APIEndpointRanking, error)
	GetEndpointsByErrorRate(limit int) (*[]api_usage.APIEndpointRanking, error)
	GetEndpointsByResponseTime(limit int) (*[]api_usage.APIEndpointRanking, error)
	CountTotal() (int64, error)
}

// APIUsageStatsWriter defines write operations for API usage statistics
type APIUsageStatsWriter interface {
	Create(stats *api_usage.APIUsageStats) (*api_usage.APIUsageStats, error)
	Update(stats *api_usage.APIUsageStats) (*api_usage.APIUsageStats, error)
	Upsert(stats *api_usage.APIUsageStats) (*api_usage.APIUsageStats, error)
	RecalculateStats(endpoint string, method string) error
	DeleteAll() error
}

// APIUsageStatsRepository combines read and write operations
type APIUsageStatsRepository interface {
	APIUsageStatsReader
	APIUsageStatsWriter
}
