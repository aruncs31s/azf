package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Standard error codes
const (
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeNotFound           = "NOT_FOUND"
	CodeBadRequest         = "BAD_REQUEST"
	CodeInternalError      = "INTERNAL_ERROR"
	CodeValidation         = "VALIDATION_ERROR"
	CodeConflict           = "CONFLICT"
	CodeRateLimited        = "RATE_LIMITED"
	CodeTokenExpired       = "TOKEN_EXPIRED"
	CodeTokenInvalid       = "TOKEN_INVALID"
	CodeInsufficientRole   = "INSUFFICIENT_ROLE"
	CodeResourceNotOwned   = "RESOURCE_NOT_OWNED"
	CodeDatabaseError      = "DATABASE_ERROR"
	CodeConfigurationError = "CONFIGURATION_ERROR"
)

// AppError represents an application error with context
type AppError struct {
	// Code is the error code for programmatic handling
	Code string `json:"code"`

	// Message is the human-readable error message
	Message string `json:"message"`

	// HTTPStatus is the HTTP status code to return
	HTTPStatus int `json:"-"`

	// Op is the operation that failed
	Op string `json:"-"`

	// Err is the underlying error
	Err error `json:"-"`

	// Details contains additional error context
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Op != "" {
		if e.Err != nil {
			return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
		}
		return fmt.Sprintf("%s: %s", e.Op, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches the target
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WithOp adds operation context to the error
func (e *AppError) WithOp(op string) *AppError {
	e.Op = op
	return e
}

// WithDetails adds additional context details
func (e *AppError) WithDetails(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Wrap wraps an underlying error
func (e *AppError) Wrap(err error) *AppError {
	e.Err = err
	return e
}

// Constructor functions for common errors

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       CodeUnauthorized,
		Message:    message,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:       CodeForbidden,
		Message:    message,
		HTTPStatus: http.StatusForbidden,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		HTTPStatus: http.StatusNotFound,
	}
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:       CodeBadRequest,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:       CodeValidation,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewInternalError creates an internal server error
func NewInternalError(message string) *AppError {
	return &AppError{
		Code:       CodeInternalError,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       CodeConflict,
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

// NewRateLimitedError creates a rate limited error
func NewRateLimitedError() *AppError {
	return &AppError{
		Code:       CodeRateLimited,
		Message:    "rate limit exceeded, please try again later",
		HTTPStatus: http.StatusTooManyRequests,
	}
}

// NewTokenExpiredError creates a token expired error
func NewTokenExpiredError() *AppError {
	return &AppError{
		Code:       CodeTokenExpired,
		Message:    "token has expired",
		HTTPStatus: http.StatusUnauthorized,
	}
}

// NewTokenInvalidError creates a token invalid error
func NewTokenInvalidError(reason string) *AppError {
	return &AppError{
		Code:       CodeTokenInvalid,
		Message:    fmt.Sprintf("invalid token: %s", reason),
		HTTPStatus: http.StatusUnauthorized,
	}
}

// NewInsufficientRoleError creates an insufficient role error
func NewInsufficientRoleError(required, actual string) *AppError {
	return &AppError{
		Code:       CodeInsufficientRole,
		Message:    "insufficient permissions for this action",
		HTTPStatus: http.StatusForbidden,
		Details: map[string]interface{}{
			"required_role": required,
			"actual_role":   actual,
		},
	}
}

// NewDatabaseError creates a database error
func NewDatabaseError(op string, err error) *AppError {
	return &AppError{
		Code:       CodeDatabaseError,
		Message:    "database operation failed",
		HTTPStatus: http.StatusInternalServerError,
		Op:         op,
		Err:        err,
	}
}

// NewConfigurationError creates a configuration error
func NewConfigurationError(message string) *AppError {
	return &AppError{
		Code:       CodeConfigurationError,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// Sentinel errors for common cases
var (
	ErrUnauthorized       = NewUnauthorizedError("authentication required")
	ErrForbidden          = NewForbiddenError("access denied")
	ErrTokenExpired       = NewTokenExpiredError()
	ErrInvalidCredentials = NewUnauthorizedError("invalid credentials")
	ErrRateLimited        = NewRateLimitedError()
)

// Helper functions

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from an error chain
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

// GetHTTPStatus returns the HTTP status code for an error
func GetHTTPStatus(err error) int {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// GetErrorCode returns the error code for an error
func GetErrorCode(err error) string {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.Code
	}
	return CodeInternalError
}
