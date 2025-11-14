package middleware

import (
	"strings"

	"github.com/aruncs31s/azf/constants"
	"github.com/aruncs31s/azf/initializer"
	initializers "github.com/aruncs31s/azf/initializer"
	"github.com/aruncs31s/azf/shared/helper"
	"github.com/aruncs31s/azf/shared/logger"
	"github.com/aruncs31s/azf/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var responseHelper = helper.NewResponseHelper()
var Exists bool = true

func getUserRole(c *gin.Context) (string, bool) {
	value, exists := c.Get("user_role")
	if exists {
		return value.(string), Exists
	}
	return constants.USER, false

}

// CasbinMiddleware is the authorization middleware that checks permissions using Casbin
// Be sure the place this CasbinMiddleware after the  JwtMiddleware() , otherwice the c will not contain any roles , and all the request will get cancelled

func CasbinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract user role from JWT token (assumes it's stored in context by auth middleware)
		userRole, exists := getUserRole(c)
		if !exists {
			logger.Warn("User role not found in context", zap.String("path", c.Request.URL.Path))
			responseHelper.Unauthorized(c, utils.ErrNoUserRoleInJWT.Error())
			c.Abort()
			return
		}

		// Get the request path eg: /api/v1/staff/profile
		resource := c.Request.URL.Path
		// Action , eg GET , POST etc
		action := c.Request.Method

		// Check if user has permission using Casbin
		allowed := initializer.CheckPermission(userRole, resource, action)

		if !allowed {
			logger.Warn(
				"Access denied",
				zap.String("role", userRole),
				zap.String("resource", resource),
				zap.String("action", action),
			)
			responseHelper.Forbidden(c, utils.ErrNoPermission.Error())
			c.Abort()
			return
		}

		logger.Debug(
			"Access granted",
			zap.String("role", userRole),
			zap.String("resource", resource),
			zap.String("action", action),
		)

		c.Next()
	}
}

// CasbinMiddlewareWithPathParams is an enhanced version that handles path parameters
// It normalizes paths with parameters to match policy patterns
func CasbinMiddlewareWithPathParams() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract user role from context
		userRole, exists := getUserRole(c)
		if !exists {
			logger.Warn("User role not found in context", zap.String("path", c.Request.URL.Path))
			responseHelper.Unauthorized(c, utils.ErrNoUserRoleInJWT.Error())
			c.Abort()
			return
		}

		resource := normalizeResourcePath(c.Request.URL.Path)
		action := c.Request.Method

		// Check if user has permission using Casbin
		allowed := initializers.CheckPermission(userRole, resource, action)
		if !allowed {
			logger.Warn(
				"Access denied",
				zap.String("role", userRole),
				zap.String("resource", resource),
				zap.String("action", action),
			)
			responseHelper.Forbidden(c, utils.ErrNoPermission.Error())
			c.Abort()
			return
		}

		c.Next()
	}
}

// normalizeResourcePath converts actual paths with IDs to policy patterns
// Example: /api/v1/staff/qualification/123 -> /api/v1/staff/qualification/:id
// Supports numeric IDs, UUIDs, and alphanumeric slugs
// Works automatically for ALL parameterized routes without manual pattern mapping
func normalizeResourcePath(path string) string {
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")

	// Smart: Replace ANY ID format (numeric, UUID, alphanumeric) with generic :id placeholder
	// This works for all resources automatically - no pattern map needed!
	for i := 0; i < len(parts); i++ {
		if isID(parts[i]) {
			parts[i] = ":id"
		}
	}

	return strings.Join(parts, "/")
}

// isID checks if a string represents an ID (numeric, UUID, alphanumeric slug, etc.)
func isID(s string) bool {
	if s == "" || s == ":" {
		return false
	}

	// Check if numeric (common for sequential IDs)
	if isNumeric(s) {
		return true
	}

	// Check if UUID format (e.g., 550e8400-e29b-41d4-a716-446655440000)
	_, err := uuid.Parse(s)
	if err == nil {
		return true
	}

	// Check if alphanumeric with hyphens/underscores (common for slugs and custom IDs)
	// Allow: a-z, A-Z, 0-9, hyphen, underscore
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '_') {
			return false
		}
	}

	// If all characters are alphanumeric/allowed, likely an ID
	// Require minimum length of 2 to avoid matching single letters
	return len(s) > 2
}

// isNumeric checks if a string is numeric (ID)
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
