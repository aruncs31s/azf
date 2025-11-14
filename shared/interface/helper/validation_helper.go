package helper

type RequestValidator interface {
	// ValidateStaffID checks if the provided staff ID is valid by ensuring it is not empty.
	//
	// Params:
	//  - staffID: string representing the staff ID to validate.
	// Returns:
	//  - error: returns an error if validation fails, nil otherwise.
	ValidateUserID(staffID string) error

	// ValidateID converts the provided qualification ID string to a int after validating it.
	//
	// Params:
	//  - qualificationIDParam: string representing the qualification ID to validate and convert.
	// Returns:
	//  - int: the converted qualification ID.
	//  - error: returns an error if validation or conversion fails, nil otherwise.
	ValidateID(qualificationIDParam string) (int, error)
	// ValidateCode validates the provided code string.
	//
	// Params:
	//  - codeParam: string representing the code to validate.
	// Returns:
	//  - string: the validated code.
	//  - error: returns an error if validation fails, nil otherwise.
	ValidateCode(codeParam string) (string, error)
}
