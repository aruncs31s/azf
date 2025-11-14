package model

import (
	"errors"
	"strings"
	"time"
)

// AdminProfile represents the admin user profile entity
// Following DDD entity pattern with identity and lifecycle
type AdminProfile struct {
	id        string
	fullName  *AdminFullName
	email     *AdminEmail
	role      string
	lastLogin *time.Time
	createdAt time.Time
	updatedAt time.Time
	isActive  bool
}

var (
	ErrNilFullName     = errors.New("full name cannot be nil")
	ErrNilEmail        = errors.New("email cannot be nil")
	ErrInvalidRole     = errors.New("role cannot be empty")
	ErrProfileInactive = errors.New("admin profile is inactive")
)

// AdminFullName represents a valid admin full name
// Following DDD value object pattern
type AdminFullName struct {
	value string
}

// NewAdminFullName creates a new AdminFullName value object
func NewAdminFullName(value string) (*AdminFullName, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil, errors.New("full name cannot be empty")
	}

	if len(value) > 100 {
		return nil, errors.New("full name cannot exceed 100 characters")
	}

	return &AdminFullName{value: value}, nil
}

// Value returns the full name string
func (n *AdminFullName) Value() string {
	return n.value
}

// Equals checks if two AdminFullName objects are equal
func (n *AdminFullName) Equals(other *AdminFullName) bool {
	if other == nil {
		return false
	}
	return n.value == other.value
}

// String returns the full name value
func (n *AdminFullName) String() string {
	return n.value
}

// AdminEmail represents a valid admin email
// Following DDD value object pattern
type AdminEmail struct {
	value string
}

// NewAdminEmail creates a new AdminEmail value object
func NewAdminEmail(value string) (*AdminEmail, error) {
	value = strings.TrimSpace(strings.ToLower(value))

	if value == "" {
		return nil, errors.New("email cannot be empty")
	}

	if len(value) > 255 {
		return nil, errors.New("email cannot exceed 255 characters")
	}

	// Basic email validation
	if !strings.Contains(value, "@") || !strings.Contains(value, ".") {
		return nil, errors.New("invalid email format")
	}

	return &AdminEmail{value: value}, nil
}

// Value returns the email string
func (e *AdminEmail) Value() string {
	return e.value
}

// Equals checks if two AdminEmail objects are equal
func (e *AdminEmail) Equals(other *AdminEmail) bool {
	if other == nil {
		return false
	}
	return e.value == other.value
}

// Domain returns the domain part of the email
func (e *AdminEmail) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// String returns the email value
func (e *AdminEmail) String() string {
	return e.value
}

// NewAdminProfile creates a new AdminProfile entity
func NewAdminProfile(
	id string,
	fullName *AdminFullName,
	email *AdminEmail,
	role string,
) (*AdminProfile, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	if fullName == nil {
		return nil, ErrNilFullName
	}

	if email == nil {
		return nil, ErrNilEmail
	}

	role = strings.TrimSpace(role)
	if role == "" {
		return nil, ErrInvalidRole
	}

	now := time.Now()
	return &AdminProfile{
		id:        id,
		fullName:  fullName,
		email:     email,
		role:      role,
		createdAt: now,
		updatedAt: now,
		isActive:  true,
	}, nil
}

// ID returns the profile ID
func (ap *AdminProfile) ID() string {
	return ap.id
}

// FullName returns the admin's full name
func (ap *AdminProfile) FullName() *AdminFullName {
	return ap.fullName
}

// Email returns the admin's email
func (ap *AdminProfile) Email() *AdminEmail {
	return ap.email
}

// Role returns the admin's role
func (ap *AdminProfile) Role() string {
	return ap.role
}

// LastLogin returns the last login time
func (ap *AdminProfile) LastLogin() *time.Time {
	return ap.lastLogin
}

// CreatedAt returns when the profile was created
func (ap *AdminProfile) CreatedAt() time.Time {
	return ap.createdAt
}

// UpdatedAt returns when the profile was last updated
func (ap *AdminProfile) UpdatedAt() time.Time {
	return ap.updatedAt
}

// IsActive returns whether the profile is active
func (ap *AdminProfile) IsActive() bool {
	return ap.isActive
}

// UpdateFullName updates the admin's full name
func (ap *AdminProfile) UpdateFullName(newName *AdminFullName) error {
	if newName == nil {
		return ErrNilFullName
	}

	ap.fullName = newName
	ap.updatedAt = time.Now()
	return nil
}

// UpdateEmail updates the admin's email
func (ap *AdminProfile) UpdateEmail(newEmail *AdminEmail) error {
	if newEmail == nil {
		return ErrNilEmail
	}

	ap.email = newEmail
	ap.updatedAt = time.Now()
	return nil
}

// UpdateRole updates the admin's role
func (ap *AdminProfile) UpdateRole(newRole string) error {
	newRole = strings.TrimSpace(newRole)
	if newRole == "" {
		return ErrInvalidRole
	}

	ap.role = newRole
	ap.updatedAt = time.Now()
	return nil
}

// RecordLogin records a successful login
func (ap *AdminProfile) RecordLogin() {
	now := time.Now()
	ap.lastLogin = &now
	ap.updatedAt = now
}

// Deactivate deactivates the profile
func (ap *AdminProfile) Deactivate() {
	ap.isActive = false
	ap.updatedAt = time.Now()
}

// Activate activates the profile
func (ap *AdminProfile) Activate() {
	ap.isActive = true
	ap.updatedAt = time.Now()
}

// DisplayName returns a display-friendly name
func (ap *AdminProfile) DisplayName() string {
	if ap.fullName != nil {
		return ap.fullName.Value()
	}
	return ap.id
}

// ProfileSummary returns a summary of the profile for display
func (ap *AdminProfile) ProfileSummary() map[string]interface{} {
	summary := map[string]interface{}{
		"id":         ap.id,
		"full_name":  ap.fullName.Value(),
		"email":      ap.email.Value(),
		"role":       ap.role,
		"is_active":  ap.isActive,
		"created_at": ap.createdAt,
		"updated_at": ap.updatedAt,
	}

	if ap.lastLogin != nil {
		summary["last_login"] = ap.lastLogin
	}

	return summary
}
