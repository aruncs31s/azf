package authorization_audit

import "context"

// AuditLogRepository defines the interface for audit log persistence operations
type AuditLogRepository interface {
	AuditLogReader
	AuditLogWriter
}

// AuditLogReader defines the interface for audit log read operations
type AuditLogReader interface {
	// FindByID retrieves an audit log by ID
	FindByID(ctx context.Context, id string) (*AuditLog, error)

	// FindByAdminID retrieves all audit logs for a specific admin
	FindByAdminID(ctx context.Context, adminID string) ([]*AuditLog, error)

	// FindByAction retrieves all audit logs for a specific action
	FindByAction(ctx context.Context, action string) ([]*AuditLog, error)

	// FindByResourceID retrieves all audit logs for a specific resource
	FindByResourceID(ctx context.Context, resourceID string) ([]*AuditLog, error)

	// FindAll retrieves all audit logs
	FindAll(ctx context.Context) ([]*AuditLog, error)
}

// AuditLogWriter defines the interface for audit log write operations
type AuditLogWriter interface {
	// Create adds a new audit log entry
	Create(ctx context.Context, log *AuditLog) (*AuditLog, error)

	// Update modifies an audit log entry
	Update(ctx context.Context, log *AuditLog) (*AuditLog, error)

	// Delete removes an audit log entry
	Delete(ctx context.Context, id string) error
}

// WebhookEventRepository defines the interface for webhook event persistence operations
type WebhookEventRepository interface {
	WebhookEventReader
	WebhookEventWriter
}

// WebhookEventReader defines the interface for webhook event read operations
type WebhookEventReader interface {
	// FindByID retrieves a webhook event by ID
	FindByID(ctx context.Context, id string) (*WebhookEvent, error)

	// FindByAuditLogID retrieves webhook events for a specific audit log
	FindByAuditLogID(ctx context.Context, auditLogID string) ([]*WebhookEvent, error)

	// FindByEventType retrieves webhook events of a specific type
	FindByEventType(ctx context.Context, eventType string) ([]*WebhookEvent, error)

	// FindByStatus retrieves webhook events with a specific delivery status
	FindByStatus(ctx context.Context, status string) ([]*WebhookEvent, error)

	// FindPending retrieves all pending webhook events
	FindPending(ctx context.Context) ([]*WebhookEvent, error)

	// FindRetryable retrieves webhook events that can be retried
	FindRetryable(ctx context.Context) ([]*WebhookEvent, error)

	// FindAll retrieves all webhook events
	FindAll(ctx context.Context) ([]*WebhookEvent, error)
}

// WebhookEventWriter defines the interface for webhook event write operations
type WebhookEventWriter interface {
	// Create adds a new webhook event
	Create(ctx context.Context, event *WebhookEvent) (*WebhookEvent, error)

	// Update modifies a webhook event
	Update(ctx context.Context, event *WebhookEvent) (*WebhookEvent, error)

	// Delete removes a webhook event
	Delete(ctx context.Context, id string) error

	// BulkCreate adds multiple webhook events
	BulkCreate(ctx context.Context, events []*WebhookEvent) ([]*WebhookEvent, error)
}

// WebhookSubscriptionRepository defines the interface for webhook subscription persistence operations
type WebhookSubscriptionRepository interface {
	WebhookSubscriptionReader
	WebhookSubscriptionWriter
}

// WebhookSubscriptionReader defines the interface for webhook subscription read operations
type WebhookSubscriptionReader interface {
	// FindByID retrieves a webhook subscription by ID
	FindByID(ctx context.Context, id string) (*WebhookSubscription, error)

	// FindByEndpoint retrieves subscriptions by endpoint URL
	FindByEndpoint(ctx context.Context, endpoint string) ([]*WebhookSubscription, error)

	// FindByStatus retrieves subscriptions with a specific status
	FindByStatus(ctx context.Context, status string) ([]*WebhookSubscription, error)

	// FindActive retrieves all active subscriptions
	FindActive(ctx context.Context) ([]*WebhookSubscription, error)

	// FindByEventType retrieves subscriptions interested in a specific event type
	FindByEventType(ctx context.Context, eventType string) ([]*WebhookSubscription, error)

	// FindAll retrieves all subscriptions
	FindAll(ctx context.Context) ([]*WebhookSubscription, error)
}

// WebhookSubscriptionWriter defines the interface for webhook subscription write operations
type WebhookSubscriptionWriter interface {
	// Create adds a new webhook subscription
	Create(ctx context.Context, subscription *WebhookSubscription) (*WebhookSubscription, error)

	// Update modifies a webhook subscription
	Update(ctx context.Context, subscription *WebhookSubscription) (*WebhookSubscription, error)

	// Delete removes a webhook subscription
	Delete(ctx context.Context, id string) error
}
