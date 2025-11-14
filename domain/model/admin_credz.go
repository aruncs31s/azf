package model

import (
	"errors"
	"time"

	"github.com/aruncs31s/azf/utils"
)

// AdminCredentials represents the aggregate root for admin authentication
// Following DDD aggregate pattern
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

// ID returns the aggregate root ID
func (ac *AdminCredentials) ID() string {
	return ac.id
}

// Username returns the admin username
func (ac *AdminCredentials) Username() *AdminUsername {
	return ac.username
}

// Password returns the admin password
func (ac *AdminCredentials) Password() *AdminPassword {
	return ac.password
}

// CreatedAt returns when the credentials were created
func (ac *AdminCredentials) CreatedAt() time.Time {
	return ac.createdAt
}

// UpdatedAt returns when the credentials were last updated
func (ac *AdminCredentials) UpdatedAt() time.Time {
	return ac.updatedAt
}

// IsActive returns whether the credentials are active
func (ac *AdminCredentials) IsActive() bool {
	return ac.isActive
}

// Authenticate verifies if the provided username and password match
func (ac *AdminCredentials) Authenticate(username string, password string) error {
	if !ac.isActive {
		return ErrCredentialInactive
	}

	usernameObj, err := NewAdminUsername(username)
	if err != nil {
		return err
	}

	if !ac.username.Equals(usernameObj) {
		return utils.ErrInvalidUsername
	}

	if !ac.password.Matches(password) {
		return utils.ErrInvalidPassword
	}

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

// UpdatePassword updates the admin password
func (ac *AdminCredentials) UpdatePassword(newPassword *AdminPassword) error {
	if newPassword == nil {
		return ErrNilPassword
	}

	ac.password = newPassword
	ac.updatedAt = time.Now()
	return nil
}
