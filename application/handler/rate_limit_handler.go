package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitStats holds statistics for a rate limiter
type RateLimitStats struct {
	IP              string
	RequestCount    int64
	ResetTime       int64
	LimitPerSecond  float64
	BurstSize       int
	CurrentAllowed  int64
	LastRequestTime time.Time
	IsBlocked       bool
}

// RateLimitManager manages rate limiting statistics and configuration
type RateLimitManager struct {
	mu             sync.RWMutex
	limiters       map[string]*rate.Limiter
	stats          map[string]*RateLimitStats
	globalRate     rate.Limit
	globalBurst    int
	endpointLimits map[string]rate.Limit
	endpointBursts map[string]int
}

// NewRateLimitManager creates a new rate limit manager
func NewRateLimitManager(globalRate rate.Limit, globalBurst int) *RateLimitManager {
	return &RateLimitManager{
		limiters:       make(map[string]*rate.Limiter),
		stats:          make(map[string]*RateLimitStats),
		globalRate:     globalRate,
		globalBurst:    globalBurst,
		endpointLimits: make(map[string]rate.Limit),
		endpointBursts: make(map[string]int),
	}
}

// GetStats returns statistics for an IP
func (m *RateLimitManager) GetStats(ip string) *RateLimitStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if stats, exists := m.stats[ip]; exists {
		return stats
	}

	return &RateLimitStats{
		IP:             ip,
		LimitPerSecond: float64(m.globalRate),
		BurstSize:      m.globalBurst,
	}
}

// GetAllStats returns statistics for all IPs
func (m *RateLimitManager) GetAllStats() []*RateLimitStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make([]*RateLimitStats, 0, len(m.stats))
	for _, stat := range m.stats {
		stats = append(stats, stat)
	}
	return stats
}

// UpdateStats updates statistics for an IP
func (m *RateLimitManager) UpdateStats(ip string, allowed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.stats[ip]; !exists {
		m.stats[ip] = &RateLimitStats{
			IP:             ip,
			LimitPerSecond: float64(m.globalRate),
			BurstSize:      m.globalBurst,
		}
	}

	stats := m.stats[ip]
	stats.LastRequestTime = time.Now()

	if allowed {
		stats.CurrentAllowed++
	} else {
		stats.IsBlocked = true
	}

	stats.RequestCount++
}

// ResetIP resets statistics for an IP
func (m *RateLimitManager) ResetIP(ip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if stats, exists := m.stats[ip]; exists {
		stats.RequestCount = 0
		stats.CurrentAllowed = 0
		stats.IsBlocked = false
		stats.LastRequestTime = time.Time{}
		return nil
	}

	return fmt.Errorf("IP %s not found", ip)
}

// SetEndpointLimit sets a custom rate limit for an endpoint
func (m *RateLimitManager) SetEndpointLimit(endpoint string, limit rate.Limit, burst int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if limit <= 0 {
		return fmt.Errorf("invalid rate limit: %v", limit)
	}
	if burst < 1 {
		return fmt.Errorf("invalid burst size: %d", burst)
	}

	m.endpointLimits[endpoint] = limit
	m.endpointBursts[endpoint] = burst
	return nil
}

// GetEndpointLimit gets the rate limit for an endpoint
func (m *RateLimitManager) GetEndpointLimit(endpoint string) (rate.Limit, int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit, exists := m.endpointLimits[endpoint]; exists {
		return limit, m.endpointBursts[endpoint]
	}

	return m.globalRate, m.globalBurst
}

// GetAllEndpointLimits returns all endpoint limits
func (m *RateLimitManager) GetAllEndpointLimits() map[string]map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]map[string]interface{})
	for endpoint, limit := range m.endpointLimits {
		result[endpoint] = map[string]interface{}{
			"limit": float64(limit),
			"burst": m.endpointBursts[endpoint],
		}
	}
	return result
}

// UpdateGlobalLimit updates the global rate limit
func (m *RateLimitManager) UpdateGlobalLimit(limit rate.Limit, burst int) error {
	if limit <= 0 {
		return fmt.Errorf("invalid rate limit: %v", limit)
	}
	if burst < 1 {
		return fmt.Errorf("invalid burst size: %d", burst)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.globalRate = limit
	m.globalBurst = burst
	return nil
}

// GetGlobalLimit returns the current global rate limit
func (m *RateLimitManager) GetGlobalLimit() (rate.Limit, int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.globalRate, m.globalBurst
}

// RateLimitHandler handles rate limit UI requests
type RateLimitHandler struct {
	manager *RateLimitManager
}

// NewRateLimitHandler creates a new rate limit handler
func NewRateLimitHandler(manager *RateLimitManager) *RateLimitHandler {
	return &RateLimitHandler{manager: manager}
}

// GetRateLimitPage returns the rate limiting UI page
func (h *RateLimitHandler) GetRateLimitPage(c *gin.Context) {
	globalLimit, globalBurst := h.manager.GetGlobalLimit()
	stats := h.manager.GetAllStats()

	data := map[string]interface{}{
		"globalLimit":   float64(globalLimit),
		"globalBurst":   globalBurst,
		"stats":         stats,
		"statsCount":    len(stats),
		"blockedCount":  countBlocked(stats),
		"totalRequests": sumRequests(stats),
	}

	c.JSON(http.StatusOK, data)
}

// GetRateLimitStats returns JSON statistics
func (h *RateLimitHandler) GetRateLimitStats(c *gin.Context) {
	stats := h.manager.GetAllStats()

	c.JSON(http.StatusOK, gin.H{
		"stats":         stats,
		"totalIPs":      len(stats),
		"blockedIPs":    countBlocked(stats),
		"totalRequests": sumRequests(stats),
	})
}

// GetIPStats returns statistics for a specific IP
func (h *RateLimitHandler) GetIPStats(c *gin.Context) {
	ip := c.Param("ip")
	stats := h.manager.GetStats(ip)

	c.JSON(http.StatusOK, stats)
}

// ResetIPLimit resets the rate limit for an IP
func (h *RateLimitHandler) ResetIPLimit(c *gin.Context) {
	ip := c.Param("ip")

	err := h.manager.ResetIP(ip)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Rate limit reset for IP %s", ip),
		"ip":      ip,
	})
}

// UpdateGlobalLimit updates the global rate limit
func (h *RateLimitHandler) UpdateGlobalLimit(c *gin.Context) {
	var req struct {
		Limit float64 `json:"limit" binding:"required,gt=0"`
		Burst int     `json:"burst" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limitValue := rate.Limit(req.Limit)
	err := h.manager.UpdateGlobalLimit(limitValue, req.Burst)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Global rate limit updated",
		"limit":   req.Limit,
		"burst":   req.Burst,
	})
}

// SetEndpointLimit sets a rate limit for an endpoint
func (h *RateLimitHandler) SetEndpointLimit(c *gin.Context) {
	var req struct {
		Endpoint string  `json:"endpoint" binding:"required"`
		Limit    float64 `json:"limit" binding:"required,gt=0"`
		Burst    int     `json:"burst" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limitValue := rate.Limit(req.Limit)
	err := h.manager.SetEndpointLimit(req.Endpoint, limitValue, req.Burst)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Endpoint rate limit set",
		"endpoint": req.Endpoint,
		"limit":    req.Limit,
		"burst":    req.Burst,
	})
}

// GetEndpointLimits returns all endpoint limits
func (h *RateLimitHandler) GetEndpointLimits(c *gin.Context) {
	limits := h.manager.GetAllEndpointLimits()

	c.JSON(http.StatusOK, gin.H{
		"limits": limits,
		"count":  len(limits),
	})
}

// ResetAllLimits resets all rate limits
func (h *RateLimitHandler) ResetAllLimits(c *gin.Context) {
	stats := h.manager.GetAllStats()

	for _, stat := range stats {
		h.manager.ResetIP(stat.IP)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "All rate limits have been reset",
		"ipsReset": len(stats),
	})
}

// Helper functions

func countBlocked(stats []*RateLimitStats) int {
	count := 0
	for _, stat := range stats {
		if stat.IsBlocked {
			count++
		}
	}
	return count
}

func sumRequests(stats []*RateLimitStats) int64 {
	var total int64
	for _, stat := range stats {
		total += stat.RequestCount
	}
	return total
}

// SearchRateLimitStats searches rate limit statistics
func (h *RateLimitHandler) SearchRateLimitStats(c *gin.Context) {
	query := c.Query("q")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	allStats := h.manager.GetAllStats()
	filtered := make([]*RateLimitStats, 0)

	for _, stat := range allStats {
		if query == "" || contains(stat.IP, query) {
			filtered = append(filtered, stat)
		}
	}

	// Pagination
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	if offset >= len(filtered) {
		filtered = make([]*RateLimitStats, 0)
	} else {
		filtered = filtered[offset:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":    filtered,
		"total":    len(allStats),
		"limit":    limit,
		"offset":   offset,
		"hasMore":  (offset + limit) < len(allStats),
		"filtered": len(filtered),
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && s[len(s)-len(substr):] == substr
}

// ExportRateLimitStats exports statistics as CSV
func (h *RateLimitHandler) ExportRateLimitStats(c *gin.Context) {
	stats := h.manager.GetAllStats()

	csv := "IP,RequestCount,BlockedStatus,LastRequestTime,LimitPerSecond,BurstSize\n"
	for _, stat := range stats {
		blocked := "No"
		if stat.IsBlocked {
			blocked = "Yes"
		}

		csv += fmt.Sprintf("%s,%d,%s,%s,%.2f,%d\n",
			stat.IP,
			stat.RequestCount,
			blocked,
			stat.LastRequestTime.Format(time.RFC3339),
			stat.LimitPerSecond,
			stat.BurstSize,
		)
	}

	c.Header("Content-Disposition", "attachment; filename=rate_limit_stats.csv")
	c.Header("Content-Type", "text/csv")
	c.String(http.StatusOK, csv)
}
