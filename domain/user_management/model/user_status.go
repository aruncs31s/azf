package user_management

import "errors"

// UserStatus represents the status of a user as a value object
type UserStatus string

const (
	StatusActive   UserStatus = "ACTIVE"
	StatusBlocked  UserStatus = "BLOCKED"
	StatusPending  UserStatus = "PENDING"
	StatusDeleted  UserStatus = "DELETED"
	StatusSuspended UserStatus = "SUSPENDED"
)

// NewUserStatus creates a new UserStatus value object with validation
func NewUserStatus(status string) (UserStatus, error) {
	validStatuses := map[string]bool{
		string(StatusActive):    true,
		string(StatusBlocked):   true,
		string(StatusPending):   true,
		string(StatusDeleted):   true,
		string(StatusSuspended): true,
	}

	if !validStatuses[status] {
		return "", errors.New("invalid user status: " + status)
	}

	return UserStatus(status), nil
}

// String returns the string representation
func (s UserStatus) String() string {
	return string(s)
}

// IsActive checks if user is in active state
func (s UserStatus) IsActive() bool {
	return s == StatusActive
}

// IsBlocked checks if user is blocked
func (s UserStatus) IsBlocked() bool {
	return s == StatusBlocked
}

// IsPending checks if user is pending verification
func (s UserStatus) IsPending() bool {
	return s == StatusPending
}

// IsDeleted checks if user is deleted
func (s UserStatus) IsDeleted() bool {
	return s == StatusDeleted
}

// IsSuspended checks if user is suspended
func (s UserStatus) IsSuspended() bool {
	return s == StatusSuspended
}

// CanLogin checks if user can login based on their status
func (s UserStatus) CanLogin() bool {
	return s == StatusActive
}

// CanPerformActions checks if user can perform actions based on their status
func (s UserStatus) CanPerformActions() bool {
	return s == StatusActive
}
