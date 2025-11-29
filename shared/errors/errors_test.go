package errors_test

import (
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/aruncs31s/azf/shared/errors"
)

func TestAppError_Error(t *testing.T) {
	err := apperrors.NewBadRequestError("invalid input")
	if err.Error() != "invalid input" {
		t.Errorf("Expected 'invalid input', got '%s'", err.Error())
	}

	err = err.WithOp("CreateUser")
	if err.Error() != "CreateUser: invalid input" {
		t.Errorf("Expected 'CreateUser: invalid input', got '%s'", err.Error())
	}
}

func TestAppError_Wrap(t *testing.T) {
	innerErr := errors.New("database connection failed")
	err := apperrors.NewDatabaseError("query", innerErr)

	if !errors.Is(err, innerErr) {
		t.Error("Expected error to wrap inner error")
	}
}

func TestAppError_WithDetails(t *testing.T) {
	err := apperrors.NewValidationError("validation failed")
	err = err.WithDetails("field", "email").WithDetails("reason", "invalid format")

	if err.Details["field"] != "email" {
		t.Errorf("Expected field to be 'email', got '%v'", err.Details["field"])
	}
	if err.Details["reason"] != "invalid format" {
		t.Errorf("Expected reason to be 'invalid format', got '%v'", err.Details["reason"])
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	err := apperrors.NewUnauthorizedError("not authenticated")
	if err.Code != apperrors.CodeUnauthorized {
		t.Errorf("Expected code '%s', got '%s'", apperrors.CodeUnauthorized, err.Code)
	}
	if err.HTTPStatus != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, err.HTTPStatus)
	}
}

func TestNewForbiddenError(t *testing.T) {
	err := apperrors.NewForbiddenError("access denied")
	if err.Code != apperrors.CodeForbidden {
		t.Errorf("Expected code '%s', got '%s'", apperrors.CodeForbidden, err.Code)
	}
	if err.HTTPStatus != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, err.HTTPStatus)
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := apperrors.NewNotFoundError("user")
	if err.Code != apperrors.CodeNotFound {
		t.Errorf("Expected code '%s', got '%s'", apperrors.CodeNotFound, err.Code)
	}
	if err.Message != "user not found" {
		t.Errorf("Expected message 'user not found', got '%s'", err.Message)
	}
	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, err.HTTPStatus)
	}
}

func TestNewInsufficientRoleError(t *testing.T) {
	err := apperrors.NewInsufficientRoleError("admin", "user")
	if err.Code != apperrors.CodeInsufficientRole {
		t.Errorf("Expected code '%s', got '%s'", apperrors.CodeInsufficientRole, err.Code)
	}
	if err.Details["required_role"] != "admin" {
		t.Errorf("Expected required_role 'admin', got '%v'", err.Details["required_role"])
	}
	if err.Details["actual_role"] != "user" {
		t.Errorf("Expected actual_role 'user', got '%v'", err.Details["actual_role"])
	}
}

func TestNewRateLimitedError(t *testing.T) {
	err := apperrors.NewRateLimitedError()
	if err.Code != apperrors.CodeRateLimited {
		t.Errorf("Expected code '%s', got '%s'", apperrors.CodeRateLimited, err.Code)
	}
	if err.HTTPStatus != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, err.HTTPStatus)
	}
}

func TestIsAppError(t *testing.T) {
	appErr := apperrors.NewBadRequestError("test")
	if !apperrors.IsAppError(appErr) {
		t.Error("Expected IsAppError to return true for AppError")
	}

	stdErr := errors.New("standard error")
	if apperrors.IsAppError(stdErr) {
		t.Error("Expected IsAppError to return false for standard error")
	}
}

func TestGetAppError(t *testing.T) {
	appErr := apperrors.NewBadRequestError("test")
	result := apperrors.GetAppError(appErr)
	if result == nil {
		t.Error("Expected GetAppError to return AppError")
	}
	if result.Code != apperrors.CodeBadRequest {
		t.Errorf("Expected code '%s', got '%s'", apperrors.CodeBadRequest, result.Code)
	}

	stdErr := errors.New("standard error")
	result = apperrors.GetAppError(stdErr)
	if result != nil {
		t.Error("Expected GetAppError to return nil for standard error")
	}
}

func TestGetHTTPStatus(t *testing.T) {
	appErr := apperrors.NewNotFoundError("user")
	status := apperrors.GetHTTPStatus(appErr)
	if status != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, status)
	}

	stdErr := errors.New("standard error")
	status = apperrors.GetHTTPStatus(stdErr)
	if status != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, status)
	}
}

func TestGetErrorCode(t *testing.T) {
	appErr := apperrors.NewConflictError("already exists")
	code := apperrors.GetErrorCode(appErr)
	if code != apperrors.CodeConflict {
		t.Errorf("Expected code '%s', got '%s'", apperrors.CodeConflict, code)
	}

	stdErr := errors.New("standard error")
	code = apperrors.GetErrorCode(stdErr)
	if code != apperrors.CodeInternalError {
		t.Errorf("Expected code '%s', got '%s'", apperrors.CodeInternalError, code)
	}
}

func TestSentinelErrors(t *testing.T) {
	if apperrors.ErrUnauthorized.Code != apperrors.CodeUnauthorized {
		t.Error("ErrUnauthorized has wrong code")
	}
	if apperrors.ErrForbidden.Code != apperrors.CodeForbidden {
		t.Error("ErrForbidden has wrong code")
	}
	if apperrors.ErrTokenExpired.Code != apperrors.CodeTokenExpired {
		t.Error("ErrTokenExpired has wrong code")
	}
	if apperrors.ErrRateLimited.Code != apperrors.CodeRateLimited {
		t.Error("ErrRateLimited has wrong code")
	}
}
