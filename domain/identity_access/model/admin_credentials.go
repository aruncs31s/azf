package identity_access

import (
	"errors"
	"time"
)

// AdminCredentials represents the aggregate root for admin authentication
// Following DDD aggregate pattern within the Identity & Access bounded context
type AdminCredentials struct {
	id        string
	username  *AdminUsername
	password  *AdminPassword
	createdAt time.Time
	updatedAt time.Time
	isActive  bool
}

var (
	ErrNilUsername        = errors.New("username cannot be nil")
	ErrNilPassword        = errors.New("password cannot be nil")
	ErrInvalidID          = errors.New("id cannot be empty")
	ErrCredentialInactive = errors.New("admin credentials are inactive")
)

// NewAdminCredentials creates a new AdminCredentials aggregate root
func NewAdminCredentials(
	id string,
	username *AdminUsername,
	password *AdminPassword,
) (*AdminCredentials, error) {
	if id == "" {
		return nil, ErrInvalidID
	}

	if username == nil {
		return nil, ErrNilUsername
	}

	if password == nil {
		return nil, ErrNilPassword
	}

	return &AdminCredentials{
		id:        id,
		username:  username,
		password:  password,
		createdAt: time.Now(),
		updatedAt: time.Now(),
		isActive:  true,
	}, nil
}

// GetID returns the aggregate root ID
func (ac *AdminCredentials) GetID() string {
	return ac.id
}

// GetUsername returns the username value object
func (ac *AdminCredentials) GetUsername() *AdminUsername {
	return ac.username
}

// GetPassword returns the password value object
func (ac *AdminCredentials) GetPassword() *AdminPassword {
	return ac.password
}

// GetCreatedAt returns the creation timestamp
func (ac *AdminCredentials) GetCreatedAt() time.Time {
	return ac.createdAt
}

// GetUpdatedAt returns the last update timestamp
func (ac *AdminCredentials) GetUpdatedAt() time.Time {
	return ac.updatedAt
}

// IsActive checks if the credential is active
func (ac *AdminCredentials) IsActive() bool {
	return ac.isActive
}

// UpdatePassword updates the password with new value
func (ac *AdminCredentials) UpdatePassword(newPassword *AdminPassword) error {
	if newPassword == nil {
		return ErrNilPassword
	}

	ac.password = newPassword
	ac.updatedAt = time.Now()
	return nil
}

// Deactivate deactivates the credentials
func (ac *AdminCredentials) Deactivate() {
	ac.isActive = false
	ac.updatedAt = time.Now()
}

// Activate activates the credentials
func (ac *AdminCredentials) Activate() {
	ac.isActive = true
	ac.updatedAt = time.Now()
}

// VerifyPassword verifies if the provided password matches
func (ac *AdminCredentials) VerifyPassword(rawPassword string) (bool, error) {
	if !ac.isActive {
		return false, ErrCredentialInactive
	}

	return ac.password.Verify(rawPassword)
}

// UpdateUsername updates the username
func (ac *AdminCredentials) UpdateUsername(newUsername *AdminUsername) error {
	if newUsername == nil {
		return ErrNilUsername
	}

	ac.username = newUsername
	ac.updatedAt = time.Now()
	return nil
}
