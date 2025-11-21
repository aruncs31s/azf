package azf

import (
	"github.com/aruncs31s/azf/application/handler"
	"github.com/aruncs31s/azf/application/middleware"
	"github.com/aruncs31s/azf/application/service"
	"github.com/aruncs31s/azf/config"
	"github.com/aruncs31s/azf/infrastructure/enterprise"
	"github.com/aruncs31s/azf/infrastructure/persistence"
	"github.com/aruncs31s/azf/initializer"
	"github.com/aruncs31s/azf/shared/logger"
	"github.com/aruncs31s/azf/utils"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

// EnterpriseRouteMetadataConfig holds the structure of the enterprise route metadata config
type EnterpriseRouteMetadataConfig struct {
	Routes []*enterprise.RouteMetadata `json:"routes"`
}

// mgr holds the initializer.Manager instance created by InitAuthZModule.
// It is used to centralize DB/enforcer access while still providing compatibility.
var mgr *initializer.Manager

// InitAuthZModule Initializes new Authorization Instance , of the AZF AuthZ Framework
//
// Params:
// tempDB - *gorm.DB , it is a temporary db used to save analytics for latter audit
func InitAuthZModule(
	tempDB *gorm.DB, // IF provided will use this
	casbinEnforcer *casbin.Enforcer, // If provide use this also
) {
	// Initialize Logger Before Anything Else
	logger.InitLogger()

	// Create and initialize the Manager. This encapsulates DB and Casbin enforcer.
	// For compatibility with existing code that still reads package-level variables,
	// we set those globals after manager initialization.
	var err error
	mgr, err = initializer.NewAndInitManager(tempDB, casbinEnforcer, logger.GetLogger())
	if err != nil {
		logger.GetLogger().Fatal("Failed to initialize authorization manager", zap.Error(err))
	}

	// For backward compatibility with code that expects initializer package globals,
	// populate them from the manager. This keeps behavior unchanged while enabling
	// a path toward removing global usage elsewhere.
	if mgr != nil {
		initializer.DB = mgr.DB
		initializer.CasbinEnforcer = mgr.Enforcer
	}

	err = enterprise.IniAuthorization(
		mgr.DB,
		nil,
		mgr.Enforcer,
		logger.GetLogger(),
	)
	if err != nil {
		logger.GetLogger().Fatal("Failed to initialize authorization", zap.Error(err))
	}

	if enterprise.EnterpriseAuth != nil {
		enterprise.RegisterEnterpriseRouteMetadata(enterprise.EnterpriseAuth)
	} else {
		logger.Warn("Enterprise authorization setup not available, running in compatibility mode")
	}
}

func InitUsageTracking() {
	// Create API Usage Tracking Repo; prefer manager DB if available
	var db *gorm.DB
	if mgr != nil && mgr.DB != nil {
		db = mgr.DB
	} else {
		db = initializer.DB
	}

	if db == nil {
		logger.Warn("API usage tracking not initialized: database is not available")
		return
	}

	apiUsageRepo := persistence.NewAPIUsageRepository(db)

	// Initialize the UsageTracking Middleware
	middleware.InitAPIUsageTracking(apiUsageRepo)

}
func SetApiTrackingMiddleware(r *gin.Engine) *gin.Engine {
	// Use The API Usage Tracking Middleware
	r.Use(middleware.APIUsageTrackingMiddleware())
	return r
}

func SetAuthZMiddleware(r *gin.Engine) *gin.Engine {
	if enterprise.EnterpriseAuth != nil {
		r.Use(enterprise.EnterpriseAuth.GetMiddleware().GinMiddleware())
	} else {
		logger.Warn("Enterprise authorization setup not available, authorization middleware not applied")
	}
	return r
}

func SetRateLimitMiddleware(r *gin.Engine, requestsPerSecond float64, burst int) *gin.Engine {
	limiter := middleware.NewIPRateLimiter(rate.Limit(requestsPerSecond), burst)
	r.Use(middleware.RateLimitMiddleware(limiter))
	return r
}

// loadEnterpriseRouteMetadata loads route metadata from a JSON configuration file

func StopAuthZModule() {
	// Stop enterprise-level resources if present
	if enterprise.EnterpriseAuth != nil {
		enterprise.EnterpriseAuth.Stop()
	}
	// Close manager resources (DB) if created
	if mgr != nil {
		_ = mgr.Close()
		mgr = nil
		// also clear compatibility globals
		initializer.DB = nil
		initializer.CasbinEnforcer = nil
	}
}

func GetLogger() *zap.Logger {
	if logger.Log == nil {
		logger.InitLogger()
	}
	return logger.GetLogger()
}
func SetupUI(r *gin.Engine) *gin.Engine {
	configProvider, _ := config.NewAdminConfigProvider()
	apiPerfHandler := handler.NewPerformanceHandler(configProvider)

	// Initialize rate limiting manager
	rateLimitManager := handler.NewRateLimitManager(10, 20) // 10 requests/second, burst 20
	rateLimitHandler := handler.NewRateLimitHandler(rateLimitManager)

	// Initialize OAuth service and handler if user repository is available
	var oauthHandler *handler.OAuthHandler
	if mgr != nil && mgr.DB != nil {
		userRepo := persistence.NewUserRepository(mgr.DB)
		baseURL := "http://localhost:8080" // default
		if envURL, err := utils.GetEnv("BASE_URL"); err == nil {
			baseURL = envURL
		}
		oauthService := service.NewOAuthService(userRepo, baseURL, "your-jwt-secret") // TODO: Get JWT secret from config
		oauthHandler = handler.NewOAuthHandler(oauthService)
	}

	r.GET("/admin-ui/login", apiPerfHandler.GetLoginPage)

	r.POST("/admin-ui/login/json", apiPerfHandler.LoginJSON)

	// OAuth routes
	if oauthHandler != nil {
		r.GET("/admin-ui/oauth/:provider", oauthHandler.Login)
		r.GET("/admin-ui/oauth/callback/:provider", oauthHandler.Callback)
		r.GET("/admin-ui/oauth/providers", oauthHandler.GetProviders)
	}

	r.GET("/admin-ui", middleware.CheckAdminAuth(), apiPerfHandler.GetHomePage)
	r.GET("/admin-ui/api_analytics", middleware.CheckAdminAuth(), apiPerfHandler.GetAPIAnalyticsPage)
	r.GET("/admin-ui/api_analytics/endpoint", middleware.CheckAdminAuth(), apiPerfHandler.GetEndpointDetailsPage)
	r.GET("/admin-ui/route_metadata", middleware.CheckAdminAuth(), apiPerfHandler.GetRouteMetadataManagementPage)
	r.POST("/admin-ui/route_metadata", middleware.CheckAdminAuth(), apiPerfHandler.SaveRouteMetadata)
	r.POST("/admin-ui/route_metadata/import", middleware.CheckAdminAuth(), apiPerfHandler.ImportRouteMetadata)
	r.POST("/admin-ui/route_metadata/delete", middleware.CheckAdminAuth(), apiPerfHandler.DeleteRouteMetadata)
	r.GET("/admin-ui/roles", middleware.CheckAdminAuth(), apiPerfHandler.GetRoleManagementPage)
	r.GET("/admin-ui/roles/:role", middleware.CheckAdminAuth(), apiPerfHandler.GetRoleDetailsPage)
	r.GET("/admin-ui/policies", middleware.CheckAdminAuth(), apiPerfHandler.GetPolicyManagementPage)
	r.GET("/admin-ui/audit_logs", middleware.CheckAdminAuth(), apiPerfHandler.GetAuditLogsPage)
	r.GET("/admin-ui/features", middleware.CheckAdminAuth(), apiPerfHandler.GetFeaturesDocumentationPage)

	// Role management API endpoints
	r.POST("/admin-ui/api/roles", middleware.CheckAdminAuth(), apiPerfHandler.CreateRole)
	r.PUT("/admin-ui/api/roles", middleware.CheckAdminAuth(), apiPerfHandler.UpdateRole)
	r.POST("/admin-ui/api/roles/assign", middleware.CheckAdminAuth(), apiPerfHandler.AssignRoleToUser)
	r.POST("/admin-ui/api/roles/remove", middleware.CheckAdminAuth(), apiPerfHandler.RemoveRoleFromUser)
	r.GET("/admin-ui/api/roles/users", middleware.CheckAdminAuth(), apiPerfHandler.GetUsersForRole)
	r.POST("/admin-ui/api/roles/delete", middleware.CheckAdminAuth(), apiPerfHandler.DeleteRole)

	// Rate limiting routes
	r.GET("/admin-ui/rate-limits", middleware.CheckAdminAuth(), rateLimitHandler.GetRateLimitPage)
	r.GET("/admin-ui/api/rate-limits/stats", middleware.CheckAdminAuth(), rateLimitHandler.GetRateLimitStats)
	r.GET("/admin-ui/api/rate-limits/ip/:ip", middleware.CheckAdminAuth(), rateLimitHandler.GetIPStats)
	r.DELETE("/admin-ui/api/rate-limits/ip/:ip", middleware.CheckAdminAuth(), rateLimitHandler.ResetIPLimit)
	r.PUT("/admin-ui/api/rate-limits/global", middleware.CheckAdminAuth(), rateLimitHandler.UpdateGlobalLimit)
	r.POST("/admin-ui/api/rate-limits/endpoint", middleware.CheckAdminAuth(), rateLimitHandler.SetEndpointLimit)
	r.GET("/admin-ui/api/rate-limits/endpoints", middleware.CheckAdminAuth(), rateLimitHandler.GetEndpointLimits)
	r.DELETE("/admin-ui/api/rate-limits/reset-all", middleware.CheckAdminAuth(), rateLimitHandler.ResetAllLimits)
	r.GET("/admin-ui/api/rate-limits/search", middleware.CheckAdminAuth(), rateLimitHandler.SearchRateLimitStats)
	r.GET("/admin-ui/api/rate-limits/export", middleware.CheckAdminAuth(), rateLimitHandler.ExportRateLimitStats)

	r.GET("/admin-ui/logout", apiPerfHandler.Logout)
	return r
}
