package config

import (
	"fmt"
	"os"

	"github.com/aruncs31s/azf/domain/model"
)

// AdminConfig represents the admin configuration following DDD principles
// This is a Value Object that encapsulates admin-related configuration
type AdminConfig struct {
	username *model.AdminUsername
	password *model.AdminPassword
}

var adminUsernameVar string = "ADMIN_USERNAME"
var adminPasswordVar string = "ADMIN_PASSWORD"

// GetAdminConfig retrieves and validates admin credentials from environment variables
// Following DDD approach: configuration is loaded from environment and validated through value objects
func GetAdminConfig() (*AdminConfig, error) {
	usernameStr := os.Getenv(adminUsernameVar)
	passwordStr := os.Getenv(adminPasswordVar)

	// Validate username
	if usernameStr == "" {
		return nil, fmt.Errorf("environment variable %s not set", adminUsernameVar)
	}

	// Validate password
	if passwordStr == "" {
		return nil, fmt.Errorf("environment variable %s not set", adminPasswordVar)
	}

	// Create value objects through their constructors
	// This ensures business rules are applied at the value object level
	username, err := model.NewAdminUsername(usernameStr)
	if err != nil {
		return nil, fmt.Errorf("environment variable %s not valid", adminUsernameVar)
	}

	password, err := model.NewAdminPassword(passwordStr)
	if err != nil {
		return nil, fmt.Errorf("environment variable %s not valid", adminPasswordVar)
	}

	return &AdminConfig{
		username: username,
		password: password,
	}, nil
}

// GetUsername returns the admin username value object
func (ac *AdminConfig) GetUsername() *model.AdminUsername {
	return ac.username
}

// GetPassword returns the admin password value object
func (ac *AdminConfig) GetPassword() *model.AdminPassword {
	return ac.password
}

// GetUsernameString returns the admin username as string
func (ac *AdminConfig) GetUsernameString() string {
	if ac.username == nil {
		return ""
	}
	return ac.username.Value()
}

// GetPasswordString returns the admin password as string
func (ac *AdminConfig) GetPasswordString() string {
	if ac.password == nil {
		return ""
	}
	return ac.password.Value()
}

// AdminConfigProvider provides access to admin configuration
// Following DDD: this is an application service that provides domain configuration
type AdminConfigProvider struct {
	config *AdminConfig
}

// NewAdminConfigProvider creates a new admin configuration provider
func NewAdminConfigProvider() (*AdminConfigProvider, error) {
	config, err := GetAdminConfig()
	if err != nil {
		return nil, err
	}

	return &AdminConfigProvider{
		config: config,
	}, nil
}

// GetAdminCredentials returns the admin credentials for authentication
// This method encapsulates the creation of admin credentials aggregate
func (acp *AdminConfigProvider) GetAdminCredentials() (*model.AdminCredentials, error) {
	if acp.config == nil {
		return nil, fmt.Errorf("admin configuration not initialized")
	}

	credentials, err := model.NewAdminCredentials(
		"admin_1", // Default admin ID
		acp.config.GetUsername(),
		acp.config.GetPassword(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create admin credentials: %w", err)
	}

	return credentials, nil
}
