package middleware

import (
	"time"

	"github.com/aruncs31s/azf/shared/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID is already present in header
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		c.Next()
	}
}

// StructuredLoggingMiddleware adds structured logging with request context
func StructuredLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get or generate request ID
		requestID, _ := c.Get("request_id")
		requestIDStr, _ := requestID.(string)
		if requestIDStr == "" {
			requestIDStr = uuid.New().String()
			c.Set("request_id", requestIDStr)
		}

		// Get user ID if available
		userID := ""
		if id, exists := c.Get("user_id"); exists {
			if idStr, ok := id.(string); ok {
				userID = idStr
			}
		}

		// Create request-scoped logger
		reqLogger := logger.NewRequestLogger(
			requestIDStr,
			userID,
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
		)

		// Store logger in context
		c.Set("logger", reqLogger)

		// Log request start
		reqLogger.Debug("request started",
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int64("content_length", c.Request.ContentLength),
		)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Determine log level based on status code
		statusCode := c.Writer.Status()
		responseSize := c.Writer.Size()

		fields := []zap.Field{
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.Int("response_size", responseSize),
		}

		// Add error info if present
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// Log based on status code
		switch {
		case statusCode >= 500:
			reqLogger.Error("request completed with server error", fields...)
		case statusCode >= 400:
			reqLogger.Warn("request completed with client error", fields...)
		default:
			reqLogger.Info("request completed", fields...)
		}
	}
}

// GetRequestLogger retrieves the request-scoped logger from context
func GetRequestLogger(c *gin.Context) *zap.Logger {
	if reqLogger, exists := c.Get("logger"); exists {
		if l, ok := reqLogger.(*zap.Logger); ok {
			return l
		}
	}
	return logger.GetLogger()
}

// GetRequestID retrieves the request ID from context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
