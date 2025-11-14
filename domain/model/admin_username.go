package model

import (
	"strings"

	"github.com/aruncs31s/azf/utils"
)

// AdminUsername represents a valid admin username
// Following DDD value object pattern
type AdminUsername struct {
	value string
}

// NewAdminUsername creates a new AdminUsername value object
func NewAdminUsername(value string) (*AdminUsername, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil, utils.ErrInvalidUsername
	}

	if len(value) > 100 {
		return nil, utils.ErrUsernameTooLong
	}

	return &AdminUsername{value: value}, nil
}

// Value returns the username string
func (u *AdminUsername) Value() string {
	return u.value
}

// Equals checks if two AdminUsername objects are equal
func (u *AdminUsername) Equals(other *AdminUsername) bool {
	if other == nil {
		return false
	}
	return u.value == other.value
}

// Matches checks if the username matches the given pattern
func (u *AdminUsername) Matches(pattern string) bool {
	return strings.HasPrefix(u.value, pattern)
}

// String returns the username value
func (u *AdminUsername) String() string {
	return u.value
}
