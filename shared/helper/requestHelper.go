package helper

/*

Request helper implementation v0.0.1
Date: 2025-10-22


*/

import (
	"strconv"

	"github.com/aruncs31s/azf/shared/interface/helper"
	"github.com/aruncs31s/azf/utils"
	"github.com/gin-gonic/gin"
)

type requestHelper struct{}

func NewRequestHelper() helper.RequestHelper {
	return &requestHelper{}
}

func (r *requestHelper) GetAndValidateUserID(c *gin.Context, validator helper.RequestValidator, responseHelper helper.ResponseHelper) (string, error) {
	userID := c.GetString("user_id")
	err := r.validateUserID(validator, responseHelper, userID, c)
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (r *requestHelper) GetAndValidateStaff(c *gin.Context, validator helper.RequestValidator, responseHelper helper.ResponseHelper) (string, error) {
	staffID := c.GetString("staff_id")
	err := r.validateStaffID(validator, responseHelper, staffID, c)
	if err != nil {
		return "", err
	}
	return staffID, nil
}

// GetLimitAndOffset extracts pagination parameters from the request context.
// Matched to client side request.
//
// Params: c *gin.Context - The Gin context containing the request.
//
// Returns: (int, int) - The limit and offset values for pagination.
func (r *requestHelper) GetLimitAndOffset(c *gin.Context) (int, int) {
	limit, _ := strconv.Atoi(c.DefaultQuery("per-page", "10"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	// calculate offset form page and limit
	offset := (page - 1) * limit
	return limit, offset
}

func (r *requestHelper) validateUserID(validator helper.RequestValidator, responseHelper helper.ResponseHelper, staffID string, c *gin.Context) error {
	if err := validator.ValidateUserID(staffID); err != nil {
		responseHelper.BadRequest(c, err.Error(), utils.FixInvalidUserID)
		return utils.ErrInvalidUserID
	}
	return nil
}

func (r *requestHelper) validateStaffID(validator helper.RequestValidator, responseHelper helper.ResponseHelper, staffID string, c *gin.Context) error {
	if err := validator.ValidateUserID(staffID); err != nil {
		responseHelper.BadRequest(c, err.Error(), utils.FixInvalidStaffID)
		return utils.ErrInvalidStaffID
	}
	return nil
}

func (r *requestHelper) GetURLParam(c *gin.Context, paramName string) string {
	return c.Param(paramName)
}

// Generic ID validation helper

// WARNING: Not Optimized.
func (r *requestHelper) ValidateAndParseID(c *gin.Context, idName string) (int, error) {
	validator := NewRequestValidator()
	idStr := c.Param(idName)
	id, err := validator.ValidateID(idStr)
	if err != nil {
		return 0, utils.ErrInvalidID
	}
	return id, nil
}
func (r *requestHelper) ValidateAndParseCode(c *gin.Context, codeName string) (string, error) {
	validator := NewRequestValidator()
	codeStr := c.Param(codeName)
	code, err := validator.ValidateCode(codeStr)
	if err != nil {
		return "", utils.ErrInvalidCode
	}
	return code, nil
}
func (r *requestHelper) GetAgentAndIPFromContext(c *gin.Context) (string, string) {
	agent := c.Request.UserAgent()
	ip := c.ClientIP()
	return agent, ip
}
