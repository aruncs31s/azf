package helper

import (
	"github.com/gin-gonic/gin"
)

type RequestHelper interface {
	// GetAndValidateUserID extracts the user ID from the request context.
	// Note: This also sends bad request response if validation fails.
	//
	// Parameters:
	//   - c: Gin context containing the request
	//   - validator: Interface providing access to validator
	//   - responseHelper: Interface providing access to response helper
	//
	// Returns:
	//   - string: Extracted user ID
	//   - error: Error if extraction failed
	GetAndValidateUserID(c *gin.Context, validator RequestValidator, responseHelper ResponseHelper) (string, error)
	// GetStaffID extracts the staff ID from the request context.
	// Note: This also sends bad request response if validation fails.
	//
	// Parameters:
	//   - c: Gin context containing the request
	//   - h: Interface providing access to validator and response helper
	//
	// Returns:
	//   - string: Extracted staff ID
	//   - bool: True if extraction failed, false if successful
	// GetAndValidateUserID(c *gin.Context, validator RequestValidator, responseHelper ResponseHelper) (string, error)
	// ValidateAndParseID validates and parses a string ID into a int
	// Parameters:
	//   - h: Interface providing access to validator and response helper
	//   - idStr: String ID to validate and parse
	//   - c: Gin context for handling the response
	//   - fixMessage: Message to display if validation fails
	// Returns:
	//   - int: Parsed ID value
	//   - bool: True if validation failed, false if successful
	ValidateAndParseID(c *gin.Context, idName string) (int, error)
	// ValidateAndParseCode validates and parses a string code
	//
	// Params:
	//  - c *gin.Context: The Gin context containing the request.
	// - codeName string: The name of the code parameter to validate and parse.
	// Returns:
	// - string: The validated code.
	// - error: An error if validation fails, nil otherwise.
	ValidateAndParseCode(c *gin.Context, codeName string) (string, error)

	// GetLimitAndOffset extracts pagination parameters from the request context.
	// Matched to client side request.
	//
	// Params: c *gin.Context - The Gin context containing the request.
	//
	// Returns: (int, int) - The limit and offset values for pagination.
	GetLimitAndOffset(c *gin.Context) (int, int)
	// GetURLParam extracts a URL parameter from the request context.
	//
	// Params:
	//   - c *gin.Context: The Gin context containing the request.
	//   - paramName string: The name of the URL parameter to extract.
	//
	// Returns:
	//   - string: The value of the specified URL parameter.
	GetAgentAndIPFromContext(c *gin.Context) (string, string)
}

// type HasValidator interface {
// 	GetValidator() RequestValidator
// }

// type HasResponseHelper interface {
// 	GetResponseHelper() ResponseHelper
// }
// type HasBothValidatorAndResponseHelper interface {
// 	HasValidator
// 	HasResponseHelper
// }
