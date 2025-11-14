package user_management

import (
	"context"
	"time"
)

// UserSearchFilter contains search and filter parameters
type UserSearchFilter struct {
	Status    *UserStatus
	IsAdmin   *bool
	RoleName  *string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	LastLoginAfter *time.Time
	Limit     int
	Offset    int
}

// UserSearchResult contains search results and metadata
type UserSearchResult struct {
	Users      []*User
	Total      int64
	HasMore    bool
	Limit      int
	Offset     int
}

// UserRepository defines the interface for user management persistence operations
type UserRepository interface {
	UserReader
	UserWriter
}

// UserReader defines read-only operations for user management
type UserReader interface {
	// GetByID retrieves a user by their unique identifier
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	GetByID(ctx context.Context, userID string) (*User, error)

	// GetByEmail retrieves a user by their email address
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByUsername retrieves a user by their username
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	GetByUsername(ctx context.Context, username string) (*User, error)

	// Search searches for users based on query and filters
	// Returns results with pagination support
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	Search(ctx context.Context, query string, filter *UserSearchFilter) (*UserSearchResult, error)

	// ListAll retrieves all users with pagination
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	ListAll(ctx context.Context, limit, offset int) ([]*User, int64, error)

	// GetAdmins retrieves all admin users
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	GetAdmins(ctx context.Context) ([]*User, error)

	// GetByRole retrieves all users with a specific role
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	GetByRole(ctx context.Context, roleName string) ([]*User, error)

	// GetByStatus retrieves all users with a specific status
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	GetByStatus(ctx context.Context, status UserStatus) ([]*User, error)
}

// UserWriter defines write operations for user management
type UserWriter interface {
	// Create adds a new user to the repository
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	Create(ctx context.Context, user *User) (*User, error)

	// Update modifies an existing user
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	Update(ctx context.Context, user *User) (*User, error)

	// Delete removes a user from the repository
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	Delete(ctx context.Context, userID string) error

	// Block marks a user as blocked
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	Block(ctx context.Context, userID string, reason string) (*User, error)

	// Unblock removes the blocked status from a user
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	Unblock(ctx context.Context, userID string) (*User, error)

	// AssignRole adds a role to a user
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	AssignRole(ctx context.Context, userID string, role *UserRole) (*User, error)

	// RemoveRole removes a role from a user
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	RemoveRole(ctx context.Context, userID string, role *UserRole) (*User, error)

	// PromoteToAdmin grants admin privileges to a user
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	PromoteToAdmin(ctx context.Context, userID string) (*User, error)

	// DemoteFromAdmin removes admin privileges from a user
	// ctx is used to manage the request lifetime, handle cancellation, and pass deadlines
	DemoteFromAdmin(ctx context.Context, userID string) (*User, error)
}
