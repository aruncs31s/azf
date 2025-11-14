package user_management

import (
	"errors"
	"regexp"
)

// UserEmail represents an email address as a value object
type UserEmail struct {
	value string
}

// NewUserEmail creates a new UserEmail value object with validation
func NewUserEmail(email string) (*UserEmail, error) {
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}
	if len(email) > 254 {
		return nil, errors.New("email cannot exceed 254 characters")
	}

	// Basic email validation using regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return nil, errors.New("invalid email format")
	}

	return &UserEmail{value: email}, nil
}

// Value returns the underlying email value
func (u *UserEmail) Value() string {
	if u == nil {
		return ""
	}
	return u.value
}

// String returns the string representation of UserEmail
func (u *UserEmail) String() string {
	return u.value
}

// Equals checks if two emails are equal
func (u *UserEmail) Equals(other *UserEmail) bool {
	if u == nil || other == nil {
		return u == other
	}
	return u.value == other.value
}
