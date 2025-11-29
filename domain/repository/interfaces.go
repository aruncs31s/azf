package repository

import (
	"context"
)

// User represents a user entity
type User struct {
	ID        string
	Email     string
	Username  string
	Role      string
	CreatedAt int64
	UpdatedAt int64
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// FindByID finds a user by their ID
	FindByID(ctx context.Context, id string) (*User, error)

	// FindByEmail finds a user by their email
	FindByEmail(ctx context.Context, email string) (*User, error)

	// FindByUsername finds a user by their username
	FindByUsername(ctx context.Context, username string) (*User, error)

	// Create creates a new user
	Create(ctx context.Context, user *User) error

	// Update updates an existing user
	Update(ctx context.Context, user *User) error

	// Delete deletes a user by ID
	Delete(ctx context.Context, id string) error

	// List lists users with pagination
	List(ctx context.Context, offset, limit int) ([]*User, error)

	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)
}

// Role represents a role entity
type Role struct {
	Name        string
	Description string
	Permissions []string
}

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	// FindByName finds a role by name
	FindByName(ctx context.Context, name string) (*Role, error)

	// Create creates a new role
	Create(ctx context.Context, role *Role) error

	// Update updates an existing role
	Update(ctx context.Context, role *Role) error

	// Delete deletes a role by name
	Delete(ctx context.Context, name string) error

	// List lists all roles
	List(ctx context.Context) ([]*Role, error)

	// GetUsersWithRole returns users that have a specific role
	GetUsersWithRole(ctx context.Context, roleName string) ([]*User, error)
}

// APIUsageLog represents an API usage log entry
type APIUsageLog struct {
	ID           string
	UserID       string
	Endpoint     string
	Method       string
	StatusCode   int
	ResponseTime int64
	RequestedAt  int64
	IPAddress    string
	UserAgent    string
}

// APIUsageRepository defines the interface for API usage tracking
type APIUsageRepository interface {
	// Log logs an API usage entry
	Log(ctx context.Context, log *APIUsageLog) error

	// GetByUser retrieves usage logs for a user
	GetByUser(ctx context.Context, userID string, offset, limit int) ([]*APIUsageLog, error)

	// GetByEndpoint retrieves usage logs for an endpoint
	GetByEndpoint(ctx context.Context, endpoint string, offset, limit int) ([]*APIUsageLog, error)

	// GetStats retrieves aggregated statistics
	GetStats(ctx context.Context, startTime, endTime int64) (*APIUsageStats, error)

	// Count returns the total number of logs
	Count(ctx context.Context) (int64, error)

	// Cleanup removes logs older than the specified time
	Cleanup(ctx context.Context, olderThan int64) (int64, error)
}

// APIUsageStats holds aggregated API usage statistics
type APIUsageStats struct {
	TotalRequests     int64
	UniqueUsers       int64
	AverageResponseMs float64
	ErrorCount        int64
	TopEndpoints      map[string]int64
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        string
	UserID    string
	Action    string
	Resource  string
	Result    string
	Details   map[string]interface{}
	IPAddress string
	Timestamp int64
}

// AuditRepository defines the interface for audit logging
type AuditRepository interface {
	// Log creates an audit log entry
	Log(ctx context.Context, log *AuditLog) error

	// FindByUser retrieves audit logs for a user
	FindByUser(ctx context.Context, userID string, offset, limit int) ([]*AuditLog, error)

	// FindByAction retrieves audit logs by action type
	FindByAction(ctx context.Context, action string, offset, limit int) ([]*AuditLog, error)

	// FindDenied retrieves denied access audit logs
	FindDenied(ctx context.Context, offset, limit int) ([]*AuditLog, error)

	// Count returns the total number of audit logs
	Count(ctx context.Context) (int64, error)

	// Cleanup removes logs older than the specified time
	Cleanup(ctx context.Context, olderThan int64) (int64, error)
}

// CasbinEnforcer defines the interface for Casbin operations
// This allows for mocking in tests
type CasbinEnforcer interface {
	// Enforce checks if a user has permission
	Enforce(rvals ...interface{}) (bool, error)

	// AddPolicy adds a policy rule
	AddPolicy(params ...interface{}) (bool, error)

	// RemovePolicy removes a policy rule
	RemovePolicy(params ...interface{}) (bool, error)

	// AddGroupingPolicy adds a role assignment
	AddGroupingPolicy(params ...interface{}) (bool, error)

	// RemoveGroupingPolicy removes a role assignment
	RemoveGroupingPolicy(params ...interface{}) (bool, error)

	// GetRolesForUser returns roles for a user
	GetRolesForUser(name string, domain ...string) ([]string, error)

	// GetUsersForRole returns users for a role
	GetUsersForRole(name string, domain ...string) ([]string, error)

	// GetPolicy returns all policy rules
	GetPolicy() ([][]string, error)

	// LoadPolicy reloads the policy from storage
	LoadPolicy() error

	// SavePolicy saves the policy to storage
	SavePolicy() error
}

// Transaction represents a database transaction
type Transaction interface {
	// Commit commits the transaction
	Commit() error

	// Rollback rolls back the transaction
	Rollback() error
}

// TransactionManager manages database transactions
type TransactionManager interface {
	// Begin starts a new transaction
	Begin(ctx context.Context) (Transaction, error)

	// WithTransaction executes a function within a transaction
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
