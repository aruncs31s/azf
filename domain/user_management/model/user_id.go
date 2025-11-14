package user_management

import "errors"

// UserID represents a unique user identifier as a value object
type UserID struct {
	value string
}

// NewUserID creates a new UserID value object
func NewUserID(id string) (*UserID, error) {
	if id == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if len(id) > 36 {
		return nil, errors.New("user ID cannot exceed 36 characters")
	}
	return &UserID{value: id}, nil
}

// Value returns the underlying ID value
func (u *UserID) Value() string {
	if u == nil {
		return ""
	}
	return u.value
}

// String returns the string representation of UserID
func (u *UserID) String() string {
	return u.value
}

// Equals checks if two UserIDs are equal
func (u *UserID) Equals(other *UserID) bool {
	if u == nil || other == nil {
		return u == other
	}
	return u.value == other.value
}
