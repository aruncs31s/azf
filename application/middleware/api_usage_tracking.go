package middleware

import (
	"bytes"
	"io"

	"time"

	"github.com/aruncs31s/azf/domain/api_usage"
	"github.com/aruncs31s/azf/domain/repository"
	"github.com/aruncs31s/azf/infrastructure/persistence"
	initializers "github.com/aruncs31s/azf/initializer"
	"github.com/aruncs31s/azf/shared/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var apiUsageRepo repository.APIUsageLogRepository

func InitAPIUsageTracking(repo repository.APIUsageLogRepository) {
	apiUsageRepo = repo
}

// APIUsageTrackingMiddleware tracks API endpoint usage
func APIUsageTrackingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip tracking for certain paths
		if shouldSkipTracking(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Record start time
		startTime := time.Now()

		// Capture request body if it exists
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Capture response using response writer wrapper
		responseWriter := &responseBodyCapture{
			ResponseWriter: c.Writer,
			body:           new(bytes.Buffer),
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Calculate response time in milliseconds
		responseTime := time.Since(startTime).Milliseconds()
		c.Set("responseTime", responseTime)

		// Extract user ID from JWT or context
		userID := extractUserID(c)

		// Prepare usage log
		usageLog := &api_usage.APIUsageLog{
			ID:           uuid.New().String(),
			Endpoint:     c.Request.URL.Path,
			Method:       c.Request.Method,
			StatusCode:   c.Writer.Status(),
			ResponseTime: responseTime,
			RequestSize:  len(requestBody),
			ResponseSize: responseWriter.body.Len(),
			UserID:       userID,
			ClientIP:     c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			RequestedAt:  startTime,
			CreatedAt:    time.Now(),
		}

		// Capture error message if present
		if c.Writer.Status() >= 400 {
			if len(responseWriter.body.Bytes()) > 0 {
				errorMsg := responseWriter.body.String()
				if len(errorMsg) > 500 {
					errorMsg = errorMsg[:500]
				}
				usageLog.ErrorMessage = &errorMsg
			}
		}

		// Store usage log asynchronously to avoid blocking
		go storeAPIUsageLog(usageLog)
	}
}

// responseBodyCapture wraps gin.ResponseWriter to capture response body
type responseBodyCapture struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyCapture) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseBodyCapture) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// shouldSkipTracking checks if the path should be skipped from tracking
func shouldSkipTracking(path string) bool {
	skipPaths := map[string]bool{
		"/":                 true,
		"/health":           true,
		"/swagger":          true,
		"/admin-ui/login":   true,
		"/admin-ui/metrics": true,
	}

	// Check exact matches
	if skipPaths[path] {
		return true
	}

	// Check prefixes
	skipPrefixes := []string{
		"/swagger/",
	}

	for _, prefix := range skipPrefixes {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}

	return false
}

// extractUserID attempts to extract user ID from context or JWT claims
func extractUserID(c *gin.Context) *string {
	// Try to get from gin context (set by JWT middleware)
	userID := GetUserID(c)
	if userID != "" {
		return &userID
	}

	// Try to get from JWT claims
	if claims, exists := c.Get("claims"); exists {
		if claimsMap, ok := claims.(map[string]interface{}); ok {
			if userID, ok := claimsMap["sub"].(string); ok {
				return &userID
			}
		}
	}

	return nil
}

// storeAPIUsageLog stores the API usage log in the database
func storeAPIUsageLog(log *api_usage.APIUsageLog) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic while storing API usage log", zap.Any("error", r))
		}
	}()

	if _, err := apiUsageRepo.Create(log); err != nil {
		logger.Error("Failed to store API usage log",
			zap.String("endpoint", log.Endpoint),
			zap.Error(err),
		)
	}
	err := UpdateAPIUsageStats(log.Endpoint, log.Method)
	if err != nil {
		// Include zap
		// logger.Error("Falied to store API Stats")
	}
}

// UpdateAPIUsageStats updates aggregated statistics for an endpoint
func UpdateAPIUsageStats(endpoint string, method string) error {
	statsRepo := persistence.NewAPIUsageStatsRepository(initializers.DB)
	return statsRepo.RecalculateStats(endpoint, method)
}

// CleanupOldAPIUsageLogs removes logs older than specified days
func CleanupOldAPIUsageLogs(days int) error {
	repo := persistence.NewAPIUsageRepository(initializers.DB)
	return repo.DeleteOlderThan(days)
}
