package model

import (
	"fmt"
	"time"
)

// AuditAction is a value object representing types of audit actions
type AuditAction struct {
	value string
}

// Predefined audit action instances
var (
	ActionCreate = &AuditAction{value: "CREATE"}
	ActionUpdate = &AuditAction{value: "UPDATE"}
	ActionDelete = &AuditAction{value: "DELETE"}
	ActionRead   = &AuditAction{value: "READ"}
	ActionLogin  = &AuditAction{value: "LOGIN"}
	ActionLogout = &AuditAction{value: "LOGOUT"}
	ActionExport = &AuditAction{value: "EXPORT"}
)

var validActions = map[string]bool{
	"CREATE": true,
	"UPDATE": true,
	"DELETE": true,
	"READ":   true,
	"LOGIN":  true,
	"LOGOUT": true,
	"EXPORT": true,
}

// NewAuditAction creates a new AuditAction with validation
func NewAuditAction(action string) (*AuditAction, error) {
	if action == "" {
		return nil, fmt.Errorf("audit action cannot be empty")
	}
	if !validActions[action] {
		return nil, fmt.Errorf("invalid audit action: %s", action)
	}
	return &AuditAction{value: action}, nil
}

// Value returns the string value of the action
func (a *AuditAction) Value() string {
	if a == nil {
		return ""
	}
	return a.value
}

// String implements Stringer interface
func (a *AuditAction) String() string {
	return a.Value()
}

// Equals checks if two actions are equal
func (a *AuditAction) Equals(other *AuditAction) bool {
	if a == nil && other == nil {
		return true
	}
	if a == nil || other == nil {
		return false
	}
	return a.value == other.value
}

// AdminID is a value object representing an admin identifier
type AdminID struct {
	value string
}

// NewAdminID creates a new AdminID with validation
func NewAdminID(id string) (*AdminID, error) {
	if id == "" {
		return nil, fmt.Errorf("admin ID cannot be empty")
	}
	return &AdminID{value: id}, nil
}

// Value returns the string value of the admin ID
func (a *AdminID) Value() string {
	if a == nil {
		return ""
	}
	return a.value
}

// String implements Stringer interface
func (a *AdminID) String() string {
	return a.Value()
}

// Equals checks if two AdminIDs are equal
func (a *AdminID) Equals(other *AdminID) bool {
	if a == nil && other == nil {
		return true
	}
	if a == nil || other == nil {
		return false
	}
	return a.value == other.value
}

// LogStatus is a value object representing the status of an audit log entry
type LogStatus struct {
	value string
}

// LogStatus constants
var (
	StatusSuccess = &LogStatus{value: "SUCCESS"}
	StatusFailure = &LogStatus{value: "FAILURE"}
	StatusPending = &LogStatus{value: "PENDING"}
)

var validStatuses = map[string]bool{
	"SUCCESS": true,
	"FAILURE": true,
	"PENDING": true,
}

// NewLogStatus creates a new LogStatus with validation
func NewLogStatus(status string) (*LogStatus, error) {
	if status == "" {
		return nil, fmt.Errorf("log status cannot be empty")
	}
	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid log status: %s", status)
	}
	return &LogStatus{value: status}, nil
}

// Value returns the string value of the status
func (s *LogStatus) Value() string {
	if s == nil {
		return ""
	}
	return s.value
}

// String implements Stringer interface
func (s *LogStatus) String() string {
	return s.Value()
}

// IsSuccess checks if the status is SUCCESS
func (s *LogStatus) IsSuccess() bool {
	return s != nil && s.value == "SUCCESS"
}

// IsFailure checks if the status is FAILURE
func (s *LogStatus) IsFailure() bool {
	return s != nil && s.value == "FAILURE"
}

// Equals checks if two statuses are equal
func (s *LogStatus) Equals(other *LogStatus) bool {
	if s == nil && other == nil {
		return true
	}
	if s == nil || other == nil {
		return false
	}
	return s.value == other.value
}

// AuditLog is a domain entity representing an audit log entry
type AuditLog struct {
	id          string    // unique identifier
	timestamp   time.Time // when the action occurred
	action      *AuditAction
	adminID     *AdminID
	ipAddress   string
	userAgent   string
	resourceID  string
	description string
	status      *LogStatus
	errorMsg    string
	details     map[string]interface{}
}

// NewAuditLog creates a new AuditLog with business rule validation
func NewAuditLog(
	id string,
	timestamp time.Time,
	action *AuditAction,
	adminID *AdminID,
	ipAddress string,
	userAgent string,
	resourceID string,
	description string,
	status *LogStatus,
	errorMsg string,
	details map[string]interface{},
) (*AuditLog, error) {
	// Business invariants validation
	if id == "" {
		return nil, fmt.Errorf("audit log ID cannot be empty")
	}
	if action == nil || action.Value() == "" {
		return nil, fmt.Errorf("audit action is required")
	}
	if adminID == nil || adminID.Value() == "" {
		return nil, fmt.Errorf("admin ID is required")
	}
	if status == nil || status.Value() == "" {
		return nil, fmt.Errorf("log status is required")
	}
	if timestamp.IsZero() {
		return nil, fmt.Errorf("timestamp cannot be empty")
	}
	// If status is FAILURE, error message should be present
	if status.IsFailure() && errorMsg == "" {
		return nil, fmt.Errorf("error message is required for failed audit logs")
	}

	if details == nil {
		details = make(map[string]interface{})
	}

	return &AuditLog{
		id:          id,
		timestamp:   timestamp,
		action:      action,
		adminID:     adminID,
		ipAddress:   ipAddress,
		userAgent:   userAgent,
		resourceID:  resourceID,
		description: description,
		status:      status,
		errorMsg:    errorMsg,
		details:     details,
	}, nil
}

// Getters (exposing immutable access)
func (al *AuditLog) ID() string {
	return al.id
}

func (al *AuditLog) Timestamp() time.Time {
	return al.timestamp
}

func (al *AuditLog) Action() *AuditAction {
	return al.action
}

func (al *AuditLog) AdminID() *AdminID {
	return al.adminID
}

func (al *AuditLog) IPAddress() string {
	return al.ipAddress
}

func (al *AuditLog) UserAgent() string {
	return al.userAgent
}

func (al *AuditLog) ResourceID() string {
	return al.resourceID
}

func (al *AuditLog) Description() string {
	return al.description
}

func (al *AuditLog) Status() *LogStatus {
	return al.status
}

func (al *AuditLog) ErrorMsg() string {
	return al.errorMsg
}

func (al *AuditLog) Details() map[string]interface{} {
	if al.details == nil {
		return make(map[string]interface{})
	}
	// Return a copy to prevent external modification
	details := make(map[string]interface{})
	for k, v := range al.details {
		details[k] = v
	}
	return details
}

// IsCritical is a business logic method - returns true if action failed or is a sensitive operation
func (al *AuditLog) IsCritical() bool {
	if al.status.IsFailure() {
		return true
	}
	// Treat DELETE and UPDATE operations as critical
	return al.action.Equals(ActionDelete) || al.action.Equals(ActionUpdate)
}

// IsRecent checks if the log entry is within the last duration
func (al *AuditLog) IsRecent(duration time.Duration) bool {
	return time.Since(al.timestamp) <= duration
}
