package model

import (
	"errors"
	"time"
)

// AuditLogFactory handles creation of AuditLog entities with proper validation
type AuditLogFactory struct{}

func NewAuditLogFactory() *AuditLogFactory {
	return &AuditLogFactory{}
}

// CreateSuccessLog creates a new successful audit log entry
func (f *AuditLogFactory) CreateSuccessLog(
	action *AuditAction,
	adminID *AdminID,
	ipAddress string,
	userAgent string,
	resourceID string,
	description string,
	details map[string]interface{},
) (*AuditLog, error) {
	status := &LogStatus{value: "SUCCESS"}
	return NewAuditLog(
		generateUniqueID(),
		time.Now(),
		action,
		adminID,
		ipAddress,
		userAgent,
		resourceID,
		description,
		status,
		"",
		details,
	)
}

// CreateErrorLog creates a new error audit log entry
func (f *AuditLogFactory) CreateErrorLog(
	action *AuditAction,
	adminID *AdminID,
	ipAddress string,
	userAgent string,
	resourceID string,
	description string,
	errorMsg string,
	details map[string]interface{},
) (*AuditLog, error) {
	if errorMsg == "" {
		return nil, errors.New("error message is required for error logs")
	}

	status := &LogStatus{value: "FAILURE"}
	return NewAuditLog(
		generateUniqueID(),
		time.Now(),
		action,
		adminID,
		ipAddress,
		userAgent,
		resourceID,
		description,
		status,
		errorMsg,
		details,
	)
}

// generateUniqueID generates a unique ID for audit logs
func generateUniqueID() string {
	return time.Now().Format("20060102150405.000") + RandomString(8)
}

// RandomString generates a random string of specified length
func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
