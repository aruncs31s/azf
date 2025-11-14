package analytics

import (
	"fmt"
	"time"
)

// HTTPMethod is a value object representing HTTP methods
type HTTPMethod struct {
	value string
}

var (
	MethodGET    = &HTTPMethod{value: "GET"}
	MethodPOST   = &HTTPMethod{value: "POST"}
	MethodPUT    = &HTTPMethod{value: "PUT"}
	MethodPATCH  = &HTTPMethod{value: "PATCH"}
	MethodDELETE = &HTTPMethod{value: "DELETE"}
)

var validMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
}

// NewHTTPMethod creates a new HTTPMethod with validation
func NewHTTPMethod(method string) (*HTTPMethod, error) {
	if !validMethods[method] {
		return nil, fmt.Errorf("invalid HTTP method: %s", method)
	}
	return &HTTPMethod{value: method}, nil
}

// Value returns the string value
func (m *HTTPMethod) Value() string {
	return m.value
}

// APIUsageLog is a domain entity representing a single API call record
type APIUsageLog struct {
	id             string
	endpoint       string
	method         *HTTPMethod
	statusCode     int
	responseTimeMs int64
	requestSize    int
	responseSize   int
	userID         string
	clientIP       string
	userAgent      string
	errorMessage   string
	requestedAt    time.Time
	lastAccessedAt time.Time
}

// NewAPIUsageLog creates a new APIUsageLog with validation
func NewAPIUsageLog(
	id string,
	endpoint string,
	method *HTTPMethod,
	statusCode int,
	responseTimeMs int64,
	requestSize int,
	responseSize int,
	userID string,
	clientIP string,
	userAgent string,
	errorMessage string,
	requestedAt time.Time,
) (*APIUsageLog, error) {
	if id == "" {
		return nil, fmt.Errorf("API usage log ID cannot be empty")
	}
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}
	if method == nil {
		return nil, fmt.Errorf("HTTP method is required")
	}
	if statusCode < 100 || statusCode > 599 {
		return nil, fmt.Errorf("invalid HTTP status code: %d", statusCode)
	}
	if responseTimeMs < 0 {
		return nil, fmt.Errorf("response time cannot be negative")
	}
	if requestedAt.IsZero() {
		return nil, fmt.Errorf("requested time cannot be empty")
	}

	return &APIUsageLog{
		id:             id,
		endpoint:       endpoint,
		method:         method,
		statusCode:     statusCode,
		responseTimeMs: responseTimeMs,
		requestSize:    requestSize,
		responseSize:   responseSize,
		userID:         userID,
		clientIP:       clientIP,
		userAgent:      userAgent,
		errorMessage:   errorMessage,
		requestedAt:    requestedAt,
		lastAccessedAt: requestedAt,
	}, nil
}

func (aul *APIUsageLog) ID() string {
	return aul.id
}

func (aul *APIUsageLog) Endpoint() string {
	return aul.endpoint
}

func (aul *APIUsageLog) Method() *HTTPMethod {
	return aul.method
}

func (aul *APIUsageLog) StatusCode() int {
	return aul.statusCode
}

func (aul *APIUsageLog) ResponseTimeMs() int64 {
	return aul.responseTimeMs
}

func (aul *APIUsageLog) RequestSize() int {
	return aul.requestSize
}

func (aul *APIUsageLog) ResponseSize() int {
	return aul.responseSize
}

func (aul *APIUsageLog) UserID() string {
	return aul.userID
}

func (aul *APIUsageLog) ClientIP() string {
	return aul.clientIP
}

func (aul *APIUsageLog) UserAgent() string {
	return aul.userAgent
}

func (aul *APIUsageLog) ErrorMessage() string {
	return aul.errorMessage
}

func (aul *APIUsageLog) RequestedAt() time.Time {
	return aul.requestedAt
}

func (aul *APIUsageLog) LastAccessedAt() time.Time {
	return aul.lastAccessedAt
}

// IsError checks if the response indicates an error
func (aul *APIUsageLog) IsError() bool {
	return aul.statusCode >= 400
}

// IsSlowResponse checks if response time exceeds threshold (ms)
func (aul *APIUsageLog) IsSlowResponse(thresholdMs int64) bool {
	return aul.responseTimeMs > thresholdMs
}
