package authorization_audit

import (
	"testing"
	"time"
)

// TestWebhookEventTypeCreation tests creating webhook event types
func TestWebhookEventTypeCreation(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		wantErr   bool
	}{
		{"valid authorization granted", "authorization.granted", false},
		{"valid authorization denied", "authorization.denied", false},
		{"valid audit log created", "audit.log.created", false},
		{"empty event type", "", true},
		{"invalid event type", "invalid.event", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et, err := NewWebhookEventType(tt.eventType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWebhookEventType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && et.Value() != tt.eventType {
				t.Errorf("expected %s, got %s", tt.eventType, et.Value())
			}
		})
	}
}

// TestWebhookEventCreation tests creating webhook events
func TestWebhookEventCreation(t *testing.T) {
	endpoint := "https://example.com/webhooks"
	eventType := EventTypeAuthorizationGranted
	auditLogID := "audit-123"
	payload := map[string]interface{}{"action": "grant"}
	timestamp := time.Now()

	event, err := NewWebhookEvent("webhook-1", eventType, auditLogID, payload, timestamp, endpoint)
	if err != nil {
		t.Fatalf("NewWebhookEvent() error = %v", err)
	}

	if event.ID() != "webhook-1" {
		t.Errorf("expected ID webhook-1, got %s", event.ID())
	}
	if !event.EventType().Equals(eventType) {
		t.Errorf("expected event type %s, got %s", eventType.Value(), event.EventType().Value())
	}
	if event.Status().Value() != "PENDING" {
		t.Errorf("expected status PENDING, got %s", event.Status().Value())
	}
}

// TestWebhookEventRetry tests webhook event retry logic
func TestWebhookEventRetry(t *testing.T) {
	event, _ := NewWebhookEvent("webhook-1", EventTypeAuthorizationGranted, "audit-123",
		map[string]interface{}{}, time.Now(), "https://example.com")

	// Mark as failed
	err := event.MarkForRetry("connection timeout")
	if err != nil {
		t.Fatalf("MarkForRetry() error = %v", err)
	}

	if !event.Status().IsRetrying() {
		t.Error("expected status RETRYING")
	}
	if event.RetryCount() != 1 {
		t.Errorf("expected retry count 1, got %d", event.RetryCount())
	}

	// Check retry attempt info
	attempt := event.GetDeliveryAttempt()
	if attempt["retry_count"] != 1 {
		t.Error("expected retry count in attempt info")
	}
	if attempt["next_retry"] == nil {
		t.Error("expected next retry time in attempt info")
	}
}

// TestWebhookEventDelivery tests webhook delivery marking
func TestWebhookEventDelivery(t *testing.T) {
	event, _ := NewWebhookEvent("webhook-1", EventTypeAuthorizationGranted, "audit-123",
		map[string]interface{}{}, time.Now(), "https://example.com")

	err := event.MarkAsDelivered()
	if err != nil {
		t.Fatalf("MarkAsDelivered() error = %v", err)
	}

	if !event.Status().IsDelivered() {
		t.Error("expected status DELIVERED")
	}
	if event.LastError() != "" {
		t.Error("expected no error message after delivery")
	}
}

// TestWebhookSubscriptionCreation tests creating webhook subscriptions
func TestWebhookSubscriptionCreation(t *testing.T) {
	endpoint, _ := NewWebhookEndpoint("https://example.com/webhooks")
	eventTypes := []*WebhookEventType{EventTypeAuthorizationGranted}

	sub, err := NewWebhookSubscription("sub-1", endpoint, eventTypes,
		"super-secret-32-character-minimum-key-here", "Test subscription")
	if err != nil {
		t.Fatalf("NewWebhookSubscription() error = %v", err)
	}

	if sub.ID() != "sub-1" {
		t.Errorf("expected ID sub-1, got %s", sub.ID())
	}
	if !sub.Status().IsActive() {
		t.Error("expected status ACTIVE")
	}
}

// TestWebhookSubscriptionEventFiltering tests subscription event filtering
func TestWebhookSubscriptionEventFiltering(t *testing.T) {
	endpoint, _ := NewWebhookEndpoint("https://example.com/webhooks")
	eventTypes := []*WebhookEventType{EventTypeAuthorizationGranted}

	sub, _ := NewWebhookSubscription("sub-1", endpoint, eventTypes,
		"super-secret-32-character-minimum-key-here", "Test")

	// Should be subscribed to granted
	if !sub.IsSubscribedTo(EventTypeAuthorizationGranted) {
		t.Error("should be subscribed to authorization.granted")
	}

	// Should not be subscribed to denied
	if sub.IsSubscribedTo(EventTypeAuthorizationDenied) {
		t.Error("should not be subscribed to authorization.denied")
	}
}

// TestWebhookSubscriptionEventTypeManagement tests adding/removing event types
func TestWebhookSubscriptionEventTypeManagement(t *testing.T) {
	endpoint, _ := NewWebhookEndpoint("https://example.com/webhooks")
	eventTypes := []*WebhookEventType{EventTypeAuthorizationGranted}

	sub, _ := NewWebhookSubscription("sub-1", endpoint, eventTypes,
		"super-secret-32-character-minimum-key-here", "Test")

	// Add event type
	err := sub.AddEventType(EventTypeAuthorizationDenied)
	if err != nil {
		t.Fatalf("AddEventType() error = %v", err)
	}

	if len(sub.EventTypes()) != 2 {
		t.Errorf("expected 2 event types, got %d", len(sub.EventTypes()))
	}

	// Remove event type
	err = sub.RemoveEventType(EventTypeAuthorizationGranted)
	if err != nil {
		t.Fatalf("RemoveEventType() error = %v", err)
	}

	if len(sub.EventTypes()) != 1 {
		t.Errorf("expected 1 event type after removal, got %d", len(sub.EventTypes()))
	}
}

// TestWebhookSubscriptionFailureTracking tests subscription failure tracking
func TestWebhookSubscriptionFailureTracking(t *testing.T) {
	endpoint, _ := NewWebhookEndpoint("https://example.com/webhooks")
	eventTypes := []*WebhookEventType{EventTypeAuthorizationGranted}

	sub, _ := NewWebhookSubscription("sub-1", endpoint, eventTypes,
		"super-secret-32-character-minimum-key-here", "Test")

	// Record successful delivery
	err := sub.RecordDelivery()
	if err != nil {
		t.Fatalf("RecordDelivery() error = %v", err)
	}
	if sub.FailureCount() != 0 {
		t.Errorf("expected 0 failures after delivery, got %d", sub.FailureCount())
	}

	// Record failures
	for i := 0; i < 3; i++ {
		err := sub.RecordFailure()
		if err != nil {
			t.Fatalf("RecordFailure() error = %v", err)
		}
	}

	if sub.FailureCount() != 3 {
		t.Errorf("expected 3 failures, got %d", sub.FailureCount())
	}

	// After 5 failures, subscription should be suspended
	for i := 0; i < 2; i++ {
		sub.RecordFailure()
	}

	if sub.FailureCount() != 5 {
		t.Errorf("expected 5 failures, got %d", sub.FailureCount())
	}
	if !sub.Status().Equals(SubscriptionStatusSuspended) {
		t.Errorf("expected status SUSPENDED after max failures, got %s", sub.Status().Value())
	}
}

// TestWebhookSubscriptionHeaders tests custom header management
func TestWebhookSubscriptionHeaders(t *testing.T) {
	endpoint, _ := NewWebhookEndpoint("https://example.com/webhooks")
	eventTypes := []*WebhookEventType{EventTypeAuthorizationGranted}

	sub, _ := NewWebhookSubscription("sub-1", endpoint, eventTypes,
		"super-secret-32-character-minimum-key-here", "Test")

	// Set headers
	err := sub.SetHeader("Authorization", "Bearer token123")
	if err != nil {
		t.Fatalf("SetHeader() error = %v", err)
	}

	err = sub.SetHeader("X-Custom-Header", "custom-value")
	if err != nil {
		t.Fatalf("SetHeader() error = %v", err)
	}

	headers := sub.Headers()
	if headers["Authorization"] != "Bearer token123" {
		t.Error("expected authorization header")
	}
	if headers["X-Custom-Header"] != "custom-value" {
		t.Error("expected custom header")
	}
}

// TestWebhookSubscriptionActivation tests subscription activation/deactivation
func TestWebhookSubscriptionActivation(t *testing.T) {
	endpoint, _ := NewWebhookEndpoint("https://example.com/webhooks")
	eventTypes := []*WebhookEventType{EventTypeAuthorizationGranted}

	sub, _ := NewWebhookSubscription("sub-1", endpoint, eventTypes,
		"super-secret-32-character-minimum-key-here", "Test")

	// Start as active
	if !sub.Status().IsActive() {
		t.Error("expected initial status ACTIVE")
	}

	// Deactivate
	err := sub.Deactivate()
	if err != nil {
		t.Fatalf("Deactivate() error = %v", err)
	}
	if sub.Status().IsActive() {
		t.Error("expected status INACTIVE after deactivation")
	}

	// Reactivate
	err = sub.Activate()
	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	if !sub.Status().IsActive() {
		t.Error("expected status ACTIVE after reactivation")
	}

	// Suspend
	err = sub.Suspend()
	if err != nil {
		t.Fatalf("Suspend() error = %v", err)
	}
	if !sub.Status().Equals(SubscriptionStatusSuspended) {
		t.Error("expected status SUSPENDED after suspension")
	}
}

// TestWebhookSubscriptionHealth tests subscription health checking
func TestWebhookSubscriptionHealth(t *testing.T) {
	endpoint, _ := NewWebhookEndpoint("https://example.com/webhooks")
	eventTypes := []*WebhookEventType{EventTypeAuthorizationGranted}

	sub, _ := NewWebhookSubscription("sub-1", endpoint, eventTypes,
		"super-secret-32-character-minimum-key-here", "Test")

	// Should be healthy initially
	if !sub.IsHealthy() {
		t.Error("expected subscription to be healthy initially")
	}

	// Record one failure (less than max_failures/2 = 2)
	sub.RecordFailure()
	if !sub.IsHealthy() {
		t.Error("expected subscription to be healthy after 1 failure (threshold is 2)")
	}

	// Record another failure (now at threshold)
	sub.RecordFailure()
	if sub.IsHealthy() {
		t.Error("expected subscription to be unhealthy at failure count >= 2")
	}
}
