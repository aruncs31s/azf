package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aruncs31s/azf/infrastructure/enterprise"
	"github.com/aruncs31s/azf/shared/logger"
	"go.uber.org/zap"
)

// AuthorizationAuditService provides business logic for authorization audit logs
type AuthorizationAuditService interface {
	GetAuditLogs(limit int, offset int) (*[]AuditLogDTO, error)
	GetAuditLogsByUser(userID string, limit int, offset int) (*[]AuditLogDTO, error)
	GetAuditLogsByResult(result string, limit int, offset int) (*[]AuditLogDTO, error)
	GetAuditLogsByTimeRange(startTime, endTime time.Time, limit int, offset int) (*[]AuditLogDTO, error)
	GetAuditLogsByResource(resource string, limit int, offset int) (*[]AuditLogDTO, error)
	GetDeniedAccessLogs(limit int, offset int) (*[]AuditLogDTO, error)
	GetAuditSummary() (*AuditSummaryDTO, error)
	GetCriticalEvents(limit int, offset int) (*[]AuditLogDTO, error)
	CleanupOldLogs(olderThan time.Duration) (int64, error)
}

// authorizationAuditService implements AuthorizationAuditService
type authorizationAuditService struct {
	auditRepo *enterprise.AuthorizationAuditRepository
}

// NewAuthorizationAuditService creates a new authorization audit service
func NewAuthorizationAuditService(auditRepo *enterprise.AuthorizationAuditRepository) AuthorizationAuditService {
	return &authorizationAuditService{
		auditRepo: auditRepo,
	}
}

// GetAuditLogs returns paginated audit logs
func (s *authorizationAuditService) GetAuditLogs(limit int, offset int) (*[]AuditLogDTO, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}

	logs, err := s.auditRepo.FindAll(context.Background(), limit, offset)
	if err != nil {
		logger.Error("Failed to get audit logs", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve audit logs: %w", err)
	}

	dtos := make([]AuditLogDTO, len(logs))
	for i, log := range logs {
		dtos[i] = s.convertToDTO(log)
	}

	return &dtos, nil
}

// GetAuditLogsByUser returns audit logs for a specific user
func (s *authorizationAuditService) GetAuditLogsByUser(userID string, limit int, offset int) (*[]AuditLogDTO, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}

	logs, err := s.auditRepo.FindByUserID(context.Background(), userID, limit, offset)
	if err != nil {
		logger.Error("Failed to get audit logs by user", zap.String("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve audit logs for user %s: %w", userID, err)
	}

	dtos := make([]AuditLogDTO, len(logs))
	for i, log := range logs {
		dtos[i] = s.convertToDTO(log)
	}

	return &dtos, nil
}

// GetAuditLogsByResult returns audit logs filtered by authorization result
func (s *authorizationAuditService) GetAuditLogsByResult(result string, limit int, offset int) (*[]AuditLogDTO, error) {
	validResults := map[string]bool{"ALLOWED": true, "DENIED": true, "WARNING": true}
	if !validResults[result] {
		return nil, fmt.Errorf("invalid result: %s", result)
	}

	logs, err := s.auditRepo.FindByResult(context.Background(), result, limit, offset)
	if err != nil {
		logger.Error("Failed to get audit logs by result", zap.String("result", result), zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve audit logs by result %s: %w", result, err)
	}

	dtos := make([]AuditLogDTO, len(logs))
	for i, log := range logs {
		dtos[i] = s.convertToDTO(log)
	}

	return &dtos, nil
}

// GetAuditLogsByTimeRange returns audit logs within a time range
func (s *authorizationAuditService) GetAuditLogsByTimeRange(startTime, endTime time.Time, limit int, offset int) (*[]AuditLogDTO, error) {
	if startTime.After(endTime) {
		return nil, fmt.Errorf("start time cannot be after end time")
	}

	logs, err := s.auditRepo.FindByTimeRange(context.Background(), startTime, endTime, limit, offset)
	if err != nil {
		logger.Error("Failed to get audit logs by time range",
			zap.Time("start_time", startTime),
			zap.Time("end_time", endTime),
			zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve audit logs by time range: %w", err)
	}

	dtos := make([]AuditLogDTO, len(logs))
	for i, log := range logs {
		dtos[i] = s.convertToDTO(log)
	}

	return &dtos, nil
}

// GetAuditLogsByResource returns audit logs for a specific resource
func (s *authorizationAuditService) GetAuditLogsByResource(resource string, limit int, offset int) (*[]AuditLogDTO, error) {
	if resource == "" {
		return nil, fmt.Errorf("resource cannot be empty")
	}

	logs, err := s.auditRepo.FindByResource(context.Background(), resource, limit, offset)
	if err != nil {
		logger.Error("Failed to get audit logs by resource", zap.String("resource", resource), zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve audit logs for resource %s: %w", resource, err)
	}

	dtos := make([]AuditLogDTO, len(logs))
	for i, log := range logs {
		dtos[i] = s.convertToDTO(log)
	}

	return &dtos, nil
}

// GetDeniedAccessLogs returns logs where access was denied
func (s *authorizationAuditService) GetDeniedAccessLogs(limit int, offset int) (*[]AuditLogDTO, error) {
	logs, err := s.auditRepo.FindDeniedAccess(context.Background(), limit, offset)
	if err != nil {
		logger.Error("Failed to get denied access logs", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve denied access logs: %w", err)
	}

	dtos := make([]AuditLogDTO, len(logs))
	for i, log := range logs {
		dtos[i] = s.convertToDTO(log)
	}

	return &dtos, nil
}

// GetAuditSummary returns summary statistics for audit logs
func (s *authorizationAuditService) GetAuditSummary() (*AuditSummaryDTO, error) {
	totalCount, err := s.auditRepo.Count(context.Background())
	if err != nil {
		logger.Error("Failed to count audit logs", zap.Error(err))
		return nil, fmt.Errorf("failed to get audit summary: %w", err)
	}

	// Get recent logs (last 24 hours)
	yesterday := time.Now().Add(-24 * time.Hour)
	recentLogs, err := s.auditRepo.FindByTimeRange(context.Background(), yesterday, time.Now(), 10000, 0)
	if err != nil {
		logger.Warn("Failed to get recent logs for summary", zap.Error(err))
		recentLogs = []*enterprise.AuthorizationAuditLogDB{}
	}

	// Calculate statistics
	var allowedCount, deniedCount, warningCount int64
	var totalExecutionTime float64
	denialReasons := make(map[string]int64)
	resources := make(map[string]int64)

	for _, log := range recentLogs {
		switch log.Result {
		case "ALLOWED":
			allowedCount++
		case "DENIED":
			deniedCount++
			if log.Reason != "" {
				denialReasons[log.Reason]++
			}
		case "WARNING":
			warningCount++
		}
		totalExecutionTime += log.ExecutionTimeMs
		if log.Resource != "" {
			resources[log.Resource]++
		}
	}

	avgExecutionTime := 0.0
	if len(recentLogs) > 0 {
		avgExecutionTime = totalExecutionTime / float64(len(recentLogs))
	}

	return &AuditSummaryDTO{
		TotalLogs:        totalCount,
		RecentLogs24h:    int64(len(recentLogs)),
		AllowedCount24h:  allowedCount,
		DeniedCount24h:   deniedCount,
		WarningCount24h:  warningCount,
		AvgExecutionTime: avgExecutionTime,
		TopDenialReasons: denialReasons,
		TopResources:     resources,
		GeneratedAt:      time.Now(),
	}, nil
}

// GetCriticalEvents returns critical audit events (denials, rate limits, deprecated routes)
func (s *authorizationAuditService) GetCriticalEvents(limit int, offset int) (*[]AuditLogDTO, error) {
	// Get denied access logs
	deniedLogs, err := s.auditRepo.FindDeniedAccess(context.Background(), limit/2, offset)
	if err != nil {
		logger.Error("Failed to get denied logs for critical events", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve critical events: %w", err)
	}

	// Get rate limit exceeded logs
	rateLimitLogs, err := s.auditRepo.FindRateLimitExceeded(context.Background(), limit/2, offset)
	if err != nil {
		logger.Warn("Failed to get rate limit logs for critical events", zap.Error(err))
	}

	// Get deprecated route logs
	deprecatedLogs, err := s.auditRepo.FindDeprecatedRouteAccess(context.Background(), limit/2, offset)
	if err != nil {
		logger.Warn("Failed to get deprecated route logs for critical events", zap.Error(err))
	}

	// Combine and deduplicate
	allLogs := make(map[string]*enterprise.AuthorizationAuditLogDB)
	for _, log := range deniedLogs {
		allLogs[log.ID] = log
	}
	for _, log := range rateLimitLogs {
		allLogs[log.ID] = log
	}
	for _, log := range deprecatedLogs {
		allLogs[log.ID] = log
	}

	// Convert to slice and sort by timestamp (most recent first)
	logs := make([]*enterprise.AuthorizationAuditLogDB, 0, len(allLogs))
	for _, log := range allLogs {
		logs = append(logs, log)
	}

	// Simple sort by timestamp (in production, you'd want a proper sort)
	// For now, just take the first 'limit' items
	if len(logs) > limit {
		logs = logs[:limit]
	}

	dtos := make([]AuditLogDTO, len(logs))
	for i, log := range logs {
		dtos[i] = s.convertToDTO(log)
	}

	return &dtos, nil
}

// CleanupOldLogs removes audit logs older than the specified duration
func (s *authorizationAuditService) CleanupOldLogs(olderThan time.Duration) (int64, error) {
	if olderThan <= 0 {
		return 0, fmt.Errorf("cleanup duration must be positive")
	}

	deletedCount, err := s.auditRepo.CleanupOldLogs(context.Background(), olderThan)
	if err != nil {
		logger.Error("Failed to cleanup old audit logs",
			zap.Duration("older_than", olderThan),
			zap.Error(err))
		return 0, fmt.Errorf("failed to cleanup old logs: %w", err)
	}

	logger.Info("Cleaned up old audit logs",
		zap.Int64("deleted_count", deletedCount),
		zap.Duration("older_than", olderThan))

	return deletedCount, nil
}

// convertToDTO converts database model to DTO
func (s *authorizationAuditService) convertToDTO(log *enterprise.AuthorizationAuditLogDB) AuditLogDTO {
	return AuditLogDTO{
		ID:              log.ID,
		Timestamp:       log.Timestamp,
		UserID:          log.UserID,
		Role:            log.Role,
		Resource:        log.Resource,
		Action:          log.Action,
		Result:          log.Result,
		DenialReason:    log.Reason,
		IPAddress:       log.IPAddress,
		UserAgent:       log.UserAgent,
		APIVersion:      log.APIVersion,
		Deprecated:      log.Deprecated,
		Environment:     log.Environment,
		RateLimitStatus: log.RateLimitStatus,
		PolicyVersion:   log.PolicyVersion,
		ExecutionTimeMs: log.ExecutionTimeMs,
	}
}

// DTOs for API responses

// AuditLogDTO represents an audit log entry for API responses
type AuditLogDTO struct {
	ID              string    `json:"id"`
	Timestamp       time.Time `json:"timestamp"`
	UserID          string    `json:"user_id"`
	Role            string    `json:"role"`
	Resource        string    `json:"resource"`
	Action          string    `json:"action"`
	Result          string    `json:"result"`
	DenialReason    string    `json:"denial_reason,omitempty"`
	IPAddress       string    `json:"ip_address"`
	UserAgent       string    `json:"user_agent"`
	APIVersion      string    `json:"api_version"`
	Deprecated      bool      `json:"deprecated"`
	Environment     string    `json:"environment"`
	RateLimitStatus string    `json:"rate_limit_status"`
	PolicyVersion   int       `json:"policy_version"`
	ExecutionTimeMs float64   `json:"execution_time_ms"`
}

// AuditSummaryDTO contains summary statistics for audit logs
type AuditSummaryDTO struct {
	TotalLogs        int64            `json:"total_logs"`
	RecentLogs24h    int64            `json:"recent_logs_24h"`
	AllowedCount24h  int64            `json:"allowed_count_24h"`
	DeniedCount24h   int64            `json:"denied_count_24h"`
	WarningCount24h  int64            `json:"warning_count_24h"`
	AvgExecutionTime float64          `json:"avg_execution_time_ms"`
	TopDenialReasons map[string]int64 `json:"top_denial_reasons"`
	TopResources     map[string]int64 `json:"top_resources"`
	GeneratedAt      time.Time        `json:"generated_at"`
}
