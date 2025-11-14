package user_management

import "errors"

// UserRole represents a role assigned to a user as a value object
type UserRole struct {
	name        string
	permissions []string
}

// NewUserRole creates a new UserRole value object with validation
func NewUserRole(name string, permissions []string) (*UserRole, error) {
	if name == "" {
		return nil, errors.New("role name cannot be empty")
	}
	if len(name) > 50 {
		return nil, errors.New("role name cannot exceed 50 characters")
	}

	// Validate permissions
	if len(permissions) > 100 {
		return nil, errors.New("cannot assign more than 100 permissions")
	}

	// Create copy to prevent external mutation
	permsCopy := make([]string, len(permissions))
	copy(permsCopy, permissions)

	return &UserRole{
		name:        name,
		permissions: permsCopy,
	}, nil
}

// Name returns the role name
func (r *UserRole) Name() string {
	if r == nil {
		return ""
	}
	return r.name
}

// Permissions returns a copy of the permissions list
func (r *UserRole) Permissions() []string {
	if r == nil {
		return []string{}
	}
	permsCopy := make([]string, len(r.permissions))
	copy(permsCopy, r.permissions)
	return permsCopy
}

// HasPermission checks if the role has a specific permission
func (r *UserRole) HasPermission(permission string) bool {
	if r == nil {
		return false
	}
	for _, p := range r.permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// String returns the string representation
func (r *UserRole) String() string {
	return r.name
}

// Equals checks if two roles are equal
func (r *UserRole) Equals(other *UserRole) bool {
	if r == nil || other == nil {
		return r == other
	}
	return r.name == other.name
}

// IsAdmin checks if this is an admin role
func (r *UserRole) IsAdmin() bool {
	if r == nil {
		return false
	}
	return r.name == "admin" || r.name == "ADMIN"
}

// IsViewer checks if this is a viewer role
func (r *UserRole) IsViewer() bool {
	if r == nil {
		return false
	}
	return r.name == "viewer" || r.name == "VIEWER"
}

// IsEditor checks if this is an editor role
func (r *UserRole) IsEditor() bool {
	if r == nil {
		return false
	}
	return r.name == "editor" || r.name == "EDITOR"
}
