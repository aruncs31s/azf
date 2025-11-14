package service

import (
	"errors"
	"fmt"

	"time"

	"github.com/aruncs31s/azf/application/dto"
	"github.com/aruncs31s/azf/config"
	"github.com/aruncs31s/azf/utils"
)

// TODO: Make it DDD Complaint
// AdminAuthenticationService handles admin authentication operations
type AdminAuthenticationService struct {
	configProvider *config.AdminConfigProvider
}

// NewAdminAuthenticationService creates a new instance of AdminAuthenticationService
func NewAdminAuthenticationService(configProvider *config.AdminConfigProvider) *AdminAuthenticationService {
	return &AdminAuthenticationService{
		configProvider: configProvider,
	}
}

// Login authenticates an admin user with username and password
// Following DDD: this service uses domain aggregates and value objects for validation
func (s *AdminAuthenticationService) Login(request *dto.LoginRequest) (*dto.AdminLoginResponse, error) {
	if request == nil {
		return nil, utils.ErrInvalidData
	}

	// Get admin credentials from configuration
	expectedCredentials, err := s.configProvider.GetAdminCredentials()
	if err != nil {
		return &dto.AdminLoginResponse{
			Success:   false,
			Message:   "Admin credentials not configured",
			Error:     "server configuration error",
			Timestamp: time.Now().Format(time.RFC3339),
		}, fmt.Errorf("failed to get admin credentials: %w", err)
	}

	// Validate credentials through the domain aggregate
	// This ensures all business rules are applied
	err = expectedCredentials.Authenticate(request.Username, request.Password)
	if err != nil {
		return &dto.AdminLoginResponse{
			Success:   false,
			Message:   "Authentication failed",
			Error:     utils.ErrInvalidUsernameOrPassword.Error(),
			Timestamp: time.Now().Format(time.RFC3339),
		}, nil
	}

	// Successful authentication - generate session ID
	sessionID := s.generateSessionID()
	return &dto.AdminLoginResponse{
		Success:   true,
		Message:   "Login successful",
		SessionID: sessionID,
		Admin: dto.AdminInfo{
			ID:       expectedCredentials.ID(),
			Username: expectedCredentials.Username().Value(),
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// generateSessionID creates a unique session identifier
func (s *AdminAuthenticationService) generateSessionID() string {
	return fmt.Sprintf("admin_session_%d", time.Now().UnixNano())
}

// ValidateSession validates an admin session
// In a production system, this would validate against a session store
func (s *AdminAuthenticationService) ValidateSession(sessionID string) (bool, error) {
	if sessionID == "" {
		return false, errors.New("session ID cannot be empty")
	}

	// TODO: Implement session validation with session storage
	// For now, return true as sessions are handled by the handler
	return true, nil
}

// Logout handles admin logout by invalidating the session
func (s *AdminAuthenticationService) Logout(sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID cannot be empty")
	}

	// TODO: Implement session cleanup and invalidation
	return nil
}
