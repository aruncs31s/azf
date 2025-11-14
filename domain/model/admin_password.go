package model

import (
	"errors"
	"strings"
)

// AdminPassword represents a valid admin password
// Following DDD value object pattern
type AdminPassword struct {
	value string
}

var (
	ErrInvalidPassword  = errors.New("password cannot be empty")
	ErrPasswordTooShort = errors.New("password must be at least 6 characters")
	ErrPasswordTooLong  = errors.New("password cannot exceed 255 characters")
)

// NewAdminPassword creates a new AdminPassword value object
func NewAdminPassword(value string) (*AdminPassword, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil, ErrInvalidPassword
	}

	if len(value) < 6 {
		return nil, ErrPasswordTooShort
	}

	if len(value) > 255 {
		return nil, ErrPasswordTooLong
	}

	return &AdminPassword{value: value}, nil
}

// Value returns the password string
func (p *AdminPassword) Value() string {
	return p.value
}

// Equals checks if two AdminPassword objects have the same value
func (p *AdminPassword) Equals(other *AdminPassword) bool {
	if other == nil {
		return false
	}
	return p.value == other.value
}

// Matches checks if the provided password string matches this password
func (p *AdminPassword) Matches(other string) bool {
	return p.value == other
}

// String returns the password value
func (p *AdminPassword) String() string {
	return p.value
}
