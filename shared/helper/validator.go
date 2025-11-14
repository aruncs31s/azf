package helper

import (
	"strconv"

	"github.com/aruncs31s/azf/shared/interface/helper"
	"github.com/aruncs31s/azf/utils"
)

// Request validation helpers - easily testable
type requestValidator struct{}

func NewRequestValidator() helper.RequestValidator {
	return &requestValidator{}
}

func (v *requestValidator) ValidateUserID(staffID string) error {
	if staffID == "" {
		return utils.ErrNoStaffIDInJWT
	}
	return nil
}
func (v *requestValidator) ValidateID(IDParam string) (int, error) {
	if IDParam == "" {
		return 0, utils.ErrNoIDProvided
	}
	ID, err := strconv.Atoi(IDParam)
	if err != nil {
		return 0, utils.ErrInvalidIDFormat
	}
	return int(ID), nil
}
func (v *requestValidator) ValidateCode(codeParam string) (string, error) {
	if codeParam == "" {
		return "", utils.ErrNoCodeProvided
	}
	return codeParam, nil
}
