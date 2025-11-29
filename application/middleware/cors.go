package middleware

import (
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	// AllowedOrigins is a list of origins that are allowed to make requests
	AllowedOrigins []string
	// AllowedMethods is a list of methods that are allowed
	AllowedMethods []string
	// AllowedHeaders is a list of headers that are allowed
	AllowedHeaders []string
	// ExposedHeaders is a list of headers that are exposed to the browser
	ExposedHeaders []string
	// AllowCredentials indicates whether credentials are allowed
	AllowCredentials bool
	// MaxAge is the max age for preflight cache in seconds
	MaxAge int
}

// DefaultCORSConfig returns a secure default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{}, // Must be explicitly set
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"Accept",
			"Cache-Control",
			"X-Requested-With",
			"X-Request-ID",
		},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// DevelopmentCORSConfig returns a permissive CORS config for development
func DevelopmentCORSConfig() *CORSConfig {
	cfg := DefaultCORSConfig()
	cfg.AllowedOrigins = []string{"*"}
	return cfg
}

// CORSMiddleware creates a CORS middleware with the given config
func CORSMiddleware(config *CORSConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig()
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if isOriginAllowed(origin, config.AllowedOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(config.AllowedOrigins) == 1 && config.AllowedOrigins[0] == "*" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// Set other CORS headers
		if config.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))

		if len(config.ExposedHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			if config.MaxAge > 0 {
				c.Writer.Header().Set("Access-Control-Max-Age", string(rune(config.MaxAge)))
			}
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isOriginAllowed checks if an origin is in the allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return false
	}

	// Check for wildcard
	if slices.Contains(allowedOrigins, "*") {
		return true
	}

	// Check for exact match
	return slices.Contains(allowedOrigins, origin)
}

// SecureCORSMiddleware creates a CORS middleware that only allows specified origins
// This should be used in production
func SecureCORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	config := DefaultCORSConfig()
	config.AllowedOrigins = allowedOrigins
	return CORSMiddleware(config)
}
