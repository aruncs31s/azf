package authorization_audit

import (
	"fmt"
	"time"
)

// WebhookEventType is a value object representing types of webhook events
type WebhookEventType struct {
	value string
}

// Predefined webhook event types
var (
	EventTypeAuthorizationGranted = &WebhookEventType{value: "authorization.granted"}
	EventTypeAuthorizationDenied  = &WebhookEventType{value: "authorization.denied"}
	EventTypeAuditLogCreated      = &WebhookEventType{value: "audit.log.created"}
	EventTypeAdminLogin           = &WebhookEventType{value: "admin.login"}
	EventTypeAdminLogout          = &WebhookEventType{value: "admin.logout"}
	EventTypeResourceAccessed     = &WebhookEventType{value: "resource.accessed"}
	EventTypeResourceModified     = &WebhookEventType{value: "resource.modified"}
	EventTypeResourceDeleted      = &WebhookEventType{value: "resource.deleted"}
	EventTypePolicyViolation      = &WebhookEventType{value: "policy.violation"}
)

var validEventTypes = map[string]bool{
	"authorization.granted":   true,
	"authorization.denied":    true,
	"audit.log.created":       true,
	"admin.login":             true,
	"admin.logout":            true,
	"resource.accessed":       true,
	"resource.modified":       true,
	"resource.deleted":        true,
	"policy.violation":        true,
}

// NewWebhookEventType creates a new WebhookEventType with validation
func NewWebhookEventType(eventType string) (*WebhookEventType, error) {
	if eventType == "" {
		return nil, fmt.Errorf("webhook event type cannot be empty")
	}
	if !validEventTypes[eventType] {
		return nil, fmt.Errorf("invalid webhook event type: %s", eventType)
	}
	return &WebhookEventType{value: eventType}, nil
}

// Value returns the string value of the event type
func (w *WebhookEventType) Value() string {
	if w == nil {
		return ""
	}
	return w.value
}

// String implements Stringer interface
func (w *WebhookEventType) String() string {
	return w.Value()
}

// Equals checks if two event types are equal
func (w *WebhookEventType) Equals(other *WebhookEventType) bool {
	if w == nil && other == nil {
		return true
	}
	if w == nil || other == nil {
		return false
	}
	return w.value == other.value
}

// WebhookEventStatus is a value object representing webhook delivery status
type WebhookEventStatus struct {
	value string
}

// Predefined webhook event statuses
var (
	WebhookStatusPending      = &WebhookEventStatus{value: "PENDING"}
	WebhookStatusDelivered    = &WebhookEventStatus{value: "DELIVERED"}
	WebhookStatusFailed       = &WebhookEventStatus{value: "FAILED"}
	WebhookStatusRetrying     = &WebhookEventStatus{value: "RETRYING"}
	WebhookStatusAbandoned    = &WebhookEventStatus{value: "ABANDONED"}
)

var validWebhookStatuses = map[string]bool{
	"PENDING":    true,
	"DELIVERED":  true,
	"FAILED":     true,
	"RETRYING":   true,
	"ABANDONED":  true,
}

// NewWebhookEventStatus creates a new WebhookEventStatus with validation
func NewWebhookEventStatus(status string) (*WebhookEventStatus, error) {
	if status == "" {
		return nil, fmt.Errorf("webhook event status cannot be empty")
	}
	if !validWebhookStatuses[status] {
		return nil, fmt.Errorf("invalid webhook event status: %s", status)
	}
	return &WebhookEventStatus{value: status}, nil
}

// Value returns the string value of the status
func (w *WebhookEventStatus) Value() string {
	if w == nil {
		return ""
	}
	return w.value
}

// String implements Stringer interface
func (w *WebhookEventStatus) String() string {
	return w.Value()
}

// IsDelivered checks if the event was successfully delivered
func (w *WebhookEventStatus) IsDelivered() bool {
	return w != nil && w.value == "DELIVERED"
}

// IsFailed checks if the event delivery failed
func (w *WebhookEventStatus) IsFailed() bool {
	return w != nil && w.value == "FAILED"
}

// IsRetrying checks if the event delivery is being retried
func (w *WebhookEventStatus) IsRetrying() bool {
	return w != nil && w.value == "RETRYING"
}

// Equals checks if two statuses are equal
func (w *WebhookEventStatus) Equals(other *WebhookEventStatus) bool {
	if w == nil && other == nil {
		return true
	}
	if w == nil || other == nil {
		return false
	}
	return w.value == other.value
}

// WebhookEvent is a domain entity representing an authorization event to be sent via webhook
type WebhookEvent struct {
	id           string
	eventType    *WebhookEventType
	auditLogID   string
	payload      map[string]interface{}
	timestamp    time.Time
	status       *WebhookEventStatus
	deliveryURL  string
	retryCount   int
	maxRetries   int
	lastError    string
	lastAttempt  *time.Time
	nextRetry    *time.Time
	metadata     map[string]interface{}
}

// NewWebhookEvent creates a new WebhookEvent with business rule validation
func NewWebhookEvent(
	id string,
	eventType *WebhookEventType,
	auditLogID string,
	payload map[string]interface{},
	timestamp time.Time,
	deliveryURL string,
) (*WebhookEvent, error) {
	if id == "" {
		return nil, fmt.Errorf("webhook event ID cannot be empty")
	}
	if eventType == nil || eventType.Value() == "" {
		return nil, fmt.Errorf("webhook event type is required")
	}
	if auditLogID == "" {
		return nil, fmt.Errorf("audit log ID is required")
	}
	if timestamp.IsZero() {
		return nil, fmt.Errorf("timestamp cannot be empty")
	}
	if deliveryURL == "" {
		return nil, fmt.Errorf("delivery URL is required")
	}
	if payload == nil {
		payload = make(map[string]interface{})
	}

	return &WebhookEvent{
		id:          id,
		eventType:   eventType,
		auditLogID:  auditLogID,
		payload:     payload,
		timestamp:   timestamp,
		status:      WebhookStatusPending,
		deliveryURL: deliveryURL,
		retryCount:  0,
		maxRetries:  3,
		metadata:    make(map[string]interface{}),
	}, nil
}

// Getters
func (w *WebhookEvent) ID() string {
	return w.id
}

func (w *WebhookEvent) EventType() *WebhookEventType {
	return w.eventType
}

func (w *WebhookEvent) AuditLogID() string {
	return w.auditLogID
}

func (w *WebhookEvent) Payload() map[string]interface{} {
	if w.payload == nil {
		return make(map[string]interface{})
	}
	payload := make(map[string]interface{})
	for k, v := range w.payload {
		payload[k] = v
	}
	return payload
}

func (w *WebhookEvent) Timestamp() time.Time {
	return w.timestamp
}

func (w *WebhookEvent) Status() *WebhookEventStatus {
	return w.status
}

func (w *WebhookEvent) DeliveryURL() string {
	return w.deliveryURL
}

func (w *WebhookEvent) RetryCount() int {
	return w.retryCount
}

func (w *WebhookEvent) MaxRetries() int {
	return w.maxRetries
}

func (w *WebhookEvent) LastError() string {
	return w.lastError
}

func (w *WebhookEvent) LastAttempt() *time.Time {
	return w.lastAttempt
}

func (w *WebhookEvent) NextRetry() *time.Time {
	return w.nextRetry
}

func (w *WebhookEvent) Metadata() map[string]interface{} {
	if w.metadata == nil {
		return make(map[string]interface{})
	}
	metadata := make(map[string]interface{})
	for k, v := range w.metadata {
		metadata[k] = v
	}
	return metadata
}

// Business logic methods

// CanRetry checks if the event can be retried
func (w *WebhookEvent) CanRetry() bool {
	if w.retryCount >= w.maxRetries {
		return false
	}
	return w.status.IsFailed() || w.status.IsRetrying()
}

// MarkAsDelivered marks the event as successfully delivered
func (w *WebhookEvent) MarkAsDelivered() error {
	if w.status == nil {
		return fmt.Errorf("invalid webhook event status")
	}
	w.status = WebhookStatusDelivered
	now := time.Now()
	w.lastAttempt = &now
	w.lastError = ""
	return nil
}

// MarkAsFailed marks the event delivery as failed
func (w *WebhookEvent) MarkAsFailed(errorMsg string) error {
	if w.status == nil {
		return fmt.Errorf("invalid webhook event status")
	}
	if errorMsg == "" {
		return fmt.Errorf("error message is required")
	}
	w.status = WebhookStatusFailed
	w.lastError = errorMsg
	now := time.Now()
	w.lastAttempt = &now
	return nil
}

// MarkForRetry marks the event for retry with exponential backoff
func (w *WebhookEvent) MarkForRetry(errorMsg string) error {
	if w.retryCount >= w.maxRetries {
		w.status = WebhookStatusAbandoned
		w.lastError = errorMsg
		return fmt.Errorf("max retries exceeded")
	}

	w.retryCount++
	w.status = WebhookStatusRetrying
	w.lastError = errorMsg
	now := time.Now()
	w.lastAttempt = &now

	// Exponential backoff: 2^retryCount minutes
	backoffMinutes := 1 << uint(w.retryCount)
	nextRetry := now.Add(time.Duration(backoffMinutes) * time.Minute)
	w.nextRetry = &nextRetry

	return nil
}

// IsRetryable checks if the event should be retried
func (w *WebhookEvent) IsRetryable() bool {
	return w.CanRetry() && w.nextRetry != nil && time.Now().After(*w.nextRetry)
}

// GetDeliveryAttempt returns the formatted attempt information
func (w *WebhookEvent) GetDeliveryAttempt() map[string]interface{} {
	attempt := map[string]interface{}{
		"retry_count": w.retryCount,
		"max_retries": w.maxRetries,
	}
	if w.lastAttempt != nil {
		attempt["last_attempt"] = w.lastAttempt
	}
	if w.nextRetry != nil {
		attempt["next_retry"] = w.nextRetry
	}
	return attempt
}
