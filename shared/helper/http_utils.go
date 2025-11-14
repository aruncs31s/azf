package helper

import (
	"strconv"

	"github.com/aruncs31s/azf/utils/customerrors"
	"github.com/gin-gonic/gin"
)

// Generic function to extract JSON data from the request body.
// It binds the JSON data to a struct of type T and handles errors appropriately.
//
// Params:
//   - c *gin.Context: The Gin context containing the request.
//   - responseHelper responsehelper.ResponseHelper: Helper for sending responses.
//
// Returns:
//   - *T: The reference to a  struct populated with JSON data.
//   - error: An error if binding fails, otherwise nil.
func GetJSONDataFromRequest[T any](c *gin.Context) (*T, error) {
	var data T
	if err := c.ShouldBindJSON(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

// NOTE: Currently moving to requestHelper struct method , so remove this in the future.
// GetLimitAndOffset extracts pagination parameters from the request context.
// Matched to client side request.
//
// Params: c *gin.Context - The Gin context containing the request.
//
// Returns: (int, int) - The limit and offset values for pagination.
func GetLimitAndOffset(c *gin.Context) (int, int) {
	limit, _ := strconv.Atoi(c.DefaultQuery("per-page", "10"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	// calculate offset form page and limit
	offset := (page - 1) * limit
	return limit, offset
}

func GetMaxFileSize() int64 {
	// 5 MB
	KB := 1024
	MB := 1024 * KB
	return 5 * int64(MB)
}

func CheckIfValidFileType(fileType string, allowedTypes []string) error {
	for _, t := range allowedTypes {
		if t == fileType {
			return nil
		}
	}
	return customerrors.ErrInvalidFileType
}

func GetAgentAndIPFromContext(c *gin.Context) (string, string) {
	agent := c.Request.UserAgent()
	ip := c.ClientIP()
	return agent, ip
}
