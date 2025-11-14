package authorization_audit

import (
	"context"
	"fmt"
	"time"
)

// WebhookEndpoint is a value object representing a webhook endpoint URL
type WebhookEndpoint struct {
	value string
}

// NewWebhookEndpoint creates a new WebhookEndpoint with validation
func NewWebhookEndpoint(url string) (*WebhookEndpoint, error) {
	if url == "" {
		return nil, fmt.Errorf("webhook endpoint URL cannot be empty")
	}
	if len(url) > 2048 {
		return nil, fmt.Errorf("webhook endpoint URL too long (max 2048 characters)")
	}
	return &WebhookEndpoint{value: url}, nil
}

// Value returns the string value of the endpoint
func (w *WebhookEndpoint) Value() string {
	if w == nil {
		return ""
	}
	return w.value
}

// String implements Stringer interface
func (w *WebhookEndpoint) String() string {
	return w.Value()
}

// Equals checks if two endpoints are equal
func (w *WebhookEndpoint) Equals(other *WebhookEndpoint) bool {
	if w == nil && other == nil {
		return true
	}
	if w == nil || other == nil {
		return false
	}
	return w.value == other.value
}

// SubscriptionStatus is a value object representing webhook subscription status
type SubscriptionStatus struct {
	value string
}

// Predefined subscription statuses
var (
	SubscriptionStatusActive   = &SubscriptionStatus{value: "ACTIVE"}
	SubscriptionStatusInactive = &SubscriptionStatus{value: "INACTIVE"}
	SubscriptionStatusSuspended = &SubscriptionStatus{value: "SUSPENDED"}
)

var validSubscriptionStatuses = map[string]bool{
	"ACTIVE":    true,
	"INACTIVE":  true,
	"SUSPENDED": true,
}

// NewSubscriptionStatus creates a new SubscriptionStatus with validation
func NewSubscriptionStatus(status string) (*SubscriptionStatus, error) {
	if status == "" {
		return nil, fmt.Errorf("subscription status cannot be empty")
	}
	if !validSubscriptionStatuses[status] {
		return nil, fmt.Errorf("invalid subscription status: %s", status)
	}
	return &SubscriptionStatus{value: status}, nil
}

// Value returns the string value of the status
func (s *SubscriptionStatus) Value() string {
	if s == nil {
		return ""
	}
	return s.value
}

// String implements Stringer interface
func (s *SubscriptionStatus) String() string {
	return s.Value()
}

// IsActive checks if the subscription is active
func (s *SubscriptionStatus) IsActive() bool {
	return s != nil && s.value == "ACTIVE"
}

// Equals checks if two statuses are equal
func (s *SubscriptionStatus) Equals(other *SubscriptionStatus) bool {
	if s == nil && other == nil {
		return true
	}
	if s == nil || other == nil {
		return false
	}
	return s.value == other.value
}

// WebhookSubscription is a domain entity representing a webhook subscription
type WebhookSubscription struct {
	id              string
	endpoint        *WebhookEndpoint
	eventTypes      []*WebhookEventType
	status          *SubscriptionStatus
	secret          string
	description     string
	createdAt       time.Time
	updatedAt       time.Time
	lastDelivery    *time.Time
	failureCount    int
	maxFailures     int
	filters         map[string]interface{}
	headers         map[string]string
	metadata        map[string]interface{}
}

// NewWebhookSubscription creates a new WebhookSubscription with business rule validation
func NewWebhookSubscription(
	id string,
	endpoint *WebhookEndpoint,
	eventTypes []*WebhookEventType,
	secret string,
	description string,
) (*WebhookSubscription, error) {
	if id == "" {
		return nil, fmt.Errorf("webhook subscription ID cannot be empty")
	}
	if endpoint == nil || endpoint.Value() == "" {
		return nil, fmt.Errorf("webhook endpoint is required")
	}
	if len(eventTypes) == 0 {
		return nil, fmt.Errorf("at least one event type is required")
	}
	if len(eventTypes) > 10 {
		return nil, fmt.Errorf("maximum 10 event types allowed per subscription")
	}
	if secret == "" {
		return nil, fmt.Errorf("webhook secret is required")
	}
	if len(secret) < 32 {
		return nil, fmt.Errorf("webhook secret must be at least 32 characters")
	}

	return &WebhookSubscription{
		id:          id,
		endpoint:    endpoint,
		eventTypes:  eventTypes,
		status:      SubscriptionStatusActive,
		secret:      secret,
		description: description,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
		maxFailures: 5,
		filters:     make(map[string]interface{}),
		headers:     make(map[string]string),
		metadata:    make(map[string]interface{}),
	}, nil
}

// Getters
func (w *WebhookSubscription) ID() string {
	return w.id
}

func (w *WebhookSubscription) Endpoint() *WebhookEndpoint {
	return w.endpoint
}

func (w *WebhookSubscription) EventTypes() []*WebhookEventType {
	if w.eventTypes == nil {
		return []*WebhookEventType{}
	}
	types := make([]*WebhookEventType, len(w.eventTypes))
	copy(types, w.eventTypes)
	return types
}

func (w *WebhookSubscription) Status() *SubscriptionStatus {
	return w.status
}

func (w *WebhookSubscription) Secret() string {
	return w.secret
}

func (w *WebhookSubscription) Description() string {
	return w.description
}

func (w *WebhookSubscription) CreatedAt() time.Time {
	return w.createdAt
}

func (w *WebhookSubscription) UpdatedAt() time.Time {
	return w.updatedAt
}

func (w *WebhookSubscription) LastDelivery() *time.Time {
	return w.lastDelivery
}

func (w *WebhookSubscription) FailureCount() int {
	return w.failureCount
}

func (w *WebhookSubscription) MaxFailures() int {
	return w.maxFailures
}

func (w *WebhookSubscription) Filters() map[string]interface{} {
	if w.filters == nil {
		return make(map[string]interface{})
	}
	filters := make(map[string]interface{})
	for k, v := range w.filters {
		filters[k] = v
	}
	return filters
}

func (w *WebhookSubscription) Headers() map[string]string {
	if w.headers == nil {
		return make(map[string]string)
	}
	headers := make(map[string]string)
	for k, v := range w.headers {
		headers[k] = v
	}
	return headers
}

func (w *WebhookSubscription) Metadata() map[string]interface{} {
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

// IsSubscribedTo checks if the subscription is interested in a specific event type
func (w *WebhookSubscription) IsSubscribedTo(eventType *WebhookEventType) bool {
	if !w.status.IsActive() {
		return false
	}
	if eventType == nil {
		return false
	}
	for _, et := range w.eventTypes {
		if et.Equals(eventType) {
			return true
		}
	}
	return false
}

// CanDeliver checks if the subscription can deliver events
func (w *WebhookSubscription) CanDeliver() bool {
	return w.status.IsActive() && w.failureCount < w.maxFailures
}

// AddEventType adds a new event type to the subscription
func (w *WebhookSubscription) AddEventType(eventType *WebhookEventType) error {
	if eventType == nil {
		return fmt.Errorf("event type cannot be nil")
	}
	if len(w.eventTypes) >= 10 {
		return fmt.Errorf("maximum 10 event types allowed")
	}
	for _, et := range w.eventTypes {
		if et.Equals(eventType) {
			return fmt.Errorf("event type already subscribed")
		}
	}
	w.eventTypes = append(w.eventTypes, eventType)
	w.updatedAt = time.Now()
	return nil
}

// RemoveEventType removes an event type from the subscription
func (w *WebhookSubscription) RemoveEventType(eventType *WebhookEventType) error {
	if eventType == nil {
		return fmt.Errorf("event type cannot be nil")
	}
	if len(w.eventTypes) == 1 {
		return fmt.Errorf("cannot remove the last event type")
	}
	for i, et := range w.eventTypes {
		if et.Equals(eventType) {
			w.eventTypes = append(w.eventTypes[:i], w.eventTypes[i+1:]...)
			w.updatedAt = time.Now()
			return nil
		}
	}
	return fmt.Errorf("event type not found in subscription")
}

// SetFilter sets a filter for event delivery
func (w *WebhookSubscription) SetFilter(key string, value interface{}) error {
	if key == "" {
		return fmt.Errorf("filter key cannot be empty")
	}
	w.filters[key] = value
	w.updatedAt = time.Now()
	return nil
}

// SetHeader sets a custom header for webhook delivery
func (w *WebhookSubscription) SetHeader(key string, value string) error {
	if key == "" {
		return fmt.Errorf("header key cannot be empty")
	}
	w.headers[key] = value
	w.updatedAt = time.Now()
	return nil
}

// Activate activates the subscription
func (w *WebhookSubscription) Activate() error {
	w.status = SubscriptionStatusActive
	w.failureCount = 0
	w.updatedAt = time.Now()
	return nil
}

// Deactivate deactivates the subscription
func (w *WebhookSubscription) Deactivate() error {
	w.status = SubscriptionStatusInactive
	w.updatedAt = time.Now()
	return nil
}

// Suspend suspends the subscription
func (w *WebhookSubscription) Suspend() error {
	w.status = SubscriptionStatusSuspended
	w.updatedAt = time.Now()
	return nil
}

// RecordDelivery records a successful delivery attempt
func (w *WebhookSubscription) RecordDelivery() error {
	w.failureCount = 0
	now := time.Now()
	w.lastDelivery = &now
	w.updatedAt = now
	return nil
}

// RecordFailure records a delivery failure
func (w *WebhookSubscription) RecordFailure() error {
	w.failureCount++
	if w.failureCount >= w.maxFailures {
		w.status = SubscriptionStatusSuspended
	}
	w.updatedAt = time.Now()
	return nil
}

// IsHealthy checks if the subscription is healthy
func (w *WebhookSubscription) IsHealthy() bool {
	return w.status.IsActive() && w.failureCount < (w.maxFailures / 2)
}

// WebhookPublisher defines the interface for publishing webhook events
type WebhookPublisher interface {
	// PublishAuthorizationGranted publishes an authorization granted event
	PublishAuthorizationGranted(ctx context.Context, auditLog *AuditLog) error

	// PublishAuthorizationDenied publishes an authorization denied event
	PublishAuthorizationDenied(ctx context.Context, auditLog *AuditLog) error

	// PublishAuditLogCreated publishes an audit log created event
	PublishAuditLogCreated(ctx context.Context, auditLog *AuditLog) error

	// PublishAdminLogin publishes an admin login event
	PublishAdminLogin(ctx context.Context, auditLog *AuditLog) error

	// PublishAdminLogout publishes an admin logout event
	PublishAdminLogout(ctx context.Context, auditLog *AuditLog) error

	// PublishResourceAccessed publishes a resource accessed event
	PublishResourceAccessed(ctx context.Context, auditLog *AuditLog) error

	// PublishResourceModified publishes a resource modified event
	PublishResourceModified(ctx context.Context, auditLog *AuditLog) error

	// PublishResourceDeleted publishes a resource deleted event
	PublishResourceDeleted(ctx context.Context, auditLog *AuditLog) error

	// PublishPolicyViolation publishes a policy violation event
	PublishPolicyViolation(ctx context.Context, auditLog *AuditLog) error
}

// WebhookDispatcher defines the interface for delivering webhook events
type WebhookDispatcher interface {
	// DispatchEvent sends a webhook event to all subscribed endpoints
	DispatchEvent(ctx context.Context, event *WebhookEvent) error

	// DispatchPending sends all pending webhook events
	DispatchPending(ctx context.Context) error

	// RetryFailed retries failed webhook events
	RetryFailed(ctx context.Context) error

	// GetEventDeliveryStatus retrieves the delivery status of an event
	GetEventDeliveryStatus(ctx context.Context, eventID string) (*WebhookEvent, error)
}

// WebhookManager defines the interface for managing webhook subscriptions
type WebhookManager interface {
	// RegisterSubscription creates a new webhook subscription
	RegisterSubscription(ctx context.Context, subscription *WebhookSubscription) (*WebhookSubscription, error)

	// UnregisterSubscription removes a webhook subscription
	UnregisterSubscription(ctx context.Context, subscriptionID string) error

	// UpdateSubscription updates a webhook subscription
	UpdateSubscription(ctx context.Context, subscription *WebhookSubscription) (*WebhookSubscription, error)

	// GetSubscription retrieves a webhook subscription
	GetSubscription(ctx context.Context, subscriptionID string) (*WebhookSubscription, error)

	// ListSubscriptions lists all webhook subscriptions
	ListSubscriptions(ctx context.Context) ([]*WebhookSubscription, error)

	// ListActiveSubscriptions lists active webhook subscriptions
	ListActiveSubscriptions(ctx context.Context) ([]*WebhookSubscription, error)

	// VerifySubscription verifies a webhook subscription endpoint
	VerifySubscription(ctx context.Context, subscriptionID string) error

	// TestSubscription sends a test event to the subscription
	TestSubscription(ctx context.Context, subscriptionID string) error
}

// WebhookEventBuilder defines the interface for building webhook events from authorization events
type WebhookEventBuilder interface {
	// BuildFromAuditLog creates webhook events from an audit log entry
	BuildFromAuditLog(auditLog *AuditLog) ([]*WebhookEvent, error)

	// FilterBySubscriptions filters webhook events to only active subscriptions
	FilterBySubscriptions(ctx context.Context, eventType *WebhookEventType) ([]*WebhookSubscription, error)
}
