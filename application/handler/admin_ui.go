package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/aruncs31s/azf/application/dto"
	"github.com/aruncs31s/azf/application/service"
	"github.com/aruncs31s/azf/application/templates"
	"github.com/aruncs31s/azf/config"
	"github.com/aruncs31s/azf/domain/api_usage"
	"github.com/aruncs31s/azf/infrastructure/enterprise"
	"github.com/aruncs31s/azf/infrastructure/persistence"
	"github.com/aruncs31s/azf/initializer"
	helperImpl "github.com/aruncs31s/azf/shared/helper"
	"github.com/aruncs31s/azf/shared/interface/helper"
	"github.com/aruncs31s/azf/shared/logger"
	"github.com/aruncs31s/azf/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// NewPerformanceHandler creates a new PerformanceHandler with its dependencies.
func NewPerformanceHandler(configProvider *config.AdminConfigProvider) PerformanceHandler {
	logRepo := persistence.NewAPIUsageRepository(initializer.DB)
	statsRepo := persistence.NewAPIUsageStatsRepository(initializer.DB)
	apiUsageAnalytics := service.NewAPIUsageAnalyticsService(logRepo, statsRepo)
	authService := service.NewAdminAuthenticationService(configProvider)
	profileService := service.NewAdminProfileService(configProvider)

	return &performanceHandler{
		apiUsageAnalytics: apiUsageAnalytics,
		authService:       *authService,
		profileService:    *profileService,
		auditService:      nil, // Will be initialized lazily
	}
}

type PerformanceHandler interface {
	PerformanceReader
	PerformanceWriter
}

type PerformanceReader interface {
	GetHomePage(c *gin.Context)
	GetAPIAnalyticsPage(c *gin.Context)
	GetEndpointDetailsPage(c *gin.Context)
	GetRouteMetadataManagementPage(c *gin.Context)
	GetRoleManagementPage(c *gin.Context)
	GetRoleDetailsPage(c *gin.Context)
	GetPolicyManagementPage(c *gin.Context)
	GetAuditLogsPage(c *gin.Context)
	GetFeaturesDocumentationPage(c *gin.Context)
	GetLoginPage(c *gin.Context)
	GetUsersForRole(c *gin.Context)
}

type PerformanceWriter interface {
	SaveRouteMetadata(c *gin.Context)
	ImportRouteMetadata(c *gin.Context)
	DeleteRouteMetadata(c *gin.Context)
	LoginJSON(c *gin.Context)
	Logout(c *gin.Context)
	CreateRole(c *gin.Context)
	UpdateRole(c *gin.Context)
	AssignRoleToUser(c *gin.Context)
	RemoveRoleFromUser(c *gin.Context)
	DeleteRole(c *gin.Context)
}

// performanceHandler serves the Admin Performance Dashboard and metrics JSON.
type performanceHandler struct {
	apiUsageAnalytics service.APIUsageAnalyticsService
	authService       service.AdminAuthenticationService
	profileService    service.AdminProfileService
	auditService      service.AuthorizationAuditService
	requestHelper     helper.RequestHelper
	responseHelper    helper.ResponseHelper
}

func (h *performanceHandler) GetLoginPage(c *gin.Context) {
	templ.Handler(templates.LoginPage("")).ServeHTTP(c.Writer, c.Request)
}

// GetHomePage renders the home dashboard page
// It shows an overview of all available features and management tools
func (h *performanceHandler) GetHomePage(c *gin.Context) {
	// Get admin username from claims if available
	adminUsername := "Admin"
	if claims, ok := c.Get("claims"); ok {
		if tokenClaims, ok := claims.(jwt.MapClaims); ok {
			if username, ok := tokenClaims["username"].(string); ok {
				adminUsername = username
			}
		}
	}

	// Get route count
	routeMetadata, err := enterprise.LoadEnterpriseRouteMetadata("")
	totalRoutes := 0
	if err == nil && routeMetadata != nil {
		totalRoutes = len(routeMetadata)
	}

	// Get audit logs count
	totalAuditLogs, _ := h.apiUsageAnalytics.GetUsageSummary()
	auditLogsCount := 0
	if totalAuditLogs != nil {
		auditLogsCount = int(totalAuditLogs.TotalRequests)
	}

	// Get total requests
	usageSummary, _ := h.apiUsageAnalytics.GetUsageSummary()
	totalRequests := 0
	avgResponseTime := 0.0
	if usageSummary != nil {
		totalRequests = int(usageSummary.TotalRequests)
		avgResponseTime = float64(usageSummary.AvgResponseTime)
	}

	// Get admin profile data
	adminProfile := make(map[string]interface{})
	profileSummary, err := h.profileService.GetProfileSummary()
	if err == nil {
		adminProfile = profileSummary
	}

	// Create home data
	homeData := templates.HomePageData{
		AdminUsername:       adminUsername,
		TotalRoutes:         totalRoutes,
		TotalAuditLogs:      auditLogsCount,
		TotalRequests:       totalRequests,
		AverageResponseTime: avgResponseTime,
		AdminProfile:        adminProfile,
	}

	// Render Templ template with sidebar
	templ.Handler(templates.HomePageWithSidebar(homeData)).ServeHTTP(c.Writer, c.Request)
}

// GetAPIAnalyticsPage renders the dedicated API Analytics page
// It shows comprehensive API usage statistics and performance metrics
func (h *performanceHandler) GetAPIAnalyticsPage(c *gin.Context) {
	// Get top endpoints
	topEndpoints, err := h.apiUsageAnalytics.GetTopEndpointsByUsage(10)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load API analytics")
		return
	}
	if topEndpoints == nil {
		topEndpoints = &[]api_usage.APIEndpointRanking{}
	}

	// Get slowest endpoints
	slowestEndpoints, err := h.apiUsageAnalytics.GetEndpointsByResponseTime(10)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load API analytics")
		return
	}
	if slowestEndpoints == nil {
		slowestEndpoints = &[]api_usage.APIEndpointRanking{}
	}

	// Get most errored endpoints
	erroredEndpoints, err := h.apiUsageAnalytics.GetEndpointsByErrorRate(10)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load API analytics")
		return
	}
	if erroredEndpoints == nil {
		erroredEndpoints = &[]api_usage.APIEndpointRanking{}
	}

	// Get usage summary
	usageSummary, err := h.apiUsageAnalytics.GetUsageSummary()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load API analytics")
		return
	}
	if usageSummary == nil {
		usageSummary = &service.UsageSummaryDTO{}
	}

	// Get trend data
	trendData, err := h.apiUsageAnalytics.GetUsageTrend(7)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load API analytics")
		return
	}
	if trendData == nil {
		trendData = &[]service.UsageTrendDTO{}
	}

	// Create analytics data structure for Templ
	analyticsData := templates.APIAnalyticsPageData{
		GeneratedAt:          time.Now(),
		TopEndpoints:         *topEndpoints,
		SlowestEndpoints:     *slowestEndpoints,
		MostErroredEndpoints: *erroredEndpoints,
		UsageSummary:         *usageSummary,
		TrendData:            *trendData,
	}

	// Render Templ template
	templ.Handler(templates.APIAnalyticsPage(analyticsData)).ServeHTTP(c.Writer, c.Request)
}

// GetEndpointDetailsPage renders the detailed view of who called a specific endpoint
func (h *performanceHandler) GetEndpointDetailsPage(c *gin.Context) {
	endpoint := c.Query("endpoint")
	method := c.Query("method")

	if endpoint == "" {
		c.String(http.StatusBadRequest, "Endpoint parameter is required")
		return
	}

	// Get endpoint details
	details, err := h.apiUsageAnalytics.GetEndpointDetails(endpoint)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load endpoint details")
		return
	}

	// Get callers
	callers, err := h.apiUsageAnalytics.GetEndpointCallers(endpoint, 50)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load endpoint callers")
		return
	}

	// Create page data
	pageData := templates.EndpointDetailsPageData{
		Endpoint: endpoint,
		Method:   method,
		Details:  *details,
		Callers:  *callers,
	}

	// Render Templ template
	templ.Handler(templates.EndpointDetailsPage(pageData)).ServeHTTP(c.Writer, c.Request)
}

// GetRouteMetadataManagementPage renders the Route Metadata Management page
// Allows admins to view and manage enterprise route metadata configuration
func (h *performanceHandler) GetRouteMetadataManagementPage(c *gin.Context) {
	// Load current route metadata
	routeMetadata, err := enterprise.LoadEnterpriseRouteMetadata("")
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load route metadata")
		return
	}

	// Create management data structure
	managementData := templates.RouteMetadataManagementPageData{
		Routes: routeMetadata,
	}

	// Render Templ template
	templ.Handler(templates.RouteMetadataManagementPage(managementData)).ServeHTTP(c.Writer, c.Request)
}
func (h *performanceHandler) LoginJSON(c *gin.Context) {
	loginRequest, err := helperImpl.GetJSONDataFromRequest[dto.LoginRequest](c)
	if err != nil {
		// Enhanced error message for debugging
		log.Printf("JSON binding error: %v", err)
		h.responseHelper.BadRequest(c, utils.ErrBadRequest.Error(), "Invalid request format. Expected JSON with 'username' and 'password' fields.")
		return
	}
	// Validate request
	if loginRequest == nil {
		h.responseHelper.BadRequest(c, utils.ErrBadRequest.Error(), "Request body cannot be empty")
		return
	}
	if loginRequest.Username == "" || loginRequest.Password == "" {
		h.responseHelper.BadRequest(c, utils.ErrBadRequest.Error(), "Username and password are required")
		return
	}
	// Perform authentication
	response, err := h.authService.Login(loginRequest)
	if err != nil {
		log.Printf("Authentication service error: %v", err)
		h.responseHelper.Unauthorized(c, "Admin credentials not configured. Please contact system administrator.")
		return
	}
	if !response.Success {
		log.Printf("Authentication failed for user: %s", loginRequest.Username)
		h.responseHelper.Unauthorized(c, response.Message)
		return
	}

	// Generate JWT token for API requests
	jwtToken := h.generateJWTToken(loginRequest.Username, "admin")

	// Set session cookie
	c.SetCookie(
		"admin_session",
		response.SessionID,
		3600*24, // 24 hours
		"/admin-ui",
		"",
		false,
		true,
	)

	// Set JWT token cookie for API requests
	c.SetCookie(
		"jwt_token",
		jwtToken,
		3600*24, // 24 hours
		"/",
		"",
		false,
		false, // Not HTTP-only to allow JavaScript access
	)

	// Add JWT token to response
	response.JWT = jwtToken

	// Return success response
	c.JSON(http.StatusOK, response)
}

func (h *performanceHandler) Logout(c *gin.Context) {
	// Get session ID from cookie
	sessionID, err := c.Cookie("admin_session")
	if err != nil {
		c.Redirect(http.StatusFound, "/admin-ui/login")
		return
	}

	// Logout
	h.authService.Logout(sessionID)

	// Clear session cookie
	c.SetCookie(
		"admin_session",
		"",
		-1,
		"/admin-ui",
		"",
		false,
		true,
	)

	// Clear JWT token cookie
	c.SetCookie(
		"jwt_token",
		"",
		-1,
		"/",
		"",
		false,
		false,
	)

	// Return HTML with script to clear localStorage and redirect
	logoutHTML := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Logging out...</title>
	</head>
	<body>
		<script>
			// Clear JWT token from localStorage
			localStorage.removeItem('jwt_token');
			// Redirect to login page
			window.location.href = '/admin-ui/login';
		</script>
	</body>
	</html>
	`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, logoutHTML)
}
func (h *performanceHandler) generateJWTToken(username, role string) string {

	claims := jwt.MapClaims{
		"username": username,
		"role":     role,
		"user_id":  "admin_" + username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
		"iat":      time.Now().Unix(),
	}
	tokenString, err := service.GenerateToken(claims)
	if err != nil {
		log.Println("Error generating JWT token:", err)
		return ""
	}

	return tokenString
}

// GetRoleManagementPage renders the Role Management page
// Allows admins to view roles and user role assignments from Casbin
func (h *performanceHandler) GetRoleManagementPage(c *gin.Context) {
	// Get all roles from Casbin
	roleNames, err := h.profileService.GetAllRolesFromCasbin()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load roles from Casbin")
		return
	}

	// Get role descriptions
	roleDescriptions := h.profileService.GetRoleDescriptions()

	// Build role info with descriptions and user counts
	allRoles := make([]templates.RoleInfo, 0, len(roleNames))
	userRoleMap := make(map[string][]string) // userID -> roles

	for _, roleName := range roleNames {
		description := roleDescriptions[roleName]
		if description == "" {
			description = roleName + " role" // Fallback description
		}

		// Get users for this role
		users, err := h.profileService.GetUsersForRole(roleName)
		userCount := 0
		if err == nil {
			userCount = len(users)
			// Collect user -> role mappings
			for _, userID := range users {
				userRoleMap[userID] = append(userRoleMap[userID], roleName)
			}
		}

		allRoles = append(allRoles, templates.RoleInfo{
			Name:        roleName,
			Description: description,
			UserCount:   userCount,
		})
	}

	// Build user role assignments from collected data
	userRoles := make([]templates.UserRoleAssignment, 0, len(userRoleMap))
	for userID, roles := range userRoleMap {
		// Use userID as username for now (in a real app, you'd have a user service)
		username := userID
		if userID == "admin" || strings.HasPrefix(userID, "admin") {
			username = "Administrator"
		}

		userRoles = append(userRoles, templates.UserRoleAssignment{
			UserID:   userID,
			Username: username,
			Roles:    roles,
		})
	}

	// Create management data structure
	managementData := templates.RoleManagementPageData{
		Roles:     allRoles,
		UserRoles: userRoles,
	}

	// Render Templ template
	templ.Handler(templates.RoleManagementPage(managementData)).ServeHTTP(c.Writer, c.Request)
}

// GetPolicyManagementPage renders the Policy Management documentation page
// Provides comprehensive guidance on managing Casbin policies following DDD principles
func (h *performanceHandler) GetPolicyManagementPage(c *gin.Context) {
	// Get current policy statistics
	policyCount := 0
	groupingCount := 0

	if enforcer := initializer.CasbinEnforcer; enforcer != nil {
		if policies, err := enforcer.GetPolicy(); err == nil {
			policyCount = len(policies)
		}
		if groupingPolicies, err := enforcer.GetGroupingPolicy(); err == nil {
			groupingCount = len(groupingPolicies)
		}
	}

	// Get roles for examples
	roles, _ := h.profileService.GetAllRolesFromCasbin()
	if len(roles) == 0 {
		roles = []string{"admin", "staff", "user"}
	}

	// Create policy management data
	policyData := templates.PolicyManagementPageData{
		PolicyCount:       policyCount,
		GroupingCount:     groupingCount,
		AvailableRoles:    roles,
		CurrentPolicyFile: "config/casbin_rbac_policy.csv",
		CurrentModelFile:  "config/casbin_rbac_model.conf",
	}

	// Render Templ template
	templ.Handler(templates.PolicyManagementPage(policyData)).ServeHTTP(c.Writer, c.Request)
}

// GetAuditLogsPage renders the Authorization Audit Logs page
// Shows comprehensive audit trail of authorization decisions
func (h *performanceHandler) GetAuditLogsPage(c *gin.Context) {
	// Lazy initialization of audit service
	if h.auditService == nil && enterprise.EnterpriseAuth != nil {
		auditRepo := enterprise.EnterpriseAuth.GetAuditRepository()
		if auditRepo != nil {
			h.auditService = service.NewAuthorizationAuditService(auditRepo)
		}
	}

	if h.auditService == nil {
		c.String(http.StatusServiceUnavailable, "Audit service not available")
		return
	}

	// Get query parameters for filtering
	limit := 50 // default limit
	offset := 0
	userID := c.Query("user_id")
	result := c.Query("result")
	resource := c.Query("resource")

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var auditLogs *[]service.AuditLogDTO
	var err error

	// Apply filters
	if userID != "" {
		auditLogs, err = h.auditService.GetAuditLogsByUser(userID, limit, offset)
	} else if result != "" {
		auditLogs, err = h.auditService.GetAuditLogsByResult(result, limit, offset)
	} else if resource != "" {
		auditLogs, err = h.auditService.GetAuditLogsByResource(resource, limit, offset)
	} else {
		auditLogs, err = h.auditService.GetAuditLogs(limit, offset)
	}

	if err != nil {
		logger.Error("Failed to get audit logs", zap.Error(err))
		c.String(http.StatusInternalServerError, "Failed to load audit logs")
		return
	}

	if auditLogs == nil {
		auditLogs = &[]service.AuditLogDTO{}
	}

	// Get audit summary
	summary, err := h.auditService.GetAuditSummary()
	if err != nil {
		logger.Warn("Failed to get audit summary", zap.Error(err))
		summary = &service.AuditSummaryDTO{}
	}

	// Create audit logs page data
	auditData := templates.AuditLogsPageData{
		AuditLogs: *auditLogs,
		Summary:   *summary,
		CurrentFilter: map[string]string{
			"user_id":  userID,
			"result":   result,
			"resource": resource,
		},
		Limit:  limit,
		Offset: offset,
	}

	// Render Templ template
	templ.Handler(templates.AuditLogsPage(auditData)).ServeHTTP(c.Writer, c.Request)
}

// GetFeaturesDocumentationPage renders the comprehensive features documentation page
// Shows all framework capabilities, architecture, and integration examples
func (h *performanceHandler) GetFeaturesDocumentationPage(c *gin.Context) {
	// Get route count
	routeMetadata, err := enterprise.LoadEnterpriseRouteMetadata("")
	totalEndpoints := 0
	if err == nil && routeMetadata != nil {
		totalEndpoints = len(routeMetadata)
	}

	// Create features page data
	featuresData := templates.FeaturesDocumentationPageData{
		TotalFeatures:  15, // Core features count
		TotalEndpoints: totalEndpoints,
	}

	// Render Templ template
	templ.Handler(templates.FeaturesDocumentationPage(featuresData)).ServeHTTP(c.Writer, c.Request)
}

// CreateRole creates a new role
func (h *performanceHandler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.profileService.CreateRole(req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Role created successfully", "role": req.Name})
}

// UpdateRole updates an existing role's name and/or description
func (h *performanceHandler) UpdateRole(c *gin.Context) {
	var req struct {
		OldName     string `json:"old_name" binding:"required"`
		NewName     string `json:"new_name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If new name is empty, use old name (only updating description)
	if req.NewName == "" {
		req.NewName = req.OldName
	}

	err := h.profileService.UpdateRole(req.OldName, req.NewName, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully", "role": req.NewName})
}

// AssignRoleToUser assigns a role to a user
func (h *performanceHandler) AssignRoleToUser(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.profileService.AssignRoleToUser(req.UserID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role assigned successfully"})
}

// RemoveRoleFromUser removes a role from a user
func (h *performanceHandler) RemoveRoleFromUser(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.profileService.RemoveRoleFromUser(req.UserID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role removed successfully"})
}

// GetUsersForRole returns users assigned to a specific role
func (h *performanceHandler) GetUsersForRole(c *gin.Context) {
	role := c.Query("role")
	if role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role parameter is required"})
		return
	}

	users, err := h.profileService.GetUsersForRole(role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"role": role, "users": users})
}

// DeleteRole deletes a role and all its assignments
func (h *performanceHandler) DeleteRole(c *gin.Context) {
	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.profileService.DeleteRole(req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

// GetRoleDetailsPage renders the detailed view of a specific role
// Shows users, permissions, and role information
func (h *performanceHandler) GetRoleDetailsPage(c *gin.Context) {
	roleName := c.Param("role")
	if roleName == "" {
		c.String(http.StatusBadRequest, "Role name is required")
		return
	}

	// Get role details from service
	roleDetails, err := h.profileService.GetRoleDetails(roleName)
	if err != nil {
		logger.Error("Failed to get role details", zap.Error(err), zap.String("role", roleName))
		c.String(http.StatusInternalServerError, "Failed to load role details")
		return
	}

	// Extract data from map
	description := ""
	if desc, ok := roleDetails["description"].(string); ok {
		description = desc
	}

	userCount := 0
	if count, ok := roleDetails["user_count"].(int); ok {
		userCount = count
	}

	users := []string{}
	if userList, ok := roleDetails["users"].([]string); ok {
		users = userList
	}

	permissions := []map[string]string{}
	if permList, ok := roleDetails["permissions"].([]map[string]string); ok {
		permissions = permList
	}

	// Create page data
	pageData := templates.RoleDetailsPageData{
		RoleName:    roleName,
		Description: description,
		UserCount:   userCount,
		Users:       users,
		Permissions: permissions,
	}

	// Render Templ template
	templ.Handler(templates.RoleDetailsPage(pageData)).ServeHTTP(c.Writer, c.Request)
}

// SaveRouteMetadata handles saving updated route metadata
func (h *performanceHandler) SaveRouteMetadata(c *gin.Context) {
	// Parse the JSON payload
	var updateRequest struct {
		Routes []*enterprise.RouteMetadata `json:"routes"`
	}

	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// Validate all routes
	for _, route := range updateRequest.Routes {
		if err := route.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid route %s %s: %v", route.Method, route.Path, err)})
			return
		}
	}

	// Save to file
	if err := enterprise.SaveEnterpriseRouteMetadata(updateRequest.Routes, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save route metadata"})
		return
	}

	// Update Casbin policies based on the new route metadata
	if err := enterprise.UpdateCasbinPoliciesFromRoutes(updateRequest.Routes, ""); err != nil {
		logger.GetLogger().Warn("Failed to update Casbin policies after route metadata save", zap.Error(err))
		// Don't fail the request, just log the warning
	}

	c.JSON(http.StatusOK, gin.H{"message": "Route metadata saved successfully"})
}

// ImportRouteMetadata handles importing route metadata from JSON
func (h *performanceHandler) ImportRouteMetadata(c *gin.Context) {
	var importRequest struct {
		Routes []*enterprise.RouteMetadata `json:"routes"`
	}

	if err := c.ShouldBindJSON(&importRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	if len(importRequest.Routes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Routes list cannot be empty"})
		return
	}

	// Validate all routes
	for _, route := range importRequest.Routes {
		if err := route.Validate(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid route %s %s: %v", route.Method, route.Path, err)})
			return
		}
	}

	// Load existing routes
	existingRoutes, _ := enterprise.LoadEnterpriseRouteMetadata("")
	if existingRoutes == nil {
		existingRoutes = make([]*enterprise.RouteMetadata, 0)
	}

	// Create a map of existing routes by method:path for easy lookup
	existingMap := make(map[string]*enterprise.RouteMetadata)
	for _, route := range existingRoutes {
		key := fmt.Sprintf("%s:%s", route.Method, route.Path)
		existingMap[key] = route
	}

	// Merge imported routes with existing ones (imported routes override existing)
	mergedRoutes := make([]*enterprise.RouteMetadata, 0, len(existingRoutes))
	importedKeys := make(map[string]bool)

	// Add imported routes first
	for _, route := range importRequest.Routes {
		key := fmt.Sprintf("%s:%s", route.Method, route.Path)
		mergedRoutes = append(mergedRoutes, route)
		importedKeys[key] = true
	}

	// Add existing routes that weren't imported
	for key, route := range existingMap {
		if !importedKeys[key] {
			mergedRoutes = append(mergedRoutes, route)
		}
	}

	// Save merged routes to file
	if err := enterprise.SaveEnterpriseRouteMetadata(mergedRoutes, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save imported routes"})
		return
	}

	// Update Casbin policies based on the merged route metadata
	if err := enterprise.UpdateCasbinPoliciesFromRoutes(mergedRoutes, ""); err != nil {
		logger.GetLogger().Warn("Failed to update Casbin policies after route import", zap.Error(err))
		// Don't fail the request, just log the warning
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Route metadata imported successfully",
		"imported": len(importRequest.Routes),
		"total": len(mergedRoutes),
	})
}

// DeleteRouteMetadata handles deleting a specific route
func (h *performanceHandler) DeleteRouteMetadata(c *gin.Context) {
	var deleteRequest struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	}

	if err := c.ShouldBindJSON(&deleteRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	if deleteRequest.Method == "" || deleteRequest.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Method and path are required"})
		return
	}

	// Load all routes
	routes, err := enterprise.LoadEnterpriseRouteMetadata("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load route metadata"})
		return
	}

	// Find and remove the route
	var updatedRoutes []*enterprise.RouteMetadata
	found := false
	for _, route := range routes {
		if route.Method == deleteRequest.Method && route.Path == deleteRequest.Path {
			found = true
			continue // Skip this route (delete it)
		}
		updatedRoutes = append(updatedRoutes, route)
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
		return
	}

	// Save updated routes
	if err := enterprise.SaveEnterpriseRouteMetadata(updatedRoutes, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete route"})
		return
	}

	// Update Casbin policies based on the updated route metadata
	if err := enterprise.UpdateCasbinPoliciesFromRoutes(updatedRoutes, ""); err != nil {
		logger.GetLogger().Warn("Failed to update Casbin policies after route deletion", zap.Error(err))
		// Don't fail the request, just log the warning
	}

	c.JSON(http.StatusOK, gin.H{"message": "Route deleted successfully"})
}
