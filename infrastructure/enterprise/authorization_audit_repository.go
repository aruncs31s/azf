package enterprise

import (
	"context"
	"fmt"
	"time"

	"github.com/aruncs31s/azf/domain/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthorizationAuditRepository handles persistence of authorization audit logs
type AuthorizationAuditRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAuthorizationAuditRepository creates a new authorization audit repository
func NewAuthorizationAuditRepository(db *gorm.DB, logger *zap.Logger) *AuthorizationAuditRepository {
	return &AuthorizationAuditRepository{
		db:     db,
		logger: logger,
	}
}

// Save persists an authorization audit log to the database
func (aar *AuthorizationAuditRepository) Save(ctx context.Context, log *model.AuthorizationAuditLog) error {
	if log == nil {
		return fmt.Errorf("authorization audit log cannot be nil")
	}

	// Convert domain model to database model
	dbLog := &AuthorizationAuditLogDB{
		ID:              log.ID(),
		UserID:          log.UserID(),
		Role:            log.Role(),
		Resource:        log.Resource(),
		Action:          log.Action(),
		Result:          log.Result().Value(),
		Reason:          "",
		IPAddress:       log.IPAddress(),
		UserAgent:       log.UserAgent(),
		Timestamp:       log.Timestamp(),
		RequestID:       "",
		ErrorMsg:        "",
		Environment:     log.Environment(),
		APIVersion:      log.APIVersion(),
		Deprecated:      log.Deprecated(),
		RateLimitStatus: log.RateLimitStatus(),
		PolicyVersion:   log.PolicyVersion(),
		ExecutionTimeMs: log.ExecutionTimeMs(),
	}

	if log.DenialReason() != nil {
		dbLog.Reason = log.DenialReason().Value()
	}

	result := aar.db.WithContext(ctx).Create(dbLog)
	if result.Error != nil {
		aar.logger.Error("Failed to save authorization audit log",
			zap.Error(result.Error),
			zap.String("user_id", log.UserID()),
			zap.String("resource", log.Resource()))
		return fmt.Errorf("failed to save audit log: %w", result.Error)
	}

	aar.logger.Debug("Authorization audit log saved",
		zap.String("user_id", log.UserID()),
		zap.String("resource", log.Resource()),
		zap.String("result", log.Result().Value()))

	return nil
}

// SaveBatch persists multiple authorization audit logs
func (aar *AuthorizationAuditRepository) SaveBatch(ctx context.Context, logs []*model.AuthorizationAuditLog) error {
	if len(logs) == 0 {
		return nil
	}

	dbLogs := make([]*AuthorizationAuditLogDB, len(logs))
	for i, log := range logs {
		dbLog := &AuthorizationAuditLogDB{
			ID:              log.ID(),
			UserID:          log.UserID(),
			Role:            log.Role(),
			Resource:        log.Resource(),
			Action:          log.Action(),
			Result:          log.Result().Value(),
			IPAddress:       log.IPAddress(),
			UserAgent:       log.UserAgent(),
			Timestamp:       log.Timestamp(),
			Environment:     log.Environment(),
			APIVersion:      log.APIVersion(),
			Deprecated:      log.Deprecated(),
			RateLimitStatus: log.RateLimitStatus(),
			PolicyVersion:   log.PolicyVersion(),
			ExecutionTimeMs: log.ExecutionTimeMs(),
		}

		if log.DenialReason() != nil {
			dbLog.Reason = log.DenialReason().Value()
		}

		dbLogs[i] = dbLog
	}

	result := aar.db.WithContext(ctx).CreateInBatches(dbLogs, 100)
	if result.Error != nil {
		aar.logger.Error("Failed to save authorization audit log batch",
			zap.Error(result.Error),
			zap.Int("count", len(logs)))
		return fmt.Errorf("failed to save audit log batch: %w", result.Error)
	}

	aar.logger.Debug("Authorization audit logs batch saved", zap.Int("count", len(logs)))
	return nil
}

// FindAll retrieves all audit logs with pagination
func (aar *AuthorizationAuditRepository) FindAll(ctx context.Context, limit int, offset int) ([]*AuthorizationAuditLogDB, error) {
	var logs []*AuthorizationAuditLogDB

	result := aar.db.WithContext(ctx).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs)

	if result.Error != nil {
		aar.logger.Error("Failed to find all audit logs",
			zap.Error(result.Error))
		return nil, fmt.Errorf("failed to find audit logs: %w", result.Error)
	}

	return logs, nil
}

// FindByUserID retrieves audit logs for a specific user
func (aar *AuthorizationAuditRepository) FindByUserID(ctx context.Context, userID string, limit int, offset int) ([]*AuthorizationAuditLogDB, error) {
	var logs []*AuthorizationAuditLogDB

	result := aar.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs)

	if result.Error != nil {
		aar.logger.Error("Failed to find audit logs by user ID",
			zap.Error(result.Error),
			zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to find audit logs: %w", result.Error)
	}

	return logs, nil
}

// FindByRole retrieves audit logs for a specific role
func (aar *AuthorizationAuditRepository) FindByRole(ctx context.Context, role string, limit int, offset int) ([]*AuthorizationAuditLogDB, error) {
	var logs []*AuthorizationAuditLogDB

	result := aar.db.WithContext(ctx).
		Where("role = ?", role).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs)

	if result.Error != nil {
		aar.logger.Error("Failed to find audit logs by role",
			zap.Error(result.Error),
			zap.String("role", role))
		return nil, fmt.Errorf("failed to find audit logs: %w", result.Error)
	}

	return logs, nil
}

// FindByResource retrieves audit logs for a specific resource
func (aar *AuthorizationAuditRepository) FindByResource(ctx context.Context, resource string, limit int, offset int) ([]*AuthorizationAuditLogDB, error) {
	var logs []*AuthorizationAuditLogDB

	result := aar.db.WithContext(ctx).
		Where("resource = ?", resource).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs)

	if result.Error != nil {
		aar.logger.Error("Failed to find audit logs by resource",
			zap.Error(result.Error),
			zap.String("resource", resource))
		return nil, fmt.Errorf("failed to find audit logs: %w", result.Error)
	}

	return logs, nil
}

// FindByResult retrieves audit logs with a specific result (ALLOWED/DENIED)
func (aar *AuthorizationAuditRepository) FindByResult(ctx context.Context, result string, limit int, offset int) ([]*AuthorizationAuditLogDB, error) {
	var logs []*AuthorizationAuditLogDB

	validResults := map[string]bool{"ALLOWED": true, "DENIED": true, "WARNING": true}
	if !validResults[result] {
		return nil, fmt.Errorf("invalid result: %s", result)
	}

	dbResult := aar.db.WithContext(ctx).
		Where("result = ?", result).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs)

	if dbResult.Error != nil {
		aar.logger.Error("Failed to find audit logs by result",
			zap.Error(dbResult.Error),
			zap.String("result", result))
		return nil, fmt.Errorf("failed to find audit logs: %w", dbResult.Error)
	}

	return logs, nil
}

// FindDeniedAccess retrieves all denied access attempts
func (aar *AuthorizationAuditRepository) FindDeniedAccess(ctx context.Context, limit int, offset int) ([]*AuthorizationAuditLogDB, error) {
	return aar.FindByResult(ctx, "DENIED", limit, offset)
}

// FindByTimeRange retrieves audit logs within a time range
func (aar *AuthorizationAuditRepository) FindByTimeRange(ctx context.Context, startTime, endTime time.Time, limit int, offset int) ([]*AuthorizationAuditLogDB, error) {
	var logs []*AuthorizationAuditLogDB

	result := aar.db.WithContext(ctx).
		Where("timestamp BETWEEN ? AND ?", startTime, endTime).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs)

	if result.Error != nil {
		aar.logger.Error("Failed to find audit logs by time range",
			zap.Error(result.Error),
			zap.Time("start_time", startTime),
			zap.Time("end_time", endTime))
		return nil, fmt.Errorf("failed to find audit logs: %w", result.Error)
	}

	return logs, nil
}

// FindByIPAddress retrieves audit logs from a specific IP address
func (aar *AuthorizationAuditRepository) FindByIPAddress(ctx context.Context, ipAddress string, limit int, offset int) ([]*AuthorizationAuditLogDB, error) {
	var logs []*AuthorizationAuditLogDB

	result := aar.db.WithContext(ctx).
		Where("ip_address = ?", ipAddress).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs)

	if result.Error != nil {
		aar.logger.Error("Failed to find audit logs by IP address",
			zap.Error(result.Error),
			zap.String("ip_address", ipAddress))
		return nil, fmt.Errorf("failed to find audit logs: %w", result.Error)
	}

	return logs, nil
}

// FindDeprecatedRouteAccess retrieves access attempts to deprecated routes
func (aar *AuthorizationAuditRepository) FindDeprecatedRouteAccess(ctx context.Context, limit int, offset int) ([]*AuthorizationAuditLogDB, error) {
	var logs []*AuthorizationAuditLogDB

	result := aar.db.WithContext(ctx).
		Where("deprecated = ?", true).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs)

	if result.Error != nil {
		aar.logger.Error("Failed to find deprecated route access logs",
			zap.Error(result.Error))
		return nil, fmt.Errorf("failed to find audit logs: %w", result.Error)
	}

	return logs, nil
}

// FindRateLimitExceeded retrieves rate limit exceeded events
func (aar *AuthorizationAuditRepository) FindRateLimitExceeded(ctx context.Context, limit int, offset int) ([]*AuthorizationAuditLogDB, error) {
	var logs []*AuthorizationAuditLogDB

	result := aar.db.WithContext(ctx).
		Where("rate_limit_status = ?", "EXCEEDED").
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs)

	if result.Error != nil {
		aar.logger.Error("Failed to find rate limit exceeded logs",
			zap.Error(result.Error))
		return nil, fmt.Errorf("failed to find audit logs: %w", result.Error)
	}

	return logs, nil
}

// GetDenialStats returns statistics about denied access attempts
func (aar *AuthorizationAuditRepository) GetDenialStats(ctx context.Context, userID string) (map[string]interface{}, error) {
	type DenialStat struct {
		Resource string
		Reason   string
		Count    int64
	}

	var stats []DenialStat

	result := aar.db.WithContext(ctx).
		Table("authorization_audit_logs").
		Select("resource, reason, COUNT(*) as count").
		Where("result = ? AND user_id = ?", "DENIED", userID).
		Group("resource, reason").
		Order("count DESC").
		Scan(&stats)

	if result.Error != nil {
		aar.logger.Error("Failed to get denial stats",
			zap.Error(result.Error),
			zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get denial stats: %w", result.Error)
	}

	statsMap := make(map[string]interface{})
	statsMap["total_denials"] = len(stats)
	statsMap["details"] = stats

	return statsMap, nil
}

// GetResourceAccessStats returns access statistics for a resource
func (aar *AuthorizationAuditRepository) GetResourceAccessStats(ctx context.Context, resource string) (map[string]interface{}, error) {
	type AccessStat struct {
		Role   string
		Action string
		Result string
		Count  int64
	}

	var stats []AccessStat

	result := aar.db.WithContext(ctx).
		Table("authorization_audit_logs").
		Select("role, action, result, COUNT(*) as count").
		Where("resource = ?", resource).
		Group("role, action, result").
		Order("count DESC").
		Scan(&stats)

	if result.Error != nil {
		aar.logger.Error("Failed to get resource access stats",
			zap.Error(result.Error),
			zap.String("resource", resource))
		return nil, fmt.Errorf("failed to get resource access stats: %w", result.Error)
	}

	statsMap := make(map[string]interface{})
	statsMap["resource"] = resource
	statsMap["total_accesses"] = len(stats)
	statsMap["details"] = stats

	return statsMap, nil
}

// GetRoleAccessStats returns access statistics for a role
func (aar *AuthorizationAuditRepository) GetRoleAccessStats(ctx context.Context, role string) (map[string]interface{}, error) {
	type AccessStat struct {
		Resource string
		Action   string
		Result   string
		Count    int64
	}

	var stats []AccessStat

	result := aar.db.WithContext(ctx).
		Table("authorization_audit_logs").
		Select("resource, action, result, COUNT(*) as count").
		Where("role = ?", role).
		Group("resource, action, result").
		Order("count DESC").
		Scan(&stats)

	if result.Error != nil {
		aar.logger.Error("Failed to get role access stats",
			zap.Error(result.Error),
			zap.String("role", role))
		return nil, fmt.Errorf("failed to get role access stats: %w", result.Error)
	}

	statsMap := make(map[string]interface{})
	statsMap["role"] = role
	statsMap["total_accesses"] = len(stats)
	statsMap["details"] = stats

	return statsMap, nil
}

// GetIPAddressStats returns access statistics for an IP address
func (aar *AuthorizationAuditRepository) GetIPAddressStats(ctx context.Context, ipAddress string) (map[string]interface{}, error) {
	type AccessStat struct {
		UserID string
		Role   string
		Result string
		Count  int64
	}

	var stats []AccessStat

	result := aar.db.WithContext(ctx).
		Table("authorization_audit_logs").
		Select("user_id, role, result, COUNT(*) as count").
		Where("ip_address = ?", ipAddress).
		Group("user_id, role, result").
		Order("count DESC").
		Scan(&stats)

	if result.Error != nil {
		aar.logger.Error("Failed to get IP address stats",
			zap.Error(result.Error),
			zap.String("ip_address", ipAddress))
		return nil, fmt.Errorf("failed to get IP address stats: %w", result.Error)
	}

	statsMap := make(map[string]interface{})
	statsMap["ip_address"] = ipAddress
	statsMap["total_accesses"] = len(stats)
	statsMap["details"] = stats

	return statsMap, nil
}

// CleanupOldLogs deletes audit logs older than the specified duration
func (aar *AuthorizationAuditRepository) CleanupOldLogs(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)

	result := aar.db.WithContext(ctx).
		Where("timestamp < ?", cutoffTime).
		Delete(&AuthorizationAuditLogDB{})

	if result.Error != nil {
		aar.logger.Error("Failed to cleanup old audit logs",
			zap.Error(result.Error),
			zap.Duration("older_than", olderThan))
		return 0, fmt.Errorf("failed to cleanup audit logs: %w", result.Error)
	}

	aar.logger.Info("Cleaned up old authorization audit logs",
		zap.Int64("deleted_count", result.RowsAffected),
		zap.Time("cutoff_time", cutoffTime))

	return result.RowsAffected, nil
}

// Count returns the total number of authorization audit logs
func (aar *AuthorizationAuditRepository) Count(ctx context.Context) (int64, error) {
	var count int64

	result := aar.db.WithContext(ctx).
		Model(&AuthorizationAuditLogDB{}).
		Count(&count)

	if result.Error != nil {
		aar.logger.Error("Failed to count authorization audit logs", zap.Error(result.Error))
		return 0, fmt.Errorf("failed to count audit logs: %w", result.Error)
	}

	return count, nil
}

// AuthorizationAuditLogDB is the database model for authorization audit logs
type AuthorizationAuditLogDB struct {
	ID              string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID          string    `gorm:"index;type:varchar(36)" json:"user_id"`
	Role            string    `gorm:"index;type:varchar(50)" json:"role"`
	Resource        string    `gorm:"index;type:varchar(500)" json:"resource"`
	Action          string    `gorm:"type:varchar(20)" json:"action"`
	Result          string    `gorm:"index;type:varchar(20)" json:"result"`
	Reason          string    `gorm:"type:text" json:"reason"`
	IPAddress       string    `gorm:"index;type:varchar(50)" json:"ip_address"`
	UserAgent       string    `gorm:"type:text" json:"user_agent"`
	Timestamp       time.Time `gorm:"index;type:timestamp" json:"timestamp"`
	RequestID       string    `gorm:"type:varchar(36)" json:"request_id"`
	ErrorMsg        string    `gorm:"type:text" json:"error_msg"`
	Environment     string    `gorm:"type:varchar(20)" json:"environment"`
	APIVersion      string    `gorm:"type:varchar(20)" json:"api_version"`
	Deprecated      bool      `gorm:"type:boolean" json:"deprecated"`
	RateLimitStatus string    `gorm:"type:varchar(20)" json:"rate_limit_status"`
	PolicyVersion   int       `gorm:"type:int" json:"policy_version"`
	ExecutionTimeMs float64   `gorm:"type:float" json:"execution_time_ms"`
}

// TableName specifies the table name
func (AuthorizationAuditLogDB) TableName() string {
	return "authorization_audit_logs"
}
