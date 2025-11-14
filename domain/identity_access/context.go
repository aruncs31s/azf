package identity_access

// BoundedContext represents the Identity & Access Management bounded context
// This context manages all aspects of admin authentication and credential management
// including user identification, password management, and session handling.
//
// Responsibilities:
// - Admin credential creation and validation
// - Username and password management
// - Admin profile information
// - Authentication workflows
//
// Does NOT handle:
// - Authorization decisions (Authorization & Audit context)
// - Audit logging of auth events (Authorization & Audit context)
// - API usage tracking (Analytics context)
type BoundedContext struct{}
