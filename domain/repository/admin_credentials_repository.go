package repository

import "github.com/aruncs31s/azf/domain/model"

// AdminCredentialsRepository defines the interface for admin credentials persistence operations
type AdminCredentialsRepository interface {
	AdminCredentialsReader
	AdminCredentialsWriter
}

// AdminCredentialsReader defines the interface for admin credentials read operations
type AdminCredentialsReader interface {
	// FindByID retrieves admin credentials by ID
	//
	// Params:
	//   - id: string The ID of the admin credentials
	//
	// Returns:
	//   - *model.AdminCredentials The admin credentials if found
	//   - error Any error that occurred during the operation
	FindByID(id string) (*model.AdminCredentials, error)

	// FindByUsername retrieves admin credentials by username
	//
	// Params:
	//   - username: string The username to search for
	//
	// Returns:
	//   - *model.AdminCredentials The admin credentials if found
	//   - error Any error that occurred during the operation
	FindByUsername(username string) (*model.AdminCredentials, error)

	// FindAll retrieves all admin credentials
	//
	// Returns:
	//   - *[]model.AdminCredentials All admin credentials
	//   - error Any error that occurred during the operation
	FindAll() (*[]model.AdminCredentials, error)
}

// AdminCredentialsWriter defines the interface for admin credentials write operations
type AdminCredentialsWriter interface {
	// Create adds new admin credentials to the repository
	//
	// Params:
	//   - credentials: *model.AdminCredentials The admin credentials to create
	//
	// Returns:
	//   - *model.AdminCredentials The created credentials
	//   - error Any error that occurred during the operation
	Create(credentials *model.AdminCredentials) (*model.AdminCredentials, error)

	// Update modifies existing admin credentials
	//
	// Params:
	//   - credentials: *model.AdminCredentials The credentials to update
	//
	// Returns:
	//   - *model.AdminCredentials The updated credentials
	//   - error Any error that occurred during the operation
	Update(credentials *model.AdminCredentials) (*model.AdminCredentials, error)

	// Delete removes admin credentials by ID
	//
	// Params:
	//   - id: string The ID of the credentials to delete
	//
	// Returns:
	//   - error Any error that occurred during the operation
	Delete(id string) error
}
