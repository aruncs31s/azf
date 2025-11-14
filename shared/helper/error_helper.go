package helper

import (
	"fmt"

	"github.com/aruncs31s/azf/shared/interface/helper"
)

type errorHelper struct {
}

func (errorHelper) GetRecordDoesNotBelongErrorMessage(id any, user string) error {
	return fmt.Errorf("the record %d does not belong to  %v", id, user)
}
func NewErrorHelper() helper.ErrorHelper {
	return &errorHelper{}
}
