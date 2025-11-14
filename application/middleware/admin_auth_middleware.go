package middleware

import (
	"fmt"
	"net/http"

	"github.com/aruncs31s/azf/utils"
	"github.com/gin-gonic/gin"
)

// AdminSessionMiddleware checks for admin session cookie and injects JWT token into Authorization header
// This must be used BEFORE JwtMiddleware to ensure JWT is available for further processing
func AdminSessionMiddleware(c *gin.Context) {
	// If Authorization header is already present, skip
	if c.GetHeader("Authorization") != "" {
		c.Next()
		return
	}
	// Check for admin session cookie
	_, err := c.Cookie("admin_session")
	if err != nil {
		// Redirect to login or return 401
		if c.GetHeader("Accept") == "text/html" {
			c.Redirect(http.StatusFound, "/admin-ui/login")
			c.Abort()
			return
		}
		responseHelper.Unauthorized(c, utils.ErrUnauthorized.Error())
		c.Abort()
		return
	}

	// Extract JWT token from cookie and inject into Authorization header for API calls
	jwtToken, err := c.Cookie("jwt_token")
	if err == nil && jwtToken != "" {
		c.Request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
	}

	c.Next()
}

// AdminAuthMiddleware checks if the user is authenticated as admin
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get admin_session cookie
		sessionID, err := c.Cookie("admin_session")
		if err != nil || sessionID == "" {
			c.Redirect(http.StatusFound, "/admin-ui/login")
			c.Abort()
			return
		}

		// For now, we validate the presence of the session cookie
		// In production, you would validate against a session store
		c.Set("admin_session_id", sessionID)
		c.Set("is_admin", true)

		c.Next()
	}
}

// CheckAdminAuth is a helper middleware that returns to login if not authenticated
func CheckAdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := c.Cookie("admin_session")
		if err != nil {
			c.Redirect(http.StatusFound, "/admin-ui/login")
			c.Abort()
			return
		}

		c.Next()
	}
}
