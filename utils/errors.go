package utils

import (
	"errors"
)

var (
	ErrNoEnvVar = errors.New("no env variable with this name")
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
