package model

import (
	"fmt"
	"time"
)

// AuthorizationResult represents the outcome of an authorization check
type AuthorizationResult struct {
	value string
}

var (
	AuthzAllowed = &AuthorizationResult{value: "ALLOWED"}
	AuthzDenied  = &AuthorizationResult{value: "DENIED"}
	AuthzWarning = &AuthorizationResult{value: "WARNING"} // For gradual rollout
)

var validAuthzResults = map[string]bool{
	"ALLOWED": true,
	"DENIED":  true,
	"WARNING": true,
}

func NewAuthorizationResult(result string) (*AuthorizationResult, error) {
	if result == "" {
		return nil, fmt.Errorf("authorization result cannot be empty")
	}
	if !validAuthzResults[result] {
		return nil, fmt.Errorf("invalid authorization result: %s", result)
	}
	return &AuthorizationResult{value: result}, nil
}

func (ar *AuthorizationResult) Value() string {
	if ar == nil {
		return ""
	}
	return ar.value
}

func (ar *AuthorizationResult) String() string {
	return ar.Value()
}

func (ar *AuthorizationResult) IsAllowed() bool {
	return ar != nil && ar.value == "ALLOWED"
}

func (ar *AuthorizationResult) IsDenied() bool {
	return ar != nil && ar.value == "DENIED"
}

func (ar *AuthorizationResult) IsWarning() bool {
	return ar != nil && ar.value == "WARNING"
}

// DenialReason explains why authorization was denied
type DenialReason struct {
	value string
}

var (
	ReasonPolicyNotFound    = &DenialReason{value: "POLICY_NOT_FOUND"}
	ReasonRoleNotFound      = &DenialReason{value: "ROLE_NOT_FOUND"}
	ReasonMethodNotAllowed  = &DenialReason{value: "METHOD_NOT_ALLOWED"}
	ReasonResourceNotFound  = &DenialReason{value: "RESOURCE_NOT_FOUND"}
	ReasonRateLimitExceeded = &DenialReason{value: "RATE_LIMIT_EXCEEDED"}
	ReasonDeprecatedRoute   = &DenialReason{value: "DEPRECATED_ROUTE"}
	ReasonUnknown           = &DenialReason{value: "UNKNOWN"}
)

var validDenialReasons = map[string]bool{
	"POLICY_NOT_FOUND":    true,
	"ROLE_NOT_FOUND":      true,
	"METHOD_NOT_ALLOWED":  true,
	"RESOURCE_NOT_FOUND":  true,
	"RATE_LIMIT_EXCEEDED": true,
	"DEPRECATED_ROUTE":    true,
	"UNKNOWN":             true,
}

func NewDenialReason(reason string) (*DenialReason, error) {
	if reason == "" {
		return nil, fmt.Errorf("denial reason cannot be empty")
	}
	if !validDenialReasons[reason] {
		return nil, fmt.Errorf("invalid denial reason: %s", reason)
	}
	return &DenialReason{value: reason}, nil
}

func (dr *DenialReason) Value() string {
	if dr == nil {
		return ""
	}
	return dr.value
}

func (dr *DenialReason) String() string {
	return dr.Value()
}

// AuthorizationAuditLog tracks authorization check events
type AuthorizationAuditLog struct {
	id              string
	timestamp       time.Time
	userID          string
	role            string
	resource        string
	action          string // HTTP method (GET, POST, etc)
	result          *AuthorizationResult
	denialReason    *DenialReason // Only populated if denied
	ipAddress       string
	userAgent       string
	apiVersion      string                 // e.g., "v1", "v2"
	deprecated      bool                   // true if route is deprecated
	environment     string                 // "dev", "staging", "production"
	rateLimitStatus string                 // "OK", "WARNING", "EXCEEDED"
	policyVersion   int                    // Which version of policy was used
	executionTimeMs float64                // Time taken to check permission
	details         map[string]interface{} // Additional metadata
}

// NewAuthorizationAuditLog creates a new authorization audit log entry
func NewAuthorizationAuditLog(
	id string,
	timestamp time.Time,
	userID string,
	role string,
	resource string,
	action string,
	result *AuthorizationResult,
	denialReason *DenialReason,
	ipAddress string,
	userAgent string,
	apiVersion string,
	deprecated bool,
	environment string,
	rateLimitStatus string,
	policyVersion int,
	executionTimeMs float64,
	details map[string]interface{},
) (*AuthorizationAuditLog, error) {
	// Business rule validation
	if id == "" {
		return nil, fmt.Errorf("authorization audit log ID cannot be empty")
	}
	if timestamp.IsZero() {
		return nil, fmt.Errorf("timestamp cannot be zero")
	}
	// Note: userID and role can be empty for failed authentication attempts
	if resource == "" {
		return nil, fmt.Errorf("resource cannot be empty")
	}
	if action == "" {
		return nil, fmt.Errorf("action cannot be empty")
	}
	if result == nil || result.Value() == "" {
		return nil, fmt.Errorf("authorization result is required")
	}
	// If denied, reason must be provided
	if result.IsDenied() && (denialReason == nil || denialReason.Value() == "") {
		return nil, fmt.Errorf("denial reason is required when authorization is denied")
	}
	if environment == "" {
		return nil, fmt.Errorf("environment cannot be empty")
	}
	if policyVersion < 0 {
		return nil, fmt.Errorf("policy version cannot be negative")
	}
	if executionTimeMs < 0 {
		return nil, fmt.Errorf("execution time cannot be negative")
	}

	if details == nil {
		details = make(map[string]interface{})
	}

	return &AuthorizationAuditLog{
		id:              id,
		timestamp:       timestamp,
		userID:          userID,
		role:            role,
		resource:        resource,
		action:          action,
		result:          result,
		denialReason:    denialReason,
		ipAddress:       ipAddress,
		userAgent:       userAgent,
		apiVersion:      apiVersion,
		deprecated:      deprecated,
		environment:     environment,
		rateLimitStatus: rateLimitStatus,
		policyVersion:   policyVersion,
		executionTimeMs: executionTimeMs,
		details:         details,
	}, nil
}

// Getters
func (aal *AuthorizationAuditLog) ID() string {
	return aal.id
}

func (aal *AuthorizationAuditLog) Timestamp() time.Time {
	return aal.timestamp
}

func (aal *AuthorizationAuditLog) UserID() string {
	return aal.userID
}

func (aal *AuthorizationAuditLog) Role() string {
	return aal.role
}

func (aal *AuthorizationAuditLog) Resource() string {
	return aal.resource
}

func (aal *AuthorizationAuditLog) Action() string {
	return aal.action
}

func (aal *AuthorizationAuditLog) Result() *AuthorizationResult {
	return aal.result
}

func (aal *AuthorizationAuditLog) DenialReason() *DenialReason {
	return aal.denialReason
}

func (aal *AuthorizationAuditLog) IPAddress() string {
	return aal.ipAddress
}

func (aal *AuthorizationAuditLog) UserAgent() string {
	return aal.userAgent
}

func (aal *AuthorizationAuditLog) APIVersion() string {
	return aal.apiVersion
}

func (aal *AuthorizationAuditLog) Deprecated() bool {
	return aal.deprecated
}

func (aal *AuthorizationAuditLog) Environment() string {
	return aal.environment
}

func (aal *AuthorizationAuditLog) RateLimitStatus() string {
	return aal.rateLimitStatus
}

func (aal *AuthorizationAuditLog) PolicyVersion() int {
	return aal.policyVersion
}

func (aal *AuthorizationAuditLog) ExecutionTimeMs() float64 {
	return aal.executionTimeMs
}

func (aal *AuthorizationAuditLog) Details() map[string]interface{} {
	if aal.details == nil {
		return make(map[string]interface{})
	}
	// Return a copy
	details := make(map[string]interface{})
	for k, v := range aal.details {
		details[k] = v
	}
	return details
}

// IsCritical returns true if this is a critical event
func (aal *AuthorizationAuditLog) IsCritical() bool {
	// Critical if denied or rate limited or deprecated
	return aal.result.IsDenied() || aal.rateLimitStatus == "EXCEEDED" || aal.deprecated
}

// IsRecent checks if the log entry is within the specified duration
func (aal *AuthorizationAuditLog) IsRecent(duration time.Duration) bool {
	return time.Since(aal.timestamp) <= duration
}
