package response

import (
	"net/http"

	apperrors "github.com/aruncs31s/azf/shared/errors"
	"github.com/gin-gonic/gin"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// APIError represents an error in the API response
type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Meta contains pagination and other metadata
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// Success sends a successful response with data
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success:   true,
		Data:      data,
		RequestID: getRequestID(c),
	})
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// SuccessWithMeta sends a successful response with data and metadata
func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, APIResponse{
		Success:   true,
		Data:      data,
		Meta:      meta,
		RequestID: getRequestID(c),
	})
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, data interface{}, page, perPage int, total int64) {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
		RequestID: getRequestID(c),
	})
}

// Error sends an error response
func Error(c *gin.Context, err error) {
	appErr := apperrors.GetAppError(err)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			},
			RequestID: getRequestID(c),
		})
		return
	}

	// Default to internal server error for unknown errors
	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    apperrors.CodeInternalError,
			Message: "an unexpected error occurred",
		},
		RequestID: getRequestID(c),
	})
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    apperrors.CodeBadRequest,
			Message: message,
		},
		RequestID: getRequestID(c),
	})
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    apperrors.CodeUnauthorized,
			Message: message,
		},
		RequestID: getRequestID(c),
	})
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    apperrors.CodeForbidden,
			Message: message,
		},
		RequestID: getRequestID(c),
	})
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    apperrors.CodeNotFound,
			Message: resource + " not found",
		},
		RequestID: getRequestID(c),
	})
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    apperrors.CodeConflict,
			Message: message,
		},
		RequestID: getRequestID(c),
	})
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, message string, details map[string]interface{}) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    apperrors.CodeValidation,
			Message: message,
			Details: details,
		},
		RequestID: getRequestID(c),
	})
}

// RateLimited sends a 429 Too Many Requests response
func RateLimited(c *gin.Context) {
	c.JSON(http.StatusTooManyRequests, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    apperrors.CodeRateLimited,
			Message: "rate limit exceeded, please try again later",
		},
		RequestID: getRequestID(c),
	})
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    apperrors.CodeInternalError,
			Message: message,
		},
		RequestID: getRequestID(c),
	})
}

// getRequestID extracts the request ID from the gin context
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
