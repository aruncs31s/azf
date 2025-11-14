package enterprise

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aruncs31s/azf/config"
	"github.com/aruncs31s/azf/shared/logger"
	"github.com/casbin/casbin/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// EnterpriseAuthorizationSetup orchestrates initialization of all enterprise auth components
type EnterpriseAuthorizationSetup struct {
	db              *gorm.DB
	redis           *redis.Client
	logger          *zap.Logger
	routeRegistry   *RouteRegistry
	policyValidator PolicyValidator
	rateLimiter     RateLimiter
	auditRepository *AuthorizationAuditRepository
	middleware      *AZFAuthMiddleware
}

// SetupOptions holds all options for enterprise authorization setup
type SetupOptions struct {
	// Database connection
	Database *gorm.DB

	// Redis connection (optional, for distributed rate limiting)
	Redis *redis.Client

	// Policy file path
	PolicyFilePath string

	// Environment (developement, staging, production)
	Environment string

	// Rate limiting configuration
	EnableRateLimit   bool
	RateLimitConfig   *RateLimitConfig
	UseRedisRateLimit bool

	// Audit logging configuration
	EnableAuditLogging bool
	AuditBatchSize     int
	AuditFlushInterval time.Duration

	// Authorization configuration
	EnableDeprecationCheck bool
	GradualRolloutMode     bool // Allow missing policies during migration
	AllowMissingPolicies   bool
	ValidatePoliciesOnInit bool

	// Casbin enforcer instance (optional)
	CasbinEnforcer *casbin.Enforcer
	// Logger instance
	Logger *zap.Logger
}

// NewEnterpriseAuthorizationSetup creates a new enterprise authorization setup
func NewEnterpriseAuthorizationSetup(opts *SetupOptions) (*EnterpriseAuthorizationSetup, error) {
	if opts == nil {
		return nil, fmt.Errorf("setup options cannot be nil")
	}

	if opts.Database == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	if opts.Logger == nil {
		opts.Logger = logger.GetLogger()
	}

	if opts.PolicyFilePath == "" {
		opts.PolicyFilePath = config.CASBIN_POLICY_DEFAULT_PATH
	}

	if opts.Environment == "" {
		opts.Environment = os.Getenv("ENVIRONMENT")
		if opts.Environment == "" {
			opts.Environment = "development"
		}
	}

	setup := &EnterpriseAuthorizationSetup{
		db:     opts.Database,
		redis:  opts.Redis,
		logger: opts.Logger,
	}

	// Initialize components in order
	if err := setup.initializeRouteRegistry(); err != nil {
		return nil, getFailedToInitializeErr("route registry", err)
	}

	if err := setup.initializePolicyValidator(opts.PolicyFilePath, opts.ValidatePoliciesOnInit); err != nil {
		return nil, getFailedToInitializeErr("policy validator", err)
	}

	if err := setup.initializeAuditRepository(); err != nil {
		return nil, getFailedToInitializeErr("audit repository", err)
	}

	if err := setup.initializeRateLimiter(opts); err != nil {
		return nil, getFailedToInitializeErr("rate limiter", err)
	}

	if err := setup.initializeMiddleware(opts); err != nil {
		return nil, getFailedToInitializeErr("middleware", err)
	}

	setup.logger.Info("Enterprise authorization setup completed",
		zap.String("environment", opts.Environment),
		zap.Bool("audit_logging", opts.EnableAuditLogging),
		zap.Bool("rate_limiting", opts.EnableRateLimit),
		zap.Bool("deprecation_check", opts.EnableDeprecationCheck),
	)

	return setup, nil
}
func getFailedToInitializeErr(resource string, err error) error {
	return fmt.Errorf("failed to initialize %s: %w", resource, err)
}

// initializeRouteRegistry sets up the route registry
func (eas *EnterpriseAuthorizationSetup) initializeRouteRegistry() error {
	eas.routeRegistry = NewRouteRegistry()
	eas.logger.Info("Route registry initialized")
	return nil
}

// initializePolicyValidator sets up the policy validator
func (eas *EnterpriseAuthorizationSetup) initializePolicyValidator(policyFilePath string, validate bool) error {
	eas.policyValidator = NewPolicyValidator(eas.routeRegistry)

	if validate {
		report := eas.policyValidator.Validate()
		if !report.IsValid {
			eas.logger.Warn("Policy validation report",
				zap.Int("error_count", len(report.Errors)),
				zap.Int("warning_count", len(report.Warnings)),
				zap.String("report", report.String()))

			// Log individual errors
			for _, err := range report.Errors {
				eas.logger.Error("Policy validation error", zap.String("error", err))
			}
		} else {
			eas.logger.Info("Policy validation passed",
				zap.Int("total_policies", report.SummaryStats.TotalPolicies),
				zap.Int("total_routes", report.SummaryStats.TotalRoutes),
				zap.Float64("coverage_percentage", report.SummaryStats.CoveragePercentage))
		}
	}

	eas.logger.Info("Policy validator initialized",
		zap.String("policy_file", policyFilePath))

	return nil
}

// initializeAuditRepository sets up the audit repository
func (eas *EnterpriseAuthorizationSetup) initializeAuditRepository() error {
	eas.auditRepository = NewAuthorizationAuditRepository(eas.db, eas.logger)

	// Create table if it doesn't exist
	if !eas.db.Migrator().HasTable(&AuthorizationAuditLogDB{}) {
		if err := eas.db.Migrator().CreateTable(&AuthorizationAuditLogDB{}); err != nil {
			return fmt.Errorf("failed to create authorization audit log table: %w", err)
		}
		eas.logger.Info("Created authorization_audit_logs table")
	}

	// Get count of existing logs
	count, err := eas.auditRepository.Count(context.Background())
	if err != nil {
		eas.logger.Warn("Failed to get audit log count", zap.Error(err))
	} else {
		eas.logger.Info("Audit repository initialized", zap.Int64("existing_logs", count))
	}

	return nil
}

// initializeRateLimiter sets up the rate limiter
func (eas *EnterpriseAuthorizationSetup) initializeRateLimiter(opts *SetupOptions) error {
	if !opts.EnableRateLimit {
		eas.logger.Info("Rate limiting disabled")
		return nil
	}

	// Set default rate limit config if not provided
	if opts.RateLimitConfig == nil {
		opts.RateLimitConfig = &RateLimitConfig{
			DefaultRequestsPerMinute: 60,
			RoleSpecificLimits: map[string]int{
				"admin":     config.ADMIN_LIMIT,
				"staff":     config.STAFF_LIMIT,
				"student":   config.STUDENT_LIMIT,
				"moderator": config.MODERATOR_LIMIT,
			},
			BurstAllowance: config.BURST_ALLOWANCE,
			WindowDuration: time.Minute,
			EnableRedis:    opts.UseRedisRateLimit && opts.Redis != nil,
		}
	}

	// Create rate limiter
	if opts.UseRedisRateLimit && opts.Redis != nil {
		eas.rateLimiter = NewRedisRateLimiter(opts.RateLimitConfig, opts.Redis, eas.logger)
		eas.logger.Info("Redis rate limiter initialized",
			zap.Int("default_limit", opts.RateLimitConfig.DefaultRequestsPerMinute),
			zap.Int("burst_allowance", opts.RateLimitConfig.BurstAllowance))
	} else {
		eas.rateLimiter = NewInMemoryRateLimiter(opts.RateLimitConfig, eas.logger)
		eas.logger.Info("In-memory rate limiter initialized",
			zap.Int("default_limit", opts.RateLimitConfig.DefaultRequestsPerMinute),
			zap.Int("burst_allowance", opts.RateLimitConfig.BurstAllowance))
	}

	return nil
}

// initializeMiddleware sets up the enterprise auth middleware
func (eas *EnterpriseAuthorizationSetup) initializeMiddleware(opts *SetupOptions) error {
	middlewareConfig := &AZFAuthMiddlewareConfig{
		RateLimiter:            eas.rateLimiter,
		AuditRepository:        eas.auditRepository,
		RouteRegistry:          eas.routeRegistry,
		PolicyValidator:        eas.policyValidator,
		CasbinEnforcer:         opts.CasbinEnforcer,
		Logger:                 eas.logger,
		Environment:            opts.Environment,
		PolicyVersion:          config.POLICY_VERSION,
		EnableAuditLogging:     opts.EnableAuditLogging,
		EnableRateLimit:        opts.EnableRateLimit,
		EnableDeprecationCheck: opts.EnableDeprecationCheck,
		GradualRolloutMode:     opts.GradualRolloutMode,
		AllowMissingPolicies:   opts.AllowMissingPolicies,
	}

	eas.middleware = NewEnterpriseAuthMiddleware(middlewareConfig)

	eas.logger.Info("Enterprise auth middleware initialized",
		zap.Bool("audit_logging", opts.EnableAuditLogging),
		zap.Bool("rate_limiting", opts.EnableRateLimit),
		zap.Bool("deprecation_check", opts.EnableDeprecationCheck))

	return nil
}

// RegisterRoute registers a route with the registry
func (eas *EnterpriseAuthorizationSetup) RegisterRoute(metadata *RouteMetadata) error {
	return eas.routeRegistry.Register(metadata)
}

// RegisterRoutes registers multiple routes
func (eas *EnterpriseAuthorizationSetup) RegisterRoutes(metadatas ...*RouteMetadata) error {
	return eas.routeRegistry.RegisterMany(metadatas...)
}

// SetRoleRateLimit sets rate limit for a specific role
func (eas *EnterpriseAuthorizationSetup) SetRoleRateLimit(role string, requestsPerMinute int, burstAllowance int) {
	if inMemLimiter, ok := eas.rateLimiter.(*InMemoryRateLimiter); ok {
		inMemLimiter.SetRoleLimit(role, requestsPerMinute, burstAllowance)
		eas.logger.Info("Updated role rate limit",
			zap.String("role", role),
			zap.Int("requests_per_minute", requestsPerMinute),
			zap.Int("burst_allowance", burstAllowance),
		)
	}
}

// GetRouteRegistry returns the route registry
func (eas *EnterpriseAuthorizationSetup) GetRouteRegistry() *RouteRegistry {
	return eas.routeRegistry
}

// GetPolicyValidator returns the policy validator
func (eas *EnterpriseAuthorizationSetup) GetPolicyValidator() PolicyValidator {
	return eas.policyValidator
}

// GetRateLimiter returns the rate limiter
func (eas *EnterpriseAuthorizationSetup) GetRateLimiter() RateLimiter {
	return eas.rateLimiter
}

// GetAuditRepository returns the audit repository
func (eas *EnterpriseAuthorizationSetup) GetAuditRepository() *AuthorizationAuditRepository {
	return eas.auditRepository
}

// GetMiddleware returns the enterprise auth middleware
func (eas *EnterpriseAuthorizationSetup) GetMiddleware() *AZFAuthMiddleware {
	return eas.middleware
}

// ValidateAllPolicies validates all policies and routes
func (eas *EnterpriseAuthorizationSetup) ValidateAllPolicies() *PolicyValidationReport {
	return eas.policyValidator.Validate()
}

// GetAuditStats returns audit statistics
func (eas *EnterpriseAuthorizationSetup) GetAuditStats(ctx context.Context) (map[string]interface{}, error) {
	totalLogs, err := eas.auditRepository.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get total logs: %w", err)
	}

	deniedLogs, err := eas.auditRepository.FindDeniedAccess(ctx, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get denied access count: %w", err)
	}

	stats := map[string]interface{}{
		"total_audit_logs": totalLogs,
		"recent_denied":    len(deniedLogs),
		"timestamp":        time.Now(),
	}

	return stats, nil
}

// CleanupOldAuditLogs removes audit logs older than the specified duration
func (eas *EnterpriseAuthorizationSetup) CleanupOldAuditLogs(ctx context.Context, olderThan time.Duration) (int64, error) {
	return eas.auditRepository.CleanupOldLogs(ctx, olderThan)
}

// Stop gracefully stops the setup (flushes batches, closes connections)
func (eas *EnterpriseAuthorizationSetup) Stop() {
	eas.logger.Info("Stopping enterprise authorization setup")

	if eas.middleware != nil {
		eas.middleware.Stop()
	}

	if inMemLimiter, ok := eas.rateLimiter.(*InMemoryRateLimiter); ok {
		inMemLimiter.Stop()
	}

	eas.logger.Info("Enterprise authorization setup stopped")
}

// PrintRoutesReport prints a detailed report of all registered routes
func (eas *EnterpriseAuthorizationSetup) PrintRoutesReport() string {
	report := "=== REGISTERED ROUTES REPORT ===\n\n"

	routes := eas.routeRegistry.GetAll()
	report += fmt.Sprintf("Total Routes: %d\n\n", len(routes))

	// Group by tag
	tagGroups := make(map[string][]*RouteMetadata)
	for _, metadata := range routes {
		if len(metadata.Tags) == 0 {
			tagGroups["untagged"] = append(tagGroups["untagged"], metadata)
		} else {
			for _, tag := range metadata.Tags {
				tagGroups[tag] = append(tagGroups[tag], metadata)
			}
		}
	}

	for tag, routes := range tagGroups {
		report += fmt.Sprintf("--- %s (%d routes) ---\n", tag, len(routes))
		for _, route := range routes {
			deprecated := ""
			if route.Deprecated {
				deprecated = " [DEPRECATED]"
			}
			report += fmt.Sprintf("  %s %s%s\n", route.Method, route.Path, deprecated)
			if route.Description != "" {
				report += fmt.Sprintf("    Description: %s\n", route.Description)
			}
			if len(route.AllowedRoles) > 0 {
				report += fmt.Sprintf("    Allowed Roles: %v\n", route.AllowedRoles)
			}
		}
		report += "\n"
	}

	return report
}

// PrintValidationReport prints the policy validation report
func (eas *EnterpriseAuthorizationSetup) PrintValidationReport() string {
	report := eas.policyValidator.Validate()
	return report.String()
}
