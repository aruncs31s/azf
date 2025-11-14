package persistence

import (
	"fmt"
	"time"

	"github.com/aruncs31s/azf/domain/api_usage"
	"github.com/aruncs31s/azf/domain/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NewAPIUsageRepository creates a new API usage repository
func NewAPIUsageRepository(db *gorm.DB) repository.APIUsageLogRepository {
	reader := newAPIUsageLogReader(db)
	writer := newAPIUsageLogWriter(db)
	return &apiUsageRepository{
		reader: reader,
		writer: writer,
		db:     db,
	}
}

type apiUsageRepository struct {
	reader repository.APIUsageLogReader
	writer repository.APIUsageLogWriter
	db     *gorm.DB
}

// Reader operations
func (r *apiUsageRepository) FindByID(id string) (*api_usage.APIUsageLog, error) {
	return r.reader.FindByID(id)
}

func (r *apiUsageRepository) FindAll(limit int, offset int) (*[]api_usage.APIUsageLog, error) {
	return r.reader.FindAll(limit, offset)
}

func (r *apiUsageRepository) FindByEndpoint(endpoint string, limit int, offset int) (*[]api_usage.APIUsageLog, error) {
	return r.reader.FindByEndpoint(endpoint, limit, offset)
}

func (r *apiUsageRepository) FindByUserID(userID string, limit int, offset int) (*[]api_usage.APIUsageLog, error) {
	return r.reader.FindByUserID(userID, limit, offset)
}

func (r *apiUsageRepository) FindByDateRange(startDate string, endDate string, limit int, offset int) (*[]api_usage.APIUsageLog, error) {
	return r.reader.FindByDateRange(startDate, endDate, limit, offset)
}

func (r *apiUsageRepository) CountByEndpoint(endpoint string) (int64, error) {
	return r.reader.CountByEndpoint(endpoint)
}

func (r *apiUsageRepository) CountTotal() (int64, error) {
	return r.reader.CountTotal()
}

// Writer operations
func (r *apiUsageRepository) Create(log *api_usage.APIUsageLog) (*api_usage.APIUsageLog, error) {
	return r.writer.Create(log)
}

func (r *apiUsageRepository) BatchCreate(logs *[]api_usage.APIUsageLog) error {
	return r.writer.BatchCreate(logs)
}

func (r *apiUsageRepository) DeleteOlderThan(days int) error {
	return r.writer.DeleteOlderThan(days)
}

func (r *apiUsageRepository) DeleteAll() error {
	return r.writer.DeleteAll()
}

// === Reader Implementation ===

type apiUsageLogReader struct {
	db *gorm.DB
}

func newAPIUsageLogReader(db *gorm.DB) repository.APIUsageLogReader {
	return &apiUsageLogReader{db: db}
}

func (r *apiUsageLogReader) FindByID(id string) (*api_usage.APIUsageLog, error) {
	var log api_usage.APIUsageLog
	if err := r.db.Where("id = ?", id).First(&log).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

func (r *apiUsageLogReader) FindAll(limit int, offset int) (*[]api_usage.APIUsageLog, error) {
	var logs []api_usage.APIUsageLog
	if err := r.db.Order("requested_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, err
	}
	return &logs, nil
}

func (r *apiUsageLogReader) FindByEndpoint(endpoint string, limit int, offset int) (*[]api_usage.APIUsageLog, error) {
	var logs []api_usage.APIUsageLog
	if err := r.db.Where("endpoint = ?", endpoint).Order("requested_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, err
	}
	return &logs, nil
}

func (r *apiUsageLogReader) FindByUserID(userID string, limit int, offset int) (*[]api_usage.APIUsageLog, error) {
	var logs []api_usage.APIUsageLog
	if err := r.db.Where("user_id = ?", userID).Order("requested_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, err
	}
	return &logs, nil
}

func (r *apiUsageLogReader) FindByDateRange(startDate string, endDate string, limit int, offset int) (*[]api_usage.APIUsageLog, error) {
	var logs []api_usage.APIUsageLog
	if err := r.db.Where("requested_at BETWEEN ? AND ?", startDate, endDate).Order("requested_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, err
	}
	return &logs, nil
}

func (r *apiUsageLogReader) CountByEndpoint(endpoint string) (int64, error) {
	var count int64
	if err := r.db.Model(&api_usage.APIUsageLog{}).Where("endpoint = ?", endpoint).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *apiUsageLogReader) CountTotal() (int64, error) {
	var count int64
	if err := r.db.Model(&api_usage.APIUsageLog{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// === Writer Implementation ===

type apiUsageLogWriter struct {
	db *gorm.DB
}

func newAPIUsageLogWriter(db *gorm.DB) repository.APIUsageLogWriter {
	return &apiUsageLogWriter{db: db}
}

func (w *apiUsageLogWriter) Create(log *api_usage.APIUsageLog) (*api_usage.APIUsageLog, error) {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	if err := w.db.Create(log).Error; err != nil {
		return nil, err
	}
	return log, nil
}

func (w *apiUsageLogWriter) BatchCreate(logs *[]api_usage.APIUsageLog) error {
	for i := range *logs {
		if (*logs)[i].ID == "" {
			(*logs)[i].ID = uuid.New().String()
		}
		if (*logs)[i].CreatedAt.IsZero() {
			(*logs)[i].CreatedAt = time.Now()
		}
	}
	if err := w.db.CreateInBatches(logs, 100).Error; err != nil {
		return err
	}
	return nil
}

func (w *apiUsageLogWriter) DeleteOlderThan(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	if err := w.db.Where("requested_at < ?", cutoffDate).Delete(&api_usage.APIUsageLog{}).Error; err != nil {
		return fmt.Errorf("failed to delete old API usage logs: %w", err)
	}
	return nil
}

func (w *apiUsageLogWriter) DeleteAll() error {
	if err := w.db.Delete(&api_usage.APIUsageLog{}).Error; err != nil {
		return fmt.Errorf("failed to delete all API usage logs: %w", err)
	}
	return nil
}

// === Stats Repository Implementation ===

// NewAPIUsageStatsRepository creates a new API usage stats repository
func NewAPIUsageStatsRepository(db *gorm.DB) repository.APIUsageStatsRepository {
	reader := newAPIUsageStatsReader(db)
	writer := newAPIUsageStatsWriter(db)
	return &apiUsageStatsRepository{
		reader: reader,
		writer: writer,
		db:     db,
	}
}

type apiUsageStatsRepository struct {
	reader repository.APIUsageStatsReader
	writer repository.APIUsageStatsWriter
	db     *gorm.DB
}

// Reader operations
func (r *apiUsageStatsRepository) FindByID(id string) (*api_usage.APIUsageStats, error) {
	return r.reader.FindByID(id)
}

func (r *apiUsageStatsRepository) FindAll(limit int, offset int) (*[]api_usage.APIUsageStats, error) {
	return r.reader.FindAll(limit, offset)
}

func (r *apiUsageStatsRepository) FindByEndpoint(endpoint string) (*api_usage.APIUsageStats, error) {
	return r.reader.FindByEndpoint(endpoint)
}

func (r *apiUsageStatsRepository) GetTopEndpointsByUsage(limit int) (*[]api_usage.APIEndpointRanking, error) {
	return r.reader.GetTopEndpointsByUsage(limit)
}

func (r *apiUsageStatsRepository) GetEndpointsByErrorRate(limit int) (*[]api_usage.APIEndpointRanking, error) {
	return r.reader.GetEndpointsByErrorRate(limit)
}

func (r *apiUsageStatsRepository) GetEndpointsByResponseTime(limit int) (*[]api_usage.APIEndpointRanking, error) {
	return r.reader.GetEndpointsByResponseTime(limit)
}

func (r *apiUsageStatsRepository) CountTotal() (int64, error) {
	return r.reader.CountTotal()
}

// Writer operations
func (r *apiUsageStatsRepository) Create(stats *api_usage.APIUsageStats) (*api_usage.APIUsageStats, error) {
	return r.writer.Create(stats)
}

func (r *apiUsageStatsRepository) Update(stats *api_usage.APIUsageStats) (*api_usage.APIUsageStats, error) {
	return r.writer.Update(stats)
}

func (r *apiUsageStatsRepository) Upsert(stats *api_usage.APIUsageStats) (*api_usage.APIUsageStats, error) {
	return r.writer.Upsert(stats)
}

func (r *apiUsageStatsRepository) RecalculateStats(endpoint string, method string) error {
	return r.writer.RecalculateStats(endpoint, method)
}

func (r *apiUsageStatsRepository) DeleteAll() error {
	return r.writer.DeleteAll()
}

// === Stats Reader Implementation ===

type apiUsageStatsReader struct {
	db *gorm.DB
}

func newAPIUsageStatsReader(db *gorm.DB) repository.APIUsageStatsReader {
	return &apiUsageStatsReader{db: db}
}

func (r *apiUsageStatsReader) FindByID(id string) (*api_usage.APIUsageStats, error) {
	var stats api_usage.APIUsageStats
	if err := r.db.Where("id = ?", id).First(&stats).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &stats, nil
}

func (r *apiUsageStatsReader) FindAll(limit int, offset int) (*[]api_usage.APIUsageStats, error) {
	var stats []api_usage.APIUsageStats
	if err := r.db.Order("total_requests DESC").Limit(limit).Offset(offset).Find(&stats).Error; err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *apiUsageStatsReader) FindByEndpoint(endpoint string) (*api_usage.APIUsageStats, error) {
	var stats api_usage.APIUsageStats
	if err := r.db.Where("endpoint = ?", endpoint).First(&stats).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &stats, nil
}

func (r *apiUsageStatsReader) GetTopEndpointsByUsage(limit int) (*[]api_usage.APIEndpointRanking, error) {
	var rankings []api_usage.APIEndpointRanking
	if err := r.db.Model(&api_usage.APIUsageStats{}).
		Select("endpoint, method, total_requests, success_requests, error_requests, avg_response_time, last24_hours, ROW_NUMBER() OVER (ORDER BY total_requests DESC) as rank").
		Order("total_requests DESC").
		Limit(limit).
		Scan(&rankings).Error; err != nil {
		return nil, err
	}
	return &rankings, nil
}

func (r *apiUsageStatsReader) GetEndpointsByErrorRate(limit int) (*[]api_usage.APIEndpointRanking, error) {
	var rankings []api_usage.APIEndpointRanking
	// Calculate error rate: (error_requests / total_requests) * 100
	if err := r.db.Model(&api_usage.APIUsageStats{}).
		Select("endpoint, method, total_requests, success_requests, error_requests, avg_response_time, last24_hours, ROW_NUMBER() OVER (ORDER BY (CAST(error_requests AS FLOAT) / total_requests) DESC) as rank").
		Where("total_requests > 0").
		Order("(CAST(error_requests AS FLOAT) / total_requests) DESC").
		Limit(limit).
		Scan(&rankings).Error; err != nil {
		return nil, err
	}
	return &rankings, nil
}

func (r *apiUsageStatsReader) GetEndpointsByResponseTime(limit int) (*[]api_usage.APIEndpointRanking, error) {
	var rankings []api_usage.APIEndpointRanking
	if err := r.db.Model(&api_usage.APIUsageStats{}).
		Select("endpoint, method, total_requests, success_requests, error_requests, avg_response_time, last24_hours, ROW_NUMBER() OVER (ORDER BY avg_response_time DESC) as rank").
		Order("avg_response_time DESC").
		Limit(limit).
		Scan(&rankings).Error; err != nil {
		return nil, err
	}
	return &rankings, nil
}

func (r *apiUsageStatsReader) CountTotal() (int64, error) {
	var count int64
	if err := r.db.Model(&api_usage.APIUsageStats{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// === Stats Writer Implementation ===

type apiUsageStatsWriter struct {
	db *gorm.DB
}

func newAPIUsageStatsWriter(db *gorm.DB) repository.APIUsageStatsWriter {
	return &apiUsageStatsWriter{db: db}
}

func (w *apiUsageStatsWriter) Create(stats *api_usage.APIUsageStats) (*api_usage.APIUsageStats, error) {
	if stats.ID == "" {
		stats.ID = uuid.New().String()
	}
	if stats.CreatedAt.IsZero() {
		stats.CreatedAt = time.Now()
	}
	if err := w.db.Create(stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func (w *apiUsageStatsWriter) Update(stats *api_usage.APIUsageStats) (*api_usage.APIUsageStats, error) {
	stats.UpdatedAt = time.Now()
	if err := w.db.Save(stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func (w *apiUsageStatsWriter) Upsert(stats *api_usage.APIUsageStats) (*api_usage.APIUsageStats, error) {
	if stats.ID == "" {
		stats.ID = uuid.New().String()
	}
	stats.UpdatedAt = time.Now()
	if err := w.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(stats).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func (w *apiUsageStatsWriter) RecalculateStats(endpoint string, method string) error {
	// Calculate stats from logs
	var stats api_usage.APIUsageStats

	// Get existing stats or create new
	if err := w.db.Where("endpoint = ? AND method = ?", endpoint, method).First(&stats).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if stats.ID == "" {
		stats.ID = uuid.New().String()
		stats.CreatedAt = time.Now()
	}

	stats.Endpoint = endpoint
	stats.Method = method
	stats.UpdatedAt = time.Now()

	// Query logs for this endpoint/method
	var logs []api_usage.APIUsageLog
	if err := w.db.Where("endpoint = ? AND method = ?", endpoint, method).Find(&logs).Error; err != nil {
		return err
	}

	// Calculate metrics
	stats.TotalRequests = int64(len(logs))
	if stats.TotalRequests == 0 {
		return w.db.Save(&stats).Error
	}

	var totalResponseTime int64
	var minResponseTime int64 = 999999999
	var maxResponseTime int64 = 0
	var successCount int64
	var errorCount int64
	last24hCount := int64(0)
	now := time.Now()
	oneDayAgo := now.AddDate(0, 0, -1)

	for _, log := range logs {
		totalResponseTime += log.ResponseTime
		if log.ResponseTime < minResponseTime {
			minResponseTime = log.ResponseTime
		}
		if log.ResponseTime > maxResponseTime {
			maxResponseTime = log.ResponseTime
		}

		if log.StatusCode >= 200 && log.StatusCode < 300 {
			successCount++
		} else {
			errorCount++
		}

		if log.RequestedAt.After(oneDayAgo) {
			last24hCount++
		}

		if log.LastAccessedAt.After(stats.LastAccessedAt) {
			stats.LastAccessedAt = log.RequestedAt
		}
	}

	stats.SuccessRequests = successCount
	stats.ErrorRequests = errorCount
	stats.AvgResponseTime = totalResponseTime / stats.TotalRequests
	stats.MinResponseTime = minResponseTime
	stats.MaxResponseTime = maxResponseTime
	stats.Last24Hours = last24hCount

	return w.db.Save(&stats).Error
}

func (w *apiUsageStatsWriter) DeleteAll() error {
	if err := w.db.Delete(&api_usage.APIUsageStats{}).Error; err != nil {
		return fmt.Errorf("failed to delete all API usage statistics: %w", err)
	}
	return nil
}
