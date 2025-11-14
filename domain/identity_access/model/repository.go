package identity_access

import "context"

// AdminCredentialsRepository defines the interface for admin credentials persistence operations
type AdminCredentialsRepository interface {
AdminCredentialsReader
AdminCredentialsWriter
}

// AdminCredentialsReader defines the interface for admin credentials read operations
type AdminCredentialsReader interface {
// FindByID retrieves admin credentials by ID
FindByID(ctx context.Context, id string) (*AdminCredentials, error)

// FindByUsername retrieves admin credentials by username
FindByUsername(ctx context.Context, username string) (*AdminCredentials, error)

// FindAll retrieves all admin credentials
FindAll(ctx context.Context) (*[]AdminCredentials, error)
}

// AdminCredentialsWriter defines the interface for admin credentials write operations
type AdminCredentialsWriter interface {
// Create adds new admin credentials to the repository
Create(ctx context.Context, credentials *AdminCredentials) (*AdminCredentials, error)

// Update modifies existing admin credentials
Update(ctx context.Context, credentials *AdminCredentials) (*AdminCredentials, error)

// Delete removes admin credentials from the repository
Delete(ctx context.Context, id string) error
}
