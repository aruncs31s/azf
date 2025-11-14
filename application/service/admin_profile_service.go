package service

import (
	"fmt"
	"sort"
	"time"

	"github.com/aruncs31s/azf/config"
	"github.com/aruncs31s/azf/domain/model"
	"github.com/aruncs31s/azf/initializer"
)

// AdminProfileService handles admin profile operations
// Following DDD application service pattern
type AdminProfileService struct {
	configProvider *config.AdminConfigProvider
}

// customDescriptions stores role descriptions that persist across requests
var customDescriptions = make(map[string]string)

// NewAdminProfileService creates a new AdminProfileService
func NewAdminProfileService(configProvider *config.AdminConfigProvider) *AdminProfileService {
	return &AdminProfileService{
		configProvider: configProvider,
	}
}

// GetAdminProfile retrieves the admin profile information
// In a full DDD implementation, this would use a repository to fetch from database
// For now, it builds a profile from configuration data
func (s *AdminProfileService) GetAdminProfile() (*model.AdminProfile, error) {
	// Get admin credentials from configuration
	credentials, err := s.configProvider.GetAdminCredentials()
	if err != nil {
		return nil, err
	}

	// Create value objects for profile
	fullName, err := model.NewAdminFullName("System Administrator")
	if err != nil {
		return nil, err
	}

	email, err := model.NewAdminEmail("admin@AZF.dev")
	if err != nil {
		return nil, err
	}

	// Create admin profile entity
	profile, err := model.NewAdminProfile(
		credentials.ID(),
		fullName,
		email,
		"Super Administrator",
	)
	if err != nil {
		return nil, err
	}

	// Record a login (simulating current session)
	profile.RecordLogin()

	return profile, nil
}

// UpdateAdminProfile updates the admin profile information
// This is a placeholder for future implementation with repository
func (s *AdminProfileService) UpdateAdminProfile(profileID string, updates map[string]interface{}) error {
	// TODO: Implement profile updates with repository
	// For now, this is a no-op as we don't have persistence
	return nil
}

// GetProfileSummary returns a summary of the admin profile for display
func (s *AdminProfileService) GetProfileSummary() (map[string]interface{}, error) {
	profile, err := s.GetAdminProfile()
	if err != nil {
		return nil, err
	}

	summary := profile.ProfileSummary()

	// Add additional computed fields for UI
	summary["display_name"] = profile.DisplayName()
	summary["role_display"] = profile.Role()
	summary["status"] = "Active"
	if !profile.IsActive() {
		summary["status"] = "Inactive"
	}

	// Add login information with defaults
	if profile.LastLogin() != nil {
		summary["last_login_display"] = profile.LastLogin().Format("Jan 2, 2006 at 3:04 PM")
		summary["days_since_login"] = int(time.Since(*profile.LastLogin()).Hours() / 24)
	} else {
		summary["last_login_display"] = "Never"
		summary["days_since_login"] = 0
	}

	// Ensure all required keys exist with defaults
	defaults := map[string]interface{}{
		"id":                 "admin_1",
		"full_name":          "System Administrator",
		"email":              "admin@AZF.dev",
		"role":               "Super Administrator",
		"display_name":       "System Administrator",
		"role_display":       "Super Administrator",
		"status":             "Active",
		"last_login_display": "Never",
		"days_since_login":   0,
		"is_active":          true,
	}

	// Apply defaults for any missing keys
	for key, defaultValue := range defaults {
		if _, exists := summary[key]; !exists {
			summary[key] = defaultValue
		}
	}

	return summary, nil
}

// GetAllRolesFromCasbin extracts all unique roles from the Casbin enforcer
// This includes roles from policies, grouping rules, and custom descriptions
func (s *AdminProfileService) GetAllRolesFromCasbin() ([]string, error) {
	enforcer := initializer.CasbinEnforcer
	if enforcer == nil {
		// Fallback to hardcoded roles if enforcer not available
		return []string{"admin", "staff", "user", "student", "parent", "moderator", "director", "principal", "dean", "hod", "teacher", "coe"}, nil
	}

	roleSet := make(map[string]bool)

	// Get all policies and extract roles (first column after 'p')
	policies, err := enforcer.GetPolicy()
	if err != nil {
		return nil, err
	}
	for _, policy := range policies {
		if len(policy) > 0 {
			roleSet[policy[0]] = true
		}
	}

	// Get all grouping policies and extract roles (second column after 'g')
	groupingPolicies, err := enforcer.GetGroupingPolicy()
	if err != nil {
		return nil, err
	}
	for _, groupPolicy := range groupingPolicies {
		if len(groupPolicy) > 1 {
			roleSet[groupPolicy[1]] = true
		}
	}

	// Include roles that have custom descriptions but haven't been used in policies yet
	for role := range customDescriptions {
		roleSet[role] = true
	}

	// Convert map to sorted slice
	var roles []string
	for role := range roleSet {
		roles = append(roles, role)
	}
	sort.Strings(roles)

	return roles, nil
}

// GetRoleDescriptions returns descriptions for known roles
func (s *AdminProfileService) GetRoleDescriptions() map[string]string {
	// Start with default descriptions
	descriptions := map[string]string{
		"admin":     "System Administrator with full access",
		"staff":     "Staff member with standard access",
		"user":      "Basic user with minimal permissions",
		"student":   "Student user account",
		"parent":    "Parent user account",
		"moderator": "Content moderation capabilities",
		"director":  "Top-level director role",
		"principal": "Principal role",
		"dean":      "Dean role",
		"hod":       "Head of Department",
		"teacher":   "Teaching staff",
		"coe":       "Controller of Examinations",
	}

	// Override with custom descriptions
	for role, desc := range customDescriptions {
		descriptions[role] = desc
	}

	return descriptions
}

// CreateRole creates a new role with the given name and description
func (s *AdminProfileService) CreateRole(name string, description string) error {
	enforcer := initializer.CasbinEnforcer
	if enforcer == nil {
		return fmt.Errorf("casbin enforcer not available")
	}

	// Check if role already exists
	roles, err := s.GetAllRolesFromCasbin()
	if err != nil {
		return err
	}

	for _, existingRole := range roles {
		if existingRole == name {
			return fmt.Errorf("role '%s' already exists", name)
		}
	}

	// Store the description
	customDescriptions[name] = description

	// Note: In Casbin, roles are typically created implicitly when used in policies
	// For now, we just store the description. Actual role creation happens when policies are added.

	return nil
}

// UpdateRole updates an existing role's name and/or description
func (s *AdminProfileService) UpdateRole(oldName string, newName string, description string) error {
	enforcer := initializer.CasbinEnforcer
	if enforcer == nil {
		return fmt.Errorf("casbin enforcer not available")
	}

	// If name is changing, we need to update all references
	if oldName != newName && newName != "" {
		// Update all policies that reference the old role name
		policies, err := enforcer.GetPolicy()
		if err != nil {
			return fmt.Errorf("failed to get policies: %w", err)
		}

		for _, policy := range policies {
			if len(policy) > 0 && policy[0] == oldName {
				// Remove old policy
				_, err := enforcer.RemovePolicy(policy)
				if err != nil {
					return fmt.Errorf("failed to remove old policy: %w", err)
				}
				// Add new policy with updated role name
				policy[0] = newName
				_, err = enforcer.AddPolicy(policy)
				if err != nil {
					return fmt.Errorf("failed to add updated policy: %w", err)
				}
			}
		}

		// Update all grouping policies that reference the old role name
		groupingPolicies, err := enforcer.GetGroupingPolicy()
		if err != nil {
			return fmt.Errorf("failed to get grouping policies: %w", err)
		}

		for _, gPolicy := range groupingPolicies {
			if len(gPolicy) > 1 && gPolicy[1] == oldName {
				// Remove old grouping policy
				_, err := enforcer.RemoveGroupingPolicy(gPolicy)
				if err != nil {
					return fmt.Errorf("failed to remove old grouping policy: %w", err)
				}
				// Add new grouping policy with updated role name
				gPolicy[1] = newName
				_, err = enforcer.AddGroupingPolicy(gPolicy)
				if err != nil {
					return fmt.Errorf("failed to add updated grouping policy: %w", err)
				}
			}
		}

		// Update custom description for the new name
		if description != "" {
			customDescriptions[newName] = description
		}
		delete(customDescriptions, oldName)
	} else {
		// Only updating description
		if description != "" {
			customDescriptions[oldName] = description
		}
	}

	return nil
}

// AssignRoleToUser assigns a role to a user in Casbin
func (s *AdminProfileService) AssignRoleToUser(userID string, role string) error {
	enforcer := initializer.CasbinEnforcer
	if enforcer == nil {
		return fmt.Errorf("casbin enforcer not available")
	}

	// Add grouping policy: user -> role
	added, err := enforcer.AddGroupingPolicy(userID, role)
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	if !added {
		return fmt.Errorf("role assignment already exists")
	}

	return nil
}

// RemoveRoleFromUser removes a role from a user in Casbin
func (s *AdminProfileService) RemoveRoleFromUser(userID string, role string) error {
	enforcer := initializer.CasbinEnforcer
	if enforcer == nil {
		return fmt.Errorf("casbin enforcer not available")
	}

	// Remove grouping policy: user -> role
	removed, err := enforcer.RemoveGroupingPolicy(userID, role)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	if !removed {
		return fmt.Errorf("role assignment does not exist")
	}

	return nil
}

// GetUsersForRole returns all users assigned to a specific role
func (s *AdminProfileService) GetUsersForRole(role string) ([]string, error) {
	enforcer := initializer.CasbinEnforcer
	if enforcer == nil {
		return nil, fmt.Errorf("casbin enforcer not available")
	}

	// Get all users with this role
	users, err := enforcer.GetUsersForRole(role)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for role: %w", err)
	}

	return users, nil
}

// DeleteRole removes a role and all its assignments
func (s *AdminProfileService) DeleteRole(role string) error {
	enforcer := initializer.CasbinEnforcer
	if enforcer == nil {
		return fmt.Errorf("casbin enforcer not available")
	}

	// Remove all grouping policies for this role
	_, err := enforcer.RemoveFilteredGroupingPolicy(1, role)
	if err != nil {
		return fmt.Errorf("failed to remove role assignments: %w", err)
	}

	// Remove all policies that use this role
	_, err = enforcer.RemoveFilteredPolicy(0, role)
	if err != nil {
		return fmt.Errorf("failed to remove role policies: %w", err)
	}

	// Remove custom description
	delete(customDescriptions, role)

	// Note: We don't return an error if nothing was removed, as the role might not have been used

	return nil
}

// GetUserRoles returns all roles assigned to a user
func (s *AdminProfileService) GetUserRoles(userID string) ([]string, error) {
	enforcer := initializer.CasbinEnforcer
	if enforcer == nil {
		return nil, fmt.Errorf("casbin enforcer not available")
	}

	// Get all roles for this user
	roles, err := enforcer.GetRolesForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for user: %w", err)
	}

	return roles, nil
}

// GetRoleDetails returns detailed information about a specific role
func (s *AdminProfileService) GetRoleDetails(roleName string) (map[string]interface{}, error) {
	enforcer := initializer.CasbinEnforcer
	if enforcer == nil {
		return nil, fmt.Errorf("casbin enforcer not available")
	}

	// Get users for this role
	users, err := enforcer.GetUsersForRole(roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for role: %w", err)
	}

	// Get permissions for this role
	permissions := make([]map[string]string, 0)
	policies, err := enforcer.GetFilteredPolicy(0, roleName)
	if err == nil {
		for _, policy := range policies {
			if len(policy) >= 3 {
				permissions = append(permissions, map[string]string{
					"resource": policy[1],
					"action":   policy[2],
				})
			}
		}
	}

	// Get description
	descriptions := s.GetRoleDescriptions()
	description := descriptions[roleName]
	if description == "" {
		description = roleName + " role"
	}

	return map[string]interface{}{
		"name":        roleName,
		"description": description,
		"user_count":  len(users),
		"users":       users,
		"permissions": permissions,
	}, nil
}

// GetRolePermissions returns all permissions for a specific role
func (s *AdminProfileService) GetRolePermissions(roleName string) ([]map[string]string, error) {
	enforcer := initializer.CasbinEnforcer
	if enforcer == nil {
		return nil, fmt.Errorf("casbin enforcer not available")
	}

	permissions := make([]map[string]string, 0)
	policies, err := enforcer.GetFilteredPolicy(0, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	for _, policy := range policies {
		if len(policy) >= 3 {
			permissions = append(permissions, map[string]string{
				"resource": policy[1],
				"action":   policy[2],
			})
		}
	}

	return permissions, nil
}
