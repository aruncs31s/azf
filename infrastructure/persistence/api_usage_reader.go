package persistence

import (
	"github.com/aruncs31s/azf/domain/api_usage"
	"github.com/aruncs31s/azf/domain/repository"
	"gorm.io/gorm"
)

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

	err := r.db.Model(&api_usage.APIUsageStats{}).
		Select(`
        endpoint,
        method,
        total_requests,
        success_requests,
        error_requests,
        avg_response_time,
        last24_hours,
        ROW_NUMBER() OVER (ORDER BY total_requests DESC) AS rank
    `).Order(
		"rank",
	).Limit(limit).Error

	// err := r.db.
	// 	Table("(?) as ranked", sub).
	// 	Order("rank").
	// 	Limit(limit).
	// 	Scan(&rankings).Error
	if err != nil {
		return nil, err
	}
	return &rankings, nil
}

func (r *apiUsageStatsReader) GetEndpointsByErrorRate(limit int) (*[]api_usage.APIEndpointRanking, error) {
	var rankings []api_usage.APIEndpointRanking
	// Calculate error rate: (error_requests / total_requests) * 100
	if err := r.db.Model(&api_usage.APIUsageStats{}).
		Select("endpoint, method, total_requests, success_requests, error_requests, avg_response_time, last24_hours, ROW_NUMBER() OVER (ORDER BY (CAST(error_requests AS FLOAT) / total_requests) DESC) as `rank`").
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
		Select(
			[]string{
				"endpoint",
				"method",
				"total_requests",
				"success_requests",
				"error_requests",
				"avg_response_time",
				"last24_hours",
				"ROW_NUMBER() OVER (ORDER BY avg_response_time DESC) as `rank`",
			}).
		// Order("avg_response_time DESC").
		Order("`rank`").
		// Limit(limit).
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
