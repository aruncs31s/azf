package utils

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// When the staff is not found for the provided staff_id (ref_id)
var (
	ErrNoPermission              = errors.New("You do not have permission to access this resource")
	ErrNoUserRoleInJWT           = errors.New("User role not found. Please authenticate first.")
	ErrInvalidUsernameOrPassword = errors.New("Invalid username or password")
	ErrStaffNotFound             = errors.New("staff with this id not found")
	ErrStaffIsNull               = errors.New("staff is nil")
	ErrNoIDProvided              = errors.New("no id provided")
	ErrInvalidIDFormat           = errors.New("invalid id format")
	ErrNoCodeProvided            = errors.New("no code provided")
	ErrNoStaffIDProvided         = errors.New("no staff_id provided")
	ErrInvalidStaffID            = errors.New("invalid staff id")
	ErrStaffAlreadyExists        = errors.New("staff already exists")
	ErrInvalidUserID             = errors.New("invalid user id")
	ErrInvalidDegreeID           = errors.New("invalid degree id")
	ErrInvalidID                 = errors.New("invalid id")
	ErrInvalidCode               = errors.New("invalid code")
	ErrDatabase                  = errors.New("database error")
	ErrRecordsNotFound           = errors.New("no records found")
	ErrNoQualificationRecords    = errors.New("no qualification records found")
	ErrRecordNotFound            = errors.New("record not found")
	ErrInvalidQualificationID    = errors.New("invalid qualification id")
	ErrUnauthorized              = errors.New("unauthorized access/forbidden")
	ErrNoAuthHeader              = errors.New("no authorization header")
	// When the request data is invalid or malformed
	//
	// - e.g., missing required fields, incorrect data types, etc.
	// - Mainly used during JSON binding and validation
	//
	// Use this error to indicate that the client has sent a request that the server cannot process due to client-side issues.
	ErrBadRequest                    = errors.New("invalid request data")
	ErrRecordDoesNotBelongToStaff    = errors.New("record does not belong to the staff")
	ErrExperienceNotBelongToStaff    = errors.New("experience does not belong to the staff")
	ErrQualificationNotBelongToStaff = errors.New("qualification does not belong to the staff")
	ErrInvalidExperienceID           = errors.New("invalid experience id")
	ErrInternalServerError           = errors.New("internal server error")
)
var (
	ErrMappingStaffDetails           = errors.New("staff detail mapping failed: foreign key mismatch or invalid data")
	ErrMappingStaffAdditionalDetails = errors.New("staff additional detail mapping failed: foreign key mismatch or invalid data")
	ErrMappingStaffPayScaleDetails   = errors.New("staff pay scale detail mapping failed: foreign key mismatch or invalid data")
	ErrInvalidExperienceType         = errors.New("invalid experience type")
)
var (
	ErrInvalidUsername = errors.New("username cannot be empty")
	ErrUsernameTooLong = errors.New("username cannot exceed 100 characters")
	ErrInvalidPassword = errors.New("password cannot be empty")
)
var (
	ErrInvalidData = errors.New("invalid data provided")
)

var (
	FixInvalidFormData       = "provide valid form data"
	FixNoUsernameAndPassword = "username and password cannot be empty"
)

var (
	FixInvalidQualificationID = "provide a valid qualification id"
	FixInvalidExperienceID    = "provide a valid experience id"
	FixInvalidProjectID       = "provide a valid project id"
	FixInvalidExperienceData  = "provide valid experience data"
	FixInvalidStaffID         = "staff id is invalid"
	FixInvalidUserID          = "user id is invalid"
	FixInvalidRequestData     = "provide valid request data"
)

var (
	ErrNoStaffIDInJWT = errors.New("no staff_id in jwt token")
	FixNoStaffIDInJWT = "provide a valid staff_id in jwt token"
)

var (
	ErrAccessDenied            = errors.New("access denied")
	ErrPermissionDenied        = errors.New("permission denied")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
)

var (
	ErrReligionAlreadyExists = errors.New("religion with this name already exists")
)

func GetDoesNotBelongErr(resource string, staffID string) error {
	return fmt.Errorf("%s does not belong to the staff %s", resource, staffID)
}
func IsNotFound(err error) bool {
	// Checking for Sentinel Errors.
	if errors.Is(err, ErrRecordsNotFound) || errors.Is(err, ErrStaffNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	}
	// Check for "record not found" substring in the error message.
	// Temporary, and will be removed once all the code is migrated to use sentinel errors.
	return strings.Contains(err.Error(), "record not found") || strings.Contains(err.Error(), "not found")
}
func IsInternalErr(err error) bool {
	return errors.Is(err, ErrInternalServerError)
}
func IsDoesNotBelongErr(err error) bool {
	return strings.Contains(err.Error(), "belong to")
}
func GetNotFoundErrorMessage(product string) error {
	return fmt.Errorf("%s not found", product)
}
func IsBadRequest(err error) bool {
	return errors.Is(err, ErrBadRequest) || errors.Is(err, ErrInvalidStaffID) || errors.Is(err, ErrInvalidID)
}
func IsAlreadyExists(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "already exists") || strings.Contains(errMsg, "duplicate")
}
func GetFailedToFetchErr(recordName string) error {
	return getfailedTotext("fetch", recordName)
}
func GetFailedToCreateErr(recordName string) error {
	return getfailedTotext("create", recordName)
}
func GetFailedToUpdateErr(resourceName string) error {
	return getfailedTotext("update", resourceName)
}
func GetDeletedSuccessfullyMessage(recordName string) string {
	return fmt.Sprintf("%s deleted successfully", recordName)
}
func getfailedTotext(operation string, recordName string) error {
	return fmt.Errorf("failed to %s %s", operation, recordName)
}
func GetInvalidResourceDataErr(resourceName string) error {
	return fmt.Errorf("invalid %s data provided", resourceName)
}
func GetRecordAlreadyExistsErr(recordName string) error {
	return fmt.Errorf("%s already exists", recordName)
}
func GetEnvVarNotSetErr(envVarName string) error {
	return fmt.Errorf("environment variable %s not set", envVarName)
}
func GetEnvVarNotValidErr(envVarName string) error {
	return fmt.Errorf("environment variable %s not valid", envVarName)
}
func GetResourceCanotBeNilErr(resource string) error {
	return fmt.Errorf("%v can not be nil", resource)
}
