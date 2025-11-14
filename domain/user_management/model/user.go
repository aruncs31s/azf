package user_management

import (
	"errors"
	"time"
)

// User represents a domain user as an aggregate root
// It encapsulates user data and business logic for user management
type User struct {
	id            *UserID
	email         *UserEmail
	username      string
	displayName   string
	status        UserStatus
	roles         []*UserRole
	isAdmin       bool
	createdAt     time.Time
	updatedAt     time.Time
	lastLoginAt   *time.Time
	blockedReason string
	metadata      map[string]interface{}
}

// NewUser creates a new User aggregate root
func NewUser(
	id string,
	email string,
	username string,
	displayName string,
) (*User, error) {
	// Validate inputs
	userID, err := NewUserID(id)
	if err != nil {
		return nil, err
	}

	userEmail, err := NewUserEmail(email)
	if err != nil {
		return nil, err
	}

	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if len(username) > 50 {
		return nil, errors.New("username cannot exceed 50 characters")
	}

	if displayName == "" {
		displayName = username
	}
	if len(displayName) > 100 {
		return nil, errors.New("display name cannot exceed 100 characters")
	}

	return &User{
		id:          userID,
		email:       userEmail,
		username:    username,
		displayName: displayName,
		status:      StatusActive,
		roles:       make([]*UserRole, 0),
		isAdmin:     false,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
		metadata:    make(map[string]interface{}),
	}, nil
}

// GetID returns the user ID
func (u *User) GetID() string {
	if u == nil || u.id == nil {
		return ""
	}
	return u.id.Value()
}

// GetEmail returns the user email
func (u *User) GetEmail() string {
	if u == nil || u.email == nil {
		return ""
	}
	return u.email.Value()
}

// GetUsername returns the username
func (u *User) GetUsername() string {
	if u == nil {
		return ""
	}
	return u.username
}

// GetDisplayName returns the display name
func (u *User) GetDisplayName() string {
	if u == nil {
		return ""
	}
	return u.displayName
}

// SetDisplayName updates the display name
func (u *User) SetDisplayName(displayName string) error {
	if displayName == "" {
		return errors.New("display name cannot be empty")
	}
	if len(displayName) > 100 {
		return errors.New("display name cannot exceed 100 characters")
	}
	u.displayName = displayName
	u.updatedAt = time.Now()
	return nil
}

// GetStatus returns the user status
func (u *User) GetStatus() UserStatus {
	if u == nil {
		return StatusDeleted
	}
	return u.status
}

// SetStatus updates the user status
func (u *User) SetStatus(status UserStatus) error {
	if !status.IsValid() {
		return errors.New("invalid user status")
	}
	u.status = status
	u.updatedAt = time.Now()
	return nil
}

// IsValid checks if status is valid (internal helper)
func (s UserStatus) IsValid() bool {
	return s == StatusActive || s == StatusBlocked || s == StatusPending || s == StatusDeleted || s == StatusSuspended
}

// GetRoles returns a copy of the user roles
func (u *User) GetRoles() []*UserRole {
	if u == nil {
		return make([]*UserRole, 0)
	}
	rolesCopy := make([]*UserRole, len(u.roles))
	copy(rolesCopy, u.roles)
	return rolesCopy
}

// AssignRole adds a role to the user
func (u *User) AssignRole(role *UserRole) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	if role == nil {
		return errors.New("role cannot be nil")
	}

	// Check if role already assigned
	for _, r := range u.roles {
		if r.Equals(role) {
			return errors.New("role already assigned to user")
		}
	}

	// Max 10 roles per user
	if len(u.roles) >= 10 {
		return errors.New("cannot assign more than 10 roles to a user")
	}

	u.roles = append(u.roles, role)
	u.updatedAt = time.Now()
	return nil
}

// RemoveRole removes a role from the user
func (u *User) RemoveRole(role *UserRole) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	if role == nil {
		return errors.New("role cannot be nil")
	}

	for i, r := range u.roles {
		if r.Equals(role) {
			u.roles = append(u.roles[:i], u.roles[i+1:]...)
			u.updatedAt = time.Now()
			return nil
		}
	}

	return errors.New("role not found on user")
}

// HasRole checks if user has a specific role
func (u *User) HasRole(roleName string) bool {
	if u == nil {
		return false
	}
	for _, r := range u.roles {
		if r.Name() == roleName {
			return true
		}
	}
	return false
}

// PromoteToAdmin makes the user an admin
func (u *User) PromoteToAdmin() error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	u.isAdmin = true
	u.updatedAt = time.Now()
	return nil
}

// DemoteFromAdmin removes admin privileges
func (u *User) DemoteFromAdmin() error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	u.isAdmin = false
	u.updatedAt = time.Now()
	return nil
}

// IsAdmin checks if user is an admin
func (u *User) IsAdmin() bool {
	if u == nil {
		return false
	}
	return u.isAdmin
}

// Block marks the user as blocked
func (u *User) Block(reason string) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	if reason == "" {
		return errors.New("block reason cannot be empty")
	}
	if len(reason) > 500 {
		return errors.New("block reason cannot exceed 500 characters")
	}
	u.status = StatusBlocked
	u.blockedReason = reason
	u.updatedAt = time.Now()
	return nil
}

// Unblock removes the blocked status
func (u *User) Unblock() error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	u.status = StatusActive
	u.blockedReason = ""
	u.updatedAt = time.Now()
	return nil
}

// GetBlockedReason returns the reason user was blocked
func (u *User) GetBlockedReason() string {
	if u == nil {
		return ""
	}
	return u.blockedReason
}

// RecordLogin updates the last login timestamp
func (u *User) RecordLogin() error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	if !u.status.CanLogin() {
		return errors.New("user cannot login in current status")
	}
	now := time.Now()
	u.lastLoginAt = &now
	u.updatedAt = now
	return nil
}

// GetLastLoginAt returns the last login time
func (u *User) GetLastLoginAt() *time.Time {
	if u == nil {
		return nil
	}
	return u.lastLoginAt
}

// GetCreatedAt returns the creation time
func (u *User) GetCreatedAt() time.Time {
	if u == nil {
		return time.Time{}
	}
	return u.createdAt
}

// GetUpdatedAt returns the last update time
func (u *User) GetUpdatedAt() time.Time {
	if u == nil {
		return time.Time{}
	}
	return u.updatedAt
}

// SetMetadata sets metadata for the user
func (u *User) SetMetadata(key string, value interface{}) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	if key == "" {
		return errors.New("metadata key cannot be empty")
	}
	u.metadata[key] = value
	u.updatedAt = time.Now()
	return nil
}

// GetAllMetadata returns a copy of all metadata
func (u *User) GetAllMetadata() map[string]interface{} {
	if u == nil {
		return make(map[string]interface{})
	}
	metadataCopy := make(map[string]interface{})
	for k, v := range u.metadata {
		metadataCopy[k] = v
	}
	return metadataCopy
}

// SetAllMetadata sets all metadata (used for loading from persistence)
func (u *User) SetAllMetadata(metadata map[string]interface{}) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	u.metadata = make(map[string]interface{})
	for k, v := range metadata {
		u.metadata[k] = v
	}
	return nil
}

// GetMetadata retrieves metadata
func (u *User) GetMetadata(key string) (interface{}, bool) {
	if u == nil {
		return nil, false
	}
	val, exists := u.metadata[key]
	return val, exists
}

// CanLogin checks if user can log in
func (u *User) CanLogin() bool {
	if u == nil {
		return false
	}
	return u.status.CanLogin()
}

// CanPerformActions checks if user can perform actions
func (u *User) CanPerformActions() bool {
	if u == nil {
		return false
	}
	return u.status.CanPerformActions()
}

// SetID sets the user ID (used for loading from persistence)
func (u *User) SetID(id string) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	userID, err := NewUserID(id)
	if err != nil {
		return err
	}
	u.id = userID
	return nil
}

// SetEmail sets the user email (used for loading from persistence)
func (u *User) SetEmail(email string) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	userEmail, err := NewUserEmail(email)
	if err != nil {
		return err
	}
	u.email = userEmail
	return nil
}

// SetUsername sets the username (used for loading from persistence)
func (u *User) SetUsername(username string) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	if username == "" {
		return errors.New("username cannot be empty")
	}
	if len(username) > 50 {
		return errors.New("username cannot exceed 50 characters")
	}
	u.username = username
	return nil
}

// SetRoles sets the user roles (used for loading from persistence)
func (u *User) SetRoles(roles []*UserRole) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	if len(roles) > 10 {
		return errors.New("cannot assign more than 10 roles")
	}
	u.roles = make([]*UserRole, len(roles))
	copy(u.roles, roles)
	return nil
}

// SetIsAdmin sets the admin flag (used for loading from persistence)
func (u *User) SetIsAdmin(isAdmin bool) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	u.isAdmin = isAdmin
	return nil
}

// SetCreatedAt sets the creation time (used for loading from persistence)
func (u *User) SetCreatedAt(createdAt time.Time) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	u.createdAt = createdAt
	return nil
}

// SetUpdatedAt sets the update time (used for loading from persistence)
func (u *User) SetUpdatedAt(updatedAt time.Time) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	u.updatedAt = updatedAt
	return nil
}

// SetLastLoginAt sets the last login time (used for loading from persistence)
func (u *User) SetLastLoginAt(lastLoginAt *time.Time) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	u.lastLoginAt = lastLoginAt
	return nil
}

// SetBlockedReason sets the blocked reason (used for loading from persistence)
func (u *User) SetBlockedReason(blockedReason string) error {
	if u == nil {
		return errors.New("user cannot be nil")
	}
	if len(blockedReason) > 500 {
		return errors.New("blocked reason cannot exceed 500 characters")
	}
	u.blockedReason = blockedReason
	return nil
}
