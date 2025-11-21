package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aruncs31s/azf/application/dto"
	usermodel "github.com/aruncs31s/azf/domain/user_management/model"
	"github.com/aruncs31s/azf/shared/logger"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// OAuthProvider represents supported OAuth providers
type OAuthProvider string

const (
	Google OAuthProvider = "google"
	GitHub OAuthProvider = "github"
)

// OAuthService handles OAuth authentication operations
type OAuthService struct {
	userRepo     usermodel.UserRepository
	oauthConfigs map[OAuthProvider]*oauth2.Config
	baseURL      string
	jwtSecret    string
}

// OAuthUserInfo represents user information from OAuth provider
type OAuthUserInfo struct {
	ID            string
	Email         string
	Name          string
	Username      string
	AvatarURL     string
	VerifiedEmail bool
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(
	userRepo usermodel.UserRepository,
	baseURL string,
	jwtSecret string,
) *OAuthService {
	service := &OAuthService{
		userRepo:     userRepo,
		oauthConfigs: make(map[OAuthProvider]*oauth2.Config),
		baseURL:      baseURL,
		jwtSecret:    jwtSecret,
	}

	// Initialize OAuth configs
	service.initOAuthConfigs()

	return service
}

// initOAuthConfigs initializes OAuth2 configurations for supported providers
func (s *OAuthService) initOAuthConfigs() {
	// Google OAuth
	if clientID := s.getEnvVar("GOOGLE_OAUTH_CLIENT_ID"); clientID != "" {
		if clientSecret := s.getEnvVar("GOOGLE_OAUTH_CLIENT_SECRET"); clientSecret != "" {
			s.oauthConfigs[Google] = &oauth2.Config{
				ClientID:     clientID,
				ClientSecret: clientSecret,
				RedirectURL:  fmt.Sprintf("%s/admin-ui/oauth/callback/google", s.baseURL),
				Scopes:       []string{"openid", "profile", "email"},
				Endpoint:     google.Endpoint,
			}
		}
	}

	// GitHub OAuth
	if clientID := s.getEnvVar("GITHUB_OAUTH_CLIENT_ID"); clientID != "" {
		if clientSecret := s.getEnvVar("GITHUB_OAUTH_CLIENT_SECRET"); clientSecret != "" {
			s.oauthConfigs[GitHub] = &oauth2.Config{
				ClientID:     clientID,
				ClientSecret: clientSecret,
				RedirectURL:  fmt.Sprintf("%s/admin-ui/oauth/callback/github", s.baseURL),
				Scopes:       []string{"user:email", "read:user"},
				Endpoint:     github.Endpoint,
			}
		}
	}
}

// getEnvVar retrieves environment variable
func (s *OAuthService) getEnvVar(key string) string {
	return os.Getenv(key)
}

// GetAuthURL generates OAuth authorization URL for the specified provider
func (s *OAuthService) GetAuthURL(provider OAuthProvider, state string) (string, error) {
	config, exists := s.oauthConfigs[provider]
	if !exists {
		return "", fmt.Errorf("OAuth provider %s not configured", provider)
	}

	if state == "" {
		state = s.generateState()
	}

	return config.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

// HandleCallback processes OAuth callback and returns user authentication result
func (s *OAuthService) HandleCallback(provider OAuthProvider, code, state string) (*dto.AdminLoginResponse, error) {
	config, exists := s.oauthConfigs[provider]
	if !exists {
		return nil, fmt.Errorf("OAuth provider %s not configured", provider)
	}

	// Exchange code for token
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		logger.GetLogger().Error("OAuth token exchange failed",
			zap.String("provider", string(provider)),
			zap.Error(err))
		return nil, fmt.Errorf("failed to exchange OAuth code: %w", err)
	}

	// Get user info from provider
	userInfo, err := s.getUserInfo(provider, token)
	if err != nil {
		logger.GetLogger().Error("Failed to get OAuth user info",
			zap.String("provider", string(provider)),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Find or create user
	user, err := s.findOrCreateUser(provider, userInfo)
	if err != nil {
		logger.GetLogger().Error("Failed to find or create OAuth user",
			zap.String("provider", string(provider)),
			zap.String("oauthID", userInfo.ID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to process user: %w", err)
	}

	// Record login
	if err := user.RecordLogin(); err != nil {
		logger.GetLogger().Warn("Failed to record login",
			zap.String("userID", user.GetID()),
			zap.Error(err))
	}

	// Save user (Create or Update based on whether it's new)
	ctx := context.Background()
	if strings.HasPrefix(user.GetID(), "user_") {
		// New user
		if _, err := s.userRepo.Create(ctx, user); err != nil {
			logger.GetLogger().Error("Failed to create user after OAuth login",
				zap.String("userID", user.GetID()),
				zap.Error(err))
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Existing user
		if _, err := s.userRepo.Update(ctx, user); err != nil {
			logger.GetLogger().Error("Failed to update user after OAuth login",
				zap.String("userID", user.GetID()),
				zap.Error(err))
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Generate JWT
	jwtToken, err := s.generateJWT(user)
	if err != nil {
		logger.GetLogger().Error("Failed to generate JWT for OAuth user",
			zap.String("userID", user.GetID()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return &dto.AdminLoginResponse{
		Success:   true,
		Message:   "OAuth login successful",
		SessionID: s.generateSessionID(),
		Admin: dto.AdminInfo{
			ID:       user.GetID(),
			Username: user.GetUsername(),
		},
		JWT:       jwtToken,
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}

// getUserInfo retrieves user information from OAuth provider
func (s *OAuthService) getUserInfo(provider OAuthProvider, token *oauth2.Token) (*OAuthUserInfo, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", s.getUserInfoURL(provider), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OAuth provider returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return s.parseUserInfo(provider, body)
}

// getUserInfoURL returns the user info endpoint for the provider
func (s *OAuthService) getUserInfoURL(provider OAuthProvider) string {
	switch provider {
	case Google:
		return "https://www.googleapis.com/oauth2/v2/userinfo"
	case GitHub:
		return "https://api.github.com/user"
	default:
		return ""
	}
}

// parseUserInfo parses user information from provider response
func (s *OAuthService) parseUserInfo(provider OAuthProvider, data []byte) (*OAuthUserInfo, error) {
	var userInfo OAuthUserInfo

	switch provider {
	case Google:
		var googleUser struct {
			ID            string `json:"id"`
			Email         string `json:"email"`
			Name          string `json:"name"`
			VerifiedEmail bool   `json:"verified_email"`
			Picture       string `json:"picture"`
		}
		if err := json.Unmarshal(data, &googleUser); err != nil {
			return nil, err
		}
		userInfo = OAuthUserInfo{
			ID:            googleUser.ID,
			Email:         googleUser.Email,
			Name:          googleUser.Name,
			Username:      strings.Split(googleUser.Email, "@")[0], // Use email prefix as username
			AvatarURL:     googleUser.Picture,
			VerifiedEmail: googleUser.VerifiedEmail,
		}

	case GitHub:
		var githubUser struct {
			ID    int    `json:"id"`
			Login string `json:"login"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := json.Unmarshal(data, &githubUser); err != nil {
			return nil, err
		}

		// GitHub may not return email in user endpoint, handle separately if needed
		userInfo = OAuthUserInfo{
			ID:        fmt.Sprintf("%d", githubUser.ID),
			Email:     githubUser.Email,
			Name:      githubUser.Name,
			Username:  githubUser.Login,
			AvatarURL: "",
		}

	default:
		return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}

	return &userInfo, nil
}

// findOrCreateUser finds existing user or creates new one
func (s *OAuthService) findOrCreateUser(provider OAuthProvider, userInfo *OAuthUserInfo) (*usermodel.User, error) {
	ctx := context.Background()

	// Try to find existing user by OAuth ID
	existingUser, err := s.userRepo.GetByOAuthID(ctx, string(provider), userInfo.ID)
	if err == nil && existingUser != nil {
		return existingUser, nil
	}

	// Try to find by email
	existingUser, err = s.userRepo.GetByEmail(ctx, userInfo.Email)
	if err == nil && existingUser != nil {
		// Link OAuth account to existing user
		if err := existingUser.SetOAuthProvider(string(provider)); err != nil {
			return nil, err
		}
		if err := existingUser.SetOAuthID(userInfo.ID); err != nil {
			return nil, err
		}
		return existingUser, nil
	}

	// Create new user
	user, err := usermodel.NewUser(
		s.generateUserID(),
		userInfo.Email,
		userInfo.Username,
		userInfo.Name,
	)
	if err != nil {
		return nil, err
	}

	if err := user.SetOAuthProvider(string(provider)); err != nil {
		return nil, err
	}
	if err := user.SetOAuthID(userInfo.ID); err != nil {
		return nil, err
	}

	return user, nil
}

// generateUserID generates a unique user ID
func (s *OAuthService) generateUserID() string {
	// TODO: Use UUID library
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}

// generateState generates a random state string for OAuth
func (s *OAuthService) generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// generateSessionID generates a unique session ID
func (s *OAuthService) generateSessionID() string {
	return fmt.Sprintf("oauth_session_%d", time.Now().UnixNano())
}

// generateJWT generates JWT token for the user
func (s *OAuthService) generateJWT(user *usermodel.User) (string, error) {
	// Create JWT claims
	claims := jwt.MapClaims{
		"user_id":  user.GetID(),
		"username": user.GetUsername(),
		"email":    user.GetEmail(),
		"role":     "user",                                // Default role for OAuth users
		"exp":      time.Now().Add(24 * time.Hour).Unix(), // 24 hours
		"iat":      time.Now().Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return tokenString, nil
}

// IsProviderConfigured checks if OAuth provider is configured
func (s *OAuthService) IsProviderConfigured(provider OAuthProvider) bool {
	_, exists := s.oauthConfigs[provider]
	return exists
}
