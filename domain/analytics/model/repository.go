package analytics

import (
	"context"
	"time"
)

// APIUsageRepository defines the interface for API usage log persistence operations
type APIUsageRepository interface {
	APIUsageReader
	APIUsageWriter
}

// APIUsageReader defines the interface for API usage read operations
type APIUsageReader interface {
	// FindByID retrieves an API usage log by ID
	FindByID(ctx context.Context, id string) (*APIUsageLog, error)

	// FindByEndpoint retrieves all usage logs for a specific endpoint
	FindByEndpoint(ctx context.Context, endpoint string) ([]*APIUsageLog, error)

	// FindByUserID retrieves all usage logs for a specific user
	FindByUserID(ctx context.Context, userID string) ([]*APIUsageLog, error)

	// FindByDateRange retrieves usage logs within a date range
	FindByDateRange(ctx context.Context, startDate time.Time, endDate time.Time) ([]*APIUsageLog, error)

	// FindAll retrieves all API usage logs
	FindAll(ctx context.Context) ([]*APIUsageLog, error)
}

// APIUsageWriter defines the interface for API usage write operations
type APIUsageWriter interface {
	// Create adds a new API usage log
	Create(ctx context.Context, log *APIUsageLog) (*APIUsageLog, error)

	// Update modifies an API usage log
	Update(ctx context.Context, log *APIUsageLog) (*APIUsageLog, error)

	// Delete removes an API usage log
	Delete(ctx context.Context, id string) error
}
