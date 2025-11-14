package handler

import (
	"github.com/gin-gonic/gin"
)

// SetupRateLimitingRoutes sets up all rate limiting related routes
func SetupRateLimitingRoutes(router *gin.Engine, rateLimitHandler *RateLimitHandler) {
	// Pages
	router.GET("/admin-ui/rate-limiting", rateLimitHandler.GetRateLimitingPage)

	// API endpoints
	api := router.Group("/admin-ui/api/rate-limit")
	{
		// Stats
		api.GET("/stats", rateLimitHandler.GetRateLimitStats)
		api.GET("/stats/:ip", rateLimitHandler.GetIPStats)
		api.GET("/search", rateLimitHandler.SearchRateLimitStats)
		api.GET("/export", rateLimitHandler.ExportRateLimitStats)

		// Reset
		api.POST("/reset/:ip", rateLimitHandler.ResetIPLimit)
		api.POST("/reset-all", rateLimitHandler.ResetAllLimits)

		// Configuration
		api.POST("/update-global", rateLimitHandler.UpdateGlobalLimit)
		api.POST("/set-endpoint", rateLimitHandler.SetEndpointLimit)
		api.GET("/endpoints", rateLimitHandler.GetEndpointLimits)
	}
}

// GetRateLimitingPage returns the rate limiting management page
func (h *RateLimitHandler) GetRateLimitingPage(c *gin.Context) {
	globalLimit, globalBurst := h.manager.GetGlobalLimit()

	// For JSON response
	c.JSON(200, gin.H{
		"globalLimit": float64(globalLimit),
		"globalBurst": globalBurst,
		"pageTitle":   "Rate Limiting Management",
	})
}
