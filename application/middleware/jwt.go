package middleware

import (
	"fmt"
	"log"
	"os"

	"strings"

	"github.com/aruncs31s/azf/constants"
	"github.com/aruncs31s/azf/shared/logger"
	"go.uber.org/zap"

	"github.com/aruncs31s/azf/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GetSecretKey() []byte {
	key := os.Getenv("JWT_SECRET")
	if key == "" {
		log.Fatal("JWT_SECRET not set in environment")
	}

	// Validate JWT_SECRET meets security requirements
	if len(key) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 characters for security compliance")
	}

	if len(key) < 64 {
		logger.GetLogger().Warn(
			"JWT_SECRET is shorter than recommended",
			zap.Int("length", len(key)),
			zap.String("recommendation", "Use at least 64 characters for enhanced security"),
		)
	}

	logger.GetLogger().Debug(
		"JWT_SECRET validation passed",
		zap.Int("length", len(key)),
	)

	return []byte(key)
}

// Currently used by enterprise middleware and routes.go
func GetUserRole(c *gin.Context) string {
	role, exists := c.Get("user_role")
	if !exists {
		return ""
	}
	if roleStr, ok := role.(string); ok {
		return roleStr
	}
	return constants.USER
}
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	if userIDStr, ok := userID.(string); ok {
		return userIDStr
	}
	return ""
}

func SetCors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
func JwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			responseHelper.Unauthorized(c, utils.ErrNoAuthHeader.Error())
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse the token
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return GetSecretKey(), nil
		})

		if err != nil || !token.Valid {
			responseHelper.Unauthorized(c, utils.ErrUnauthorized.Error())
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("jwt_claims", claims)

			if claims["user_id"] != nil {
				c.Set("user_id", claims["user_id"])
			}

			// Extract role from claims for Casbin authorization
			if role, exists := claims["role"]; exists {
				c.Set("user_role", role)
			} else {
				// Default to staff role if not specified
				c.Set("user_role", "user")
			}
		}

		c.Next()
	}
}
