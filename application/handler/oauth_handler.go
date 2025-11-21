package handler

import (
	"net/http"

	"github.com/aruncs31s/azf/application/service"
	"github.com/aruncs31s/azf/shared/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OAuthHandler handles OAuth authentication endpoints
type OAuthHandler struct {
	oauthService *service.OAuthService
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(oauthService *service.OAuthService) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
	}
}

// Login initiates OAuth login flow
func (h *OAuthHandler) Login(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider is required"})
		return
	}

	var oauthProvider service.OAuthProvider
	switch provider {
	case "google":
		oauthProvider = service.Google
	case "github":
		oauthProvider = service.GitHub
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported OAuth provider"})
		return
	}

	if !h.oauthService.IsProviderConfigured(oauthProvider) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth provider not configured"})
		return
	}

	state := c.Query("state")
	if state == "" {
		state = "random_state" // In production, generate secure random state
	}

	authURL, err := h.oauthService.GetAuthURL(oauthProvider, state)
	if err != nil {
		logger.GetLogger().Error("Failed to generate OAuth URL",
			zap.String("provider", provider),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate OAuth login"})
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

// Callback handles OAuth provider callback
func (h *OAuthHandler) Callback(c *gin.Context) {
	provider := c.Param("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider is required"})
		return
	}

	var oauthProvider service.OAuthProvider
	switch provider {
	case "google":
		oauthProvider = service.Google
	case "github":
		oauthProvider = service.GitHub
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported OAuth provider"})
		return
	}

	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is required"})
		return
	}

	// Handle the OAuth callback
	response, err := h.oauthService.HandleCallback(oauthProvider, code, state)
	if err != nil {
		logger.GetLogger().Error("OAuth callback failed",
			zap.String("provider", provider),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OAuth authentication failed"})
		return
	}

	if !response.Success {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": response.Message,
			"error":   response.Error,
		})
		return
	}

	// Set JWT token in response
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   response.Message,
		"jwt":       response.JWT,
		"admin":     response.Admin,
		"timestamp": response.Timestamp,
	})
}

// GetProviders returns list of configured OAuth providers
func (h *OAuthHandler) GetProviders(c *gin.Context) {
	providers := []string{}

	if h.oauthService.IsProviderConfigured(service.Google) {
		providers = append(providers, "google")
	}
	if h.oauthService.IsProviderConfigured(service.GitHub) {
		providers = append(providers, "github")
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
	})
}
