package persistence

import (
	"github.com/aruncs31s/azf/domain/api_usage"
	"github.com/aruncs31s/azf/domain/repository"
	"gorm.io/gorm"
)

// NewAPIUsageRepository creates a new API usage repository
func NewAPIUsageRepository(
	db *gorm.DB,
) repository.APIUsageLogRepository {
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
