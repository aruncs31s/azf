package enterprise

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aruncs31s/azf/application/dto"
	"github.com/aruncs31s/azf/config"
	"github.com/aruncs31s/azf/shared/logger"
	"github.com/aruncs31s/azf/utils"

	"github.com/aruncs31s/azf/application/middleware"
	"github.com/aruncs31s/azf/domain/model"
	helperImpl "github.com/aruncs31s/azf/shared/helper"
	"github.com/aruncs31s/azf/shared/interface/helper"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AZFAuthMiddlewareConfig holds configuration for the middleware
type AZFAuthMiddlewareConfig struct {
	CasbinEnforcer         *casbin.Enforcer // *casbin.Enforcer
	PolicyValidator        PolicyValidator
	RateLimiter            RateLimiter
	AuditRepository        *AuthorizationAuditRepository
	RouteRegistry          *RouteRegistry
	Logger                 *zap.Logger
	Environment            string // dev, staging, production
	PolicyVersion          int
	EnableAuditLogging     bool
	EnableRateLimit        bool
	EnableDeprecationCheck bool
	GradualRolloutMode     bool // If true, denies access but logs as WARNING instead of DENIED
	AllowMissingPolicies   bool // If true, missing policies are allowed (soft migration)
}

// AZFAuthMiddleware provides comprehensive authorization with audit trail
type AZFAuthMiddleware struct {
	config                *AZFAuthMiddlewareConfig
	responseHelper        helper.ResponseHelper
	requestHelper         helper.RequestHelper
	auditBatch            []*model.AuthorizationAuditLog
	batchSize             int
	batchFlushInterval    time.Duration
	stopBatchProcessor    chan bool
	batchProcessorRunning bool
	auditMutex            sync.Mutex
}

// NewEnterpriseAuthMiddleware creates a new enterprise auth middleware
func NewEnterpriseAuthMiddleware(
	config *AZFAuthMiddlewareConfig) *AZFAuthMiddleware {
	requestHelper := helperImpl.NewRequestHelper()
	responseHelper := helperImpl.NewResponseHelper()
	if config.Logger == nil {
		config.Logger = logger.GetLogger()
	}

	middleware := &AZFAuthMiddleware{
		config:             config,
		responseHelper:     responseHelper,
		requestHelper:      requestHelper,
		auditBatch:         make([]*model.AuthorizationAuditLog, 0),
		batchSize:          100,
		batchFlushInterval: 10 * time.Second,
		stopBatchProcessor: make(chan bool),
	}

	// Start batch processor if audit logging is enabled
	if config.EnableAuditLogging && config.AuditRepository != nil {
		middleware.startBatchProcessor()
	}

	return middleware
}

// GinMiddleware returns a Gin middleware handler
// This is used insted of the casbin middleware to provide enhanced features
// but for some reason if this is not available we  fallback to casbin middleware
func (eam *AZFAuthMiddleware) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		eam.authorizeRequest(c)
	}
}

// authorizeRequest handles the authorization logic
func (eam *AZFAuthMiddleware) authorizeRequest(c *gin.Context) {
	requestID := uuid.New().String()
	startTime := time.Now()

	// Get route information
	path := utils.NormalizePathForLookup(c.Request.URL.Path)
	eam.config.Logger.Info("Path Notmalization Cheking",
		zap.String("path before normalization", c.Request.URL.Path),
		zap.String("path after normalization", path),
	)

	method := c.Request.Method
	// Check if route is in registry
	routeMetadata, routeExists := eam.config.RouteRegistry.Get(path, method)
	// 1. Check if route is public - if so, allow access without authentication
	if routeExists && routeMetadata.IsPublic {
		eam.config.Logger.Debug(
			"Public route accessed",
			zap.String("path", path),
			zap.String("method", method),
		)
		c.Set("meta", eam.buildResponseMeta("PUBLIC"))
		c.Next()
		return
	}

	// Get user context
	userRole, userID, ipAddress := eam.extractUserContext(c)
	if userRole == "" {
		eam.handleUnauthorized(c, "User role not found", requestID)
		return
	}
	// 2. Check for deprecation
	if routeExists && routeMetadata.Deprecated {
		eam.config.Logger.Warn(
			"Deprecated route accessed",
			zap.String("user_id", userID),
			zap.String("role", userRole),
			zap.String("path", path),
			zap.String("message", routeMetadata.GetDeprecationMessage()),
		)

		// Add deprecation warning header
		c.Header("X-API-Warn", routeMetadata.GetDeprecationMessage())

		if routeMetadata.ReplacedBy != "" {
			c.Header("X-API-Deprecation-Use-Instead", routeMetadata.ReplacedBy)
		}
	}

	// 3. Check rate limiting
	if eam.config.EnableRateLimit && routeExists && routeMetadata.RateLimit != nil {
		rateLimitStatus, err := eam.config.RateLimiter.CheckLimit(c.Request.Context(), userID, userRole)
		if err != nil {
			eam.config.Logger.Error("Rate limit check failed", zap.Error(err))
		}

		if rateLimitStatus != nil && rateLimitStatus.LimitExceeded {
			eam.config.Logger.Warn(
				"Rate limit exceeded",
				zap.String("user_id", userID),
				zap.String("role", userRole),
				zap.String("path", path),
				zap.Int("retry_after", rateLimitStatus.RetryAfterSeconds),
			)

			// Log audit
			if eam.config.EnableAuditLogging {
				eam.logAuthorizationAudit(
					requestID, userID, userRole, path, method,
					model.AuthzDenied, model.ReasonRateLimitExceeded,
					ipAddress, c.Request.UserAgent(),
					time.Since(startTime).Milliseconds(),
					true, // rate limit exceeded
				)
			}

			c.Header("Retry-After", fmt.Sprintf("%d", rateLimitStatus.RetryAfterSeconds))
			c.Header("X-Rate-Limit-Remaining", fmt.Sprintf("%d", rateLimitStatus.RemainingRequests))
			c.Header("X-Rate-Limit-Reset", fmt.Sprintf("%d", rateLimitStatus.ResetAtTime.Unix()))

			eam.responseHelper.BadRequest(c, "Rate limit exceeded", "")
			c.Abort()
			return
		}

		// Add rate limit headers
		if rateLimitStatus != nil {
			c.Header("X-Rate-Limit-Remaining", fmt.Sprintf("%d", rateLimitStatus.RemainingRequests))
			c.Header("X-Rate-Limit-Reset", fmt.Sprintf("%d", rateLimitStatus.ResetAtTime.Unix()))
		}
	}

	// 4. Check authorization via Casbin
	eam.config.Logger.Debug("About to check permission",
		zap.String("user_id", userID),
		zap.String("role", userRole),
		zap.String("path", path),
		zap.String("method", method),
	)
	allowed := eam.checkPermission(userRole, path, method)

	// 5. Log audit
	if eam.config.EnableAuditLogging {
		if allowed {
			eam.logAuthorizationAudit(
				requestID, userID, userRole, path, method,
				model.AuthzAllowed, nil,
				ipAddress, c.Request.UserAgent(),
				time.Since(startTime).Milliseconds(),
				false,
			)
		} else {
			reason := model.ReasonPolicyNotFound
			if routeExists {
				reason = model.ReasonRoleNotFound
			}

			eam.logAuthorizationAudit(
				requestID, userID, userRole, path, method,
				model.AuthzDenied, reason,
				ipAddress, c.Request.UserAgent(),
				time.Since(startTime).Milliseconds(),
				false,
			)
		}
	}

	// 6. Handle authorization result
	if !allowed {
		// Check if we're in gradual rollout mode
		if eam.config.GradualRolloutMode {
			eam.config.Logger.Warn(
				"Access denied (gradual rollout mode - allowing)",
				zap.String("user_id", userID),
				zap.String("role", userRole),
				zap.String("path", path),
			)
			c.Header("X-Authorization-Mode", "GRADUAL_ROLLOUT")
			c.Set("meta", eam.buildResponseMeta(config.AUTH_MODE_GRADUAL_ROLLOUT))
			c.Next()
			return
		}

		// Check if missing policies are allowed
		if eam.config.AllowMissingPolicies && !routeExists {
			eam.config.Logger.Debug(
				"Route not in registry, allowing (soft migration mode)",
				zap.String("user_id", userID),
				zap.String("role", userRole),
				zap.String("path", path),
			)
			c.Header("X-Authorization-Mode", config.AUTH_MODE_SOFT_MIGRATION)
			c.Set("meta", eam.buildResponseMeta(config.AUTH_MODE_SOFT_MIGRATION))
			c.Next()
			return
		}

		eam.config.Logger.Warn(
			"Access denied",
			zap.String("user_id", userID),
			zap.String("role", userRole),
			zap.String("path", path),
			zap.String("method", method),
		)

		c.Set("meta", eam.buildResponseMeta(config.AUTH_MODE_CASBIN))
		eam.responseHelper.Forbidden(c, "Access denied")
		c.Abort()
		return
	}

	// Set context values for handlers
	c.Set("request_id", requestID)
	c.Set("authorization_checked", true)

	eam.config.Logger.Debug(
		"Authorization granted",
		zap.String("user_id", userID),
		zap.String("role", userRole),
		zap.String("path", path),
		zap.String("method", method),
	)

	c.Set("meta", eam.buildResponseMeta(config.AUTH_MODE_CASBIN))
	c.Next()
}
func (eam *AZFAuthMiddleware) buildResponseMeta(mode string) dto.ResponseMeta {
	return dto.ResponseMeta{
		APIVersion:        os.Getenv("API_VERSION"),
		AuthorizationMode: mode,
		ResponseTimeMs:    "DEV: Will Implement Later",
		RequestID:         uuid.NewString(),
	}
}

// extractUserContext extracts  user information from request
func (eam *AZFAuthMiddleware) extractUserContext(c *gin.Context) (role, userID, ipAddress string) {
	// Get role from context (set by JWT middleware)
	role = middleware.GetUserRole(c)

	// Get user ID from context
	userID = middleware.GetUserID(c)

	// Get IP address
	ipAddress = c.ClientIP()

	return
}

// checkPermission checks if user has permission using Casbin
func (eam *AZFAuthMiddleware) checkPermission(role, resource, action string) bool {
	// Normalize path to align with policy patterns (e.g., convert numeric IDs to :id)
	normalized := utils.NormalizePathForLookup(resource)

	eam.config.Logger.Debug("Permission check started",
		zap.String("role", role),
		zap.String("original_resource", resource),
		zap.String("normalized_resource", normalized),
		zap.String("action", action),
	)

	// Enforcer must be provided via config; if missing deny
	if eam.config.CasbinEnforcer == nil {
		eam.config.Logger.Warn("Casbin enforcer not configured - denying",
			zap.String("role", role),
			zap.String("resource", normalized),
			zap.String("action", action),
		)
		return false
	}

	enforcer := eam.config.CasbinEnforcer
	// enforcer, ok := eam.config.CasbinEnforcer
	// if !ok || enforcer == nil {
	// 	eam.config.Logger.Error("Invalid Casbin enforcer type - denying")
	// 	return false
	// }

	allowed, err := enforcer.Enforce(role, normalized, action)
	if err != nil {
		eam.config.Logger.Error("Casbin enforce error", zap.Error(err),
			zap.String("role", role),
			zap.String("resource", normalized),
			zap.String("action", action),
		)
		return false
	}

	eam.config.Logger.Debug("Permission check result",
		zap.String("role", role),
		zap.String("resource", normalized),
		zap.String("action", action),
		zap.Bool("allowed", allowed),
	)

	return allowed
}

// logAuthorizationAudit logs authorization event
func (eam *AZFAuthMiddleware) logAuthorizationAudit(
	requestID, userID, role, resource, action string,
	result *model.AuthorizationResult,
	denialReason *model.DenialReason,
	ipAddress, userAgent string,
	executionTimeMs int64,
	rateLimitExceeded bool,
) {
	auditLog, err := model.NewAuthorizationAuditLog(
		uuid.New().String(),
		time.Now(),
		userID,
		role,
		resource,
		action,
		result,
		denialReason,
		ipAddress,
		userAgent,
		"v1",  // API version
		false, // deprecated
		eam.config.Environment,
		"OK", // rate limit status
		eam.config.PolicyVersion,
		float64(executionTimeMs),
		make(map[string]interface{}),
	)

	if err != nil {
		eam.config.Logger.Error("Failed to create authorization audit log", zap.Error(err))
		return
	}
	// Thread-safe append to batch
	eam.auditMutex.Lock()
	eam.auditBatch = append(eam.auditBatch, auditLog)
	shouldFlush := len(eam.auditBatch) >= eam.batchSize
	eam.auditMutex.Unlock()

	if shouldFlush {
		// Run flush asynchronously to avoid blocking request path
		go eam.flushAuditBatch()
	}

}

// flushAuditBatch saves batched audit logs to database
func (eam *AZFAuthMiddleware) flushAuditBatch() {
	if len(eam.auditBatch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := eam.config.AuditRepository.SaveBatch(ctx, eam.auditBatch)
	if err != nil {
		eam.config.Logger.Error("Failed to flush audit batch", zap.Error(err), zap.Int("count", len(eam.auditBatch)))
		return
	}

	eam.config.Logger.Debug("Audit batch flushed", zap.Int("count", len(eam.auditBatch)))
	eam.auditBatch = make([]*model.AuthorizationAuditLog, 0)
}

// startBatchProcessor starts the batch processor goroutine
func (eam *AZFAuthMiddleware) startBatchProcessor() {
	if eam.batchProcessorRunning {
		return
	}

	eam.batchProcessorRunning = true

	go func() {
		ticker := time.NewTicker(eam.batchFlushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				eam.flushAuditBatch()
			case <-eam.stopBatchProcessor:
				eam.flushAuditBatch() // Final flush
				return
			}
		}
	}()
}

// Stop stops the middleware and flushes remaining audit logs
func (eam *AZFAuthMiddleware) Stop() {
	if eam.batchProcessorRunning {
		eam.stopBatchProcessor <- true
		eam.batchProcessorRunning = false
	}
}

// handleUnauthorized handles unauthorized requests
func (eam *AZFAuthMiddleware) handleUnauthorized(c *gin.Context, message, requestID string) {
	path := utils.NormalizePathForLookup(c.Request.URL.Path)
	method := c.Request.Method

	eam.config.Logger.Warn(
		"Unauthorized access attempt",
		zap.String("request_id", requestID),
		zap.String("path", c.Request.URL.Path),
		zap.String("message", message),
	)

	// Create audit log for registered routes
	if eam.config.EnableAuditLogging {
		routeMetadata, routeExists := eam.config.RouteRegistry.Get(path, method)
		if routeExists && routeMetadata.AuditRequired {
			// Extract what we can from the request
			ipAddress := c.ClientIP()
			userAgent := c.Request.UserAgent()

			eam.logAuthorizationAudit(
				requestID, "", "", path, method, // No user info available
				model.AuthzDenied, model.ReasonRoleNotFound,
				ipAddress, userAgent,
				0, // execution time not available
				false,
			)
		}
	}

	eam.responseHelper.Unauthorized(c, message)
	c.Abort()
}
