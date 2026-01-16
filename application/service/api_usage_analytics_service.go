package service

import (
	"fmt"
	"sort"
	"time"

	"github.com/aruncs31s/azf/domain/api_usage"
	"github.com/aruncs31s/azf/domain/repository"
	"github.com/aruncs31s/azf/shared/logger"
	"go.uber.org/zap"
)

// APIUsageAnalyticsService provides analytics and insights for API usage
type APIUsageAnalyticsService interface {
	GetTopEndpointsByUsage(limit int) (*[]api_usage.APIEndpointRanking, error)
	GetEndpointsByErrorRate(limit int) (*[]api_usage.APIEndpointRanking, error)
	GetEndpointsByResponseTime(limit int) (*[]api_usage.APIEndpointRanking, error)
	GetEndpointDetails(endpoint string) (*EndpointDetailsDTO, error)
	GetEndpointCallers(endpoint string, limit int) (*[]CallerDTO, error)
	GetUsageSummary() (*UsageSummaryDTO, error)
	GetUsageTrend(days int) (*[]UsageTrendDTO, error)
	GetUserActivitySummary(userID string) (*UserActivityDTO, error)
	RecalculateAllStats() error
	ClearAllStatistics() error
}

// apiUsageAnalyticsService implements APIUsageAnalyticsService
type apiUsageAnalyticsService struct {
	logRepo   repository.APIUsageLogRepository
	statsRepo repository.APIUsageStatsRepository
}

// NewAPIUsageAnalyticsService creates a new API usage analytics service
func NewAPIUsageAnalyticsService(
	logRepo repository.APIUsageLogRepository,
	statsRepo repository.APIUsageStatsRepository,
) APIUsageAnalyticsService {
	return &apiUsageAnalyticsService{
		logRepo:   logRepo,
		statsRepo: statsRepo,
	}
}

// GetTopEndpointsByUsage returns top endpoints sorted by usage count
func (s *apiUsageAnalyticsService) GetTopEndpointsByUsage(limit int) (*[]api_usage.APIEndpointRanking, error) {
	rankings, err := s.statsRepo.GetTopEndpointsByUsage(limit)
	if err != nil {
		logger.GetLogger().Error("Failed to get top endpoints by usage", zap.Error(err))
		return nil, err
	}
	return rankings, nil
}

// GetEndpointsByErrorRate returns endpoints sorted by error rate
func (s *apiUsageAnalyticsService) GetEndpointsByErrorRate(limit int) (*[]api_usage.APIEndpointRanking, error) {
	rankings, err := s.statsRepo.GetEndpointsByErrorRate(limit)
	if err != nil {
		logger.GetLogger().Error("Failed to get endpoints by error rate", zap.Error(err))
		return nil, err
	}
	return rankings, nil
}

// GetEndpointsByResponseTime returns endpoints sorted by response time
func (s *apiUsageAnalyticsService) GetEndpointsByResponseTime(limit int) (*[]api_usage.APIEndpointRanking, error) {
	rankings, err := s.statsRepo.GetEndpointsByResponseTime(limit)
	if err != nil {
		logger.GetLogger().Error(
			"Failed to get endpoints by response time",
			zap.Error(err),
		)
		return nil, err
	}
	return rankings, nil
}

// GetEndpointDetails returns detailed statistics for a specific endpoint
func (s *apiUsageAnalyticsService) GetEndpointDetails(endpoint string) (*EndpointDetailsDTO, error) {
	stats, err := s.statsRepo.FindByEndpoint(endpoint)
	if err != nil {
		logger.Error("Failed to get endpoint stats", zap.Error(err), zap.String("endpoint", endpoint))
		return nil, err
	}

	if stats == nil {
		return nil, nil
	}

	// Calculate error rate and success rate
	errorRate := 0.0
	successRate := 100.0
	if stats.TotalRequests > 0 {
		errorRate = (float64(stats.ErrorRequests) / float64(stats.TotalRequests)) * 100
		successRate = (float64(stats.SuccessRequests) / float64(stats.TotalRequests)) * 100
	}

	// Get recent logs for this endpoint
	logs, err := s.logRepo.FindByEndpoint(endpoint, 100, 0)
	if err != nil {
		logger.Warn("Failed to get endpoint logs", zap.Error(err), zap.String("endpoint", endpoint))
	}

	return &EndpointDetailsDTO{
		Endpoint:        stats.Endpoint,
		Method:          stats.Method,
		TotalRequests:   stats.TotalRequests,
		SuccessRequests: stats.SuccessRequests,
		ErrorRequests:   stats.ErrorRequests,
		SuccessRate:     successRate,
		ErrorRate:       errorRate,
		AvgResponseTime: stats.AvgResponseTime,
		MinResponseTime: stats.MinResponseTime,
		MaxResponseTime: stats.MaxResponseTime,
		Last24Hours:     stats.Last24Hours,
		LastAccessedAt:  stats.LastAccessedAt,
		RecentLogs:      logs,
	}, nil
}

// GetUsageSummary returns overall API usage summary
func (s *apiUsageAnalyticsService) GetUsageSummary() (*UsageSummaryDTO, error) {
	totalLogs, err := s.logRepo.CountTotal()
	if err != nil {
		logger.Error("Failed to count total logs", zap.Error(err))
		return nil, err
	}

	allStats, err := s.statsRepo.FindAll(1000, 0)
	if err != nil {
		logger.Error("Failed to get all stats", zap.Error(err))
		return nil, err
	}

	summary := &UsageSummaryDTO{
		TotalRequests: totalLogs,
		Timestamp:     time.Now(),
	}

	if allStats == nil || len(*allStats) == 0 {
		return summary, nil
	}

	var totalSuccess int64
	var totalErrors int64
	var totalResponseTime int64
	var maxResponseTime int64
	var minResponseTime int64 = 999999999
	endpointCount := len(*allStats)

	for _, stat := range *allStats {
		totalSuccess += stat.SuccessRequests
		totalErrors += stat.ErrorRequests
		totalResponseTime += stat.AvgResponseTime
		if stat.MaxResponseTime > maxResponseTime {
			maxResponseTime = stat.MaxResponseTime
		}
		if stat.MinResponseTime < minResponseTime && stat.MinResponseTime > 0 {
			minResponseTime = stat.MinResponseTime
		}
	}

	summary.SuccessfulRequests = totalSuccess
	summary.FailedRequests = totalErrors
	summary.UniqueEndpoints = int64(endpointCount)

	if endpointCount > 0 {
		summary.AvgResponseTime = totalResponseTime / int64(endpointCount)
		summary.MaxResponseTime = maxResponseTime
		if minResponseTime < 999999999 {
			summary.MinResponseTime = minResponseTime
		}
	}

	if totalLogs > 0 {
		summary.SuccessRate = (float64(totalSuccess) / float64(totalLogs)) * 100
		summary.ErrorRate = (float64(totalErrors) / float64(totalLogs)) * 100
	}

	return summary, nil
}

// GetUsageTrend returns usage trend over specified days
func (s *apiUsageAnalyticsService) GetUsageTrend(days int) (*[]UsageTrendDTO, error) {
	trends := make([]UsageTrendDTO, 0, days)

	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		logs, err := s.logRepo.FindByDateRange(
			startOfDay.Format(time.RFC3339),
			endOfDay.Format(time.RFC3339),
			10000,
			0,
		)
		if err != nil {
			logger.Warn("Failed to get logs for trend", zap.Error(err), zap.Time("date", date))
			continue
		}

		if logs == nil || len(*logs) == 0 {
			trends = append(trends, UsageTrendDTO{
				Date:            startOfDay,
				RequestCount:    0,
				SuccessCount:    0,
				ErrorCount:      0,
				AvgResponseTime: 0,
			})
			continue
		}

		var successCount, errorCount int64
		var totalResponseTime int64

		for _, log := range *logs {
			if log.StatusCode >= 200 && log.StatusCode < 300 {
				successCount++
			} else {
				errorCount++
			}
			totalResponseTime += log.ResponseTime
		}

		avgResponseTime := int64(0)
		if len(*logs) > 0 {
			avgResponseTime = totalResponseTime / int64(len(*logs))
		}

		trends = append(trends, UsageTrendDTO{
			Date:            startOfDay,
			RequestCount:    int64(len(*logs)),
			SuccessCount:    successCount,
			ErrorCount:      errorCount,
			AvgResponseTime: avgResponseTime,
		})
	}

	return &trends, nil
}

// GetUserActivitySummary returns activity summary for a specific user
func (s *apiUsageAnalyticsService) GetUserActivitySummary(userID string) (*UserActivityDTO, error) {
	logs, err := s.logRepo.FindByUserID(userID, 10000, 0)
	if err != nil {
		logger.Error("Failed to get user logs", zap.Error(err), zap.String("user_id", userID))
		return nil, err
	}

	summary := &UserActivityDTO{
		UserID: userID,
	}

	if logs == nil || len(*logs) == 0 {
		return summary, nil
	}

	summary.TotalRequests = int64(len(*logs))

	endpointMap := make(map[string]int64)
	var totalResponseTime int64
	var successCount, errorCount int64

	for _, log := range *logs {
		endpointMap[log.Endpoint]++
		totalResponseTime += log.ResponseTime
		if log.StatusCode >= 200 && log.StatusCode < 300 {
			successCount++
		} else {
			errorCount++
		}
	}

	summary.SuccessfulRequests = successCount
	summary.FailedRequests = errorCount
	summary.AvgResponseTime = totalResponseTime / summary.TotalRequests
	summary.UniqueEndpointsAccessed = int64(len(endpointMap))
	summary.MostUsedEndpoint = getMostUsedEndpoint(endpointMap)

	return summary, nil
}

// RecalculateAllStats recalculates all endpoint statistics
func (s *apiUsageAnalyticsService) RecalculateAllStats() error {
	// Get all unique endpoint+method combinations from logs
	allStats, err := s.statsRepo.FindAll(10000, 0)
	if err != nil {
		logger.Error("Failed to get all stats for recalculation", zap.Error(err))
		return err
	}

	if allStats == nil || len(*allStats) == 0 {
		logger.Info("No stats to recalculate")
		return nil
	}

	for _, stat := range *allStats {
		if err := s.statsRepo.RecalculateStats(stat.Endpoint, stat.Method); err != nil {
			logger.Error("Failed to recalculate stats",
				zap.String("endpoint", stat.Endpoint),
				zap.String("method", stat.Method),
				zap.Error(err),
			)
		}
	}

	logger.Info("Recalculated stats for all endpoints")
	return nil
}

// ClearAllStatistics clears all API usage statistics and logs
func (s *apiUsageAnalyticsService) ClearAllStatistics() error {
	// Clear all logs
	if err := s.logRepo.DeleteAll(); err != nil {
		logger.Error("Failed to clear API usage logs", zap.Error(err))
		return fmt.Errorf("failed to clear logs: %w", err)
	}

	// Clear all statistics
	if err := s.statsRepo.DeleteAll(); err != nil {
		logger.Error("Failed to clear API usage statistics", zap.Error(err))
		return fmt.Errorf("failed to clear statistics: %w", err)
	}

	return nil
}

// GetEndpointCallers retrieves the users who called a specific endpoint
func (s *apiUsageAnalyticsService) GetEndpointCallers(endpoint string, limit int) (*[]CallerDTO, error) {
	logs, err := s.logRepo.FindByEndpoint(endpoint, 10000, 0)
	if err != nil {
		logger.Error("Failed to get endpoint logs for callers", zap.Error(err), zap.String("endpoint", endpoint))
		return nil, err
	}

	if logs == nil || len(*logs) == 0 {
		return &[]CallerDTO{}, nil
	}

	userMap := make(map[string]*CallerDTO)
	for _, log := range *logs {
		if log.UserID == nil {
			continue // Skip unauthenticated calls
		}
		userID := *log.UserID
		if _, exists := userMap[userID]; !exists {
			userMap[userID] = &CallerDTO{
				UserID:     userID,
				TotalCalls: 0,
				LastCall:   log.RequestedAt,
			}
		}
		userMap[userID].TotalCalls++
		if log.RequestedAt.After(userMap[userID].LastCall) {
			userMap[userID].LastCall = log.RequestedAt
		}
	}

	callers := make([]CallerDTO, 0, len(userMap))
	for _, caller := range userMap {
		callers = append(callers, *caller)
	}

	// Sort by total calls descending
	sort.Slice(callers, func(i, j int) bool {
		return callers[i].TotalCalls > callers[j].TotalCalls
	})

	if len(callers) > limit {
		callers = callers[:limit]
	}

	return &callers, nil
}

// Helper function to find most used endpoint
func getMostUsedEndpoint(endpointMap map[string]int64) string {
	maxCount := int64(0)
	mostUsed := ""

	for endpoint, count := range endpointMap {
		if count > maxCount {
			maxCount = count
			mostUsed = endpoint
		}
	}

	return mostUsed
}

// DTOs for API responses

// EndpointDetailsDTO contains detailed information about an endpoint
type EndpointDetailsDTO struct {
	Endpoint        string                   `json:"endpoint"`
	Method          string                   `json:"method"`
	TotalRequests   int64                    `json:"total_requests"`
	SuccessRequests int64                    `json:"success_requests"`
	ErrorRequests   int64                    `json:"error_requests"`
	SuccessRate     float64                  `json:"success_rate"`
	ErrorRate       float64                  `json:"error_rate"`
	AvgResponseTime int64                    `json:"avg_response_time_ms"`
	MinResponseTime int64                    `json:"min_response_time_ms"`
	MaxResponseTime int64                    `json:"max_response_time_ms"`
	Last24Hours     int64                    `json:"last_24_hours"`
	LastAccessedAt  time.Time                `json:"last_accessed_at"`
	RecentLogs      *[]api_usage.APIUsageLog `json:"recent_logs,omitempty"`
}

// UsageSummaryDTO contains overall usage summary
type UsageSummaryDTO struct {
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	FailedRequests     int64     `json:"failed_requests"`
	SuccessRate        float64   `json:"success_rate"`
	ErrorRate          float64   `json:"error_rate"`
	UniqueEndpoints    int64     `json:"unique_endpoints"`
	AvgResponseTime    int64     `json:"avg_response_time_ms"`
	MinResponseTime    int64     `json:"min_response_time_ms"`
	MaxResponseTime    int64     `json:"max_response_time_ms"`
	Timestamp          time.Time `json:"timestamp"`
}

// UsageTrendDTO represents usage trend data
type UsageTrendDTO struct {
	Date            time.Time `json:"date"`
	RequestCount    int64     `json:"request_count"`
	SuccessCount    int64     `json:"success_count"`
	ErrorCount      int64     `json:"error_count"`
	AvgResponseTime int64     `json:"avg_response_time_ms"`
}

// UserActivityDTO contains user activity summary
type UserActivityDTO struct {
	UserID                  string `json:"user_id"`
	TotalRequests           int64  `json:"total_requests"`
	SuccessfulRequests      int64  `json:"successful_requests"`
	FailedRequests          int64  `json:"failed_requests"`
	AvgResponseTime         int64  `json:"avg_response_time_ms"`
	UniqueEndpointsAccessed int64  `json:"unique_endpoints_accessed"`
	MostUsedEndpoint        string `json:"most_used_endpoint"`
}

// CallerDTO contains information about who called an endpoint
type CallerDTO struct {
	UserID     string    `json:"user_id"`
	TotalCalls int64     `json:"total_calls"`
	LastCall   time.Time `json:"last_call"`
}
