package dto

// RoleDetailsDTO represents detailed information about a role
type RoleDetailsDTO struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	UserCount   int      `json:"user_count"`
	Users       []string `json:"users"`
	Permissions []string `json:"permissions"`
}

// RolePermissionDTO represents a permission assigned to a role
type RolePermissionDTO struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}
