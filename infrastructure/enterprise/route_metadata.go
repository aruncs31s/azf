package enterprise

import (
	"fmt"
	"strings"
)

// RouteMetadata contains metadata for a route that can be used to auto-generate policies
type RouteMetadata struct {
	Path             string           `json:"path"`          // /api/v1/staff/profile
	Method           string           `json:"method"`        // GET, POST, PUT, DELETE
	AllowedRoles     []string         `json:"allowed_roles"` // ["staff", "admin"]
	Description      string           `json:"description"`   // For Swagger docs
	Deprecated       bool             `json:"deprecated"`    // true if route is deprecated
	DeprecatedReason string           `json:"deprecated_reason"`
	ReplacedBy       string           `json:"replaced_by"`     // New route to use instead
	APIVersion       string           `json:"api_version"`     // "v1", "v2", etc
	RateLimit        *RateLimitConfig `json:"rate_limit"`      // Rate limit per role
	RequiredScopes   []string         `json:"required_scopes"` // OAuth scopes
	IsPublic         bool             `json:"is_public"`       // true if route doesn't require auth
	OwnershipCheck   bool             `json:"ownership_check"` // true if record ownership should be validated
	AuditRequired    bool             `json:"audit_required"`  // true if action should be logged
	Tags             []string         `json:"tags"`            // Grouping tags
}

// }

// Validate checks if the metadata is valid
func (rm *RouteMetadata) Validate() error {
	if rm.Path == "" {
		return fmt.Errorf("route path cannot be empty")
	}
	if !strings.HasPrefix(rm.Path, "/") {
		return fmt.Errorf("route path must start with /: %s", rm.Path)
	}
	if rm.Method == "" {
		return fmt.Errorf("route method cannot be empty for path %s", rm.Path)
	}

	validMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "DELETE": true, "PATCH": true, "OPTIONS": true, "HEAD": true}
	if !validMethods[strings.ToUpper(rm.Method)] {
		return fmt.Errorf("invalid HTTP method %s for path %s", rm.Method, rm.Path)
	}

	if !rm.IsPublic && len(rm.AllowedRoles) == 0 {
		return fmt.Errorf("protected route %s %s must have at least one allowed role", rm.Method, rm.Path)
	}

	if rm.Deprecated && rm.ReplacedBy == "" {
		return fmt.Errorf("deprecated route %s %s should specify a replacement", rm.Method, rm.Path)
	}

	if rm.APIVersion == "" {
		return fmt.Errorf("api_version is required for route %s %s", rm.Method, rm.Path)
	}

	return nil
}

// MatchesPath checks if this metadata matches the given path and method
func (rm *RouteMetadata) MatchesPath(path, method string) bool {
	return rm.Path == path && strings.EqualFold(rm.Method, method)
}

// GetRateLimit returns the rate limit for a specific role
func (rm *RouteMetadata) GetRateLimit(role string) int {
	if rm.RateLimit == nil {
		return 0
	}
	if limit, exists := rm.RateLimit.RoleSpecificLimits[role]; exists {
		return limit
	}
	return rm.RateLimit.DefaultRequestsPerMinute
}

// IsDeprecated returns true if the route is deprecated
func (rm *RouteMetadata) IsDeprecated() bool {
	return rm.Deprecated
}

// GetDeprecationMessage returns a formatted deprecation message
func (rm *RouteMetadata) GetDeprecationMessage() string {
	msg := fmt.Sprintf("Route %s %s is deprecated", rm.Method, rm.Path)
	if rm.DeprecatedReason != "" {
		msg += fmt.Sprintf(": %s", rm.DeprecatedReason)
	}
	if rm.ReplacedBy != "" {
		msg += fmt.Sprintf(". Use %s instead", rm.ReplacedBy)
	}
	return msg
}

// RouteRegistry maintains a registry of all routes with their metadata
type RouteRegistry struct {
	routes map[string]*RouteMetadata // key: "METHOD:PATH"
}

// NewRouteRegistry creates a new route registry
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		routes: make(map[string]*RouteMetadata),
	}
}

// CheckIfPublic checks if any route for the given path is public
func (rr *RouteRegistry) CheckIfPublic(path string) bool {
	for _, metadata := range rr.routes {
		if metadata.Path == path && metadata.IsPublic {
			return true
		}
	}
	return false
}

// Register registers a route with its metadata
func (rr *RouteRegistry) Register(metadata *RouteMetadata) error {
	if err := metadata.Validate(); err != nil {
		return err
	}

	key := fmt.Sprintf("%s:%s", strings.ToUpper(metadata.Method), metadata.Path)
	if _, exists := rr.routes[key]; exists {
		return fmt.Errorf("route already registered: %s", key)
	}

	rr.routes[key] = metadata
	return nil
}

// RegisterMany registers multiple routes
func (rr *RouteRegistry) RegisterMany(metadatas ...*RouteMetadata) error {
	for _, metadata := range metadatas {
		if err := rr.Register(metadata); err != nil {
			return err
		}
	}
	return nil
}

// Get retrieves route metadata by path and method
// Supports both exact matches and pattern matching for parameterized routes
func (rr *RouteRegistry) Get(path, method string) (*RouteMetadata, bool) {
	upperMethod := strings.ToUpper(method)

	// First try exact match
	key := fmt.Sprintf("%s:%s", upperMethod, path)
	if metadata, exists := rr.routes[key]; exists {
		return metadata, exists
	}

	// If exact match fails, try pattern matching for parameterized routes
	// Convert numeric IDs in the path to :id and search again
	normalizedPath := normalizePathForPattern(path)
	if normalizedPath != path {
		key := fmt.Sprintf("%s:%s", upperMethod, normalizedPath)
		if metadata, exists := rr.routes[key]; exists {
			return metadata, exists
		}
	}

	return nil, false
}

// normalizePathForPattern converts numeric path segments to :id pattern
// e.g., /api/v1/staff/qualification/630 -> /api/v1/staff/qualification/:id
func normalizePathForPattern(path string) string {
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	for i := 0; i < len(parts); i++ {
		if isNumericSegment(parts[i]) {
			parts[i] = ":id"
		}
	}
	return strings.Join(parts, "/")
}

// isNumericSegment checks if a string is purely numeric
func isNumericSegment(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// GetAll returns all registered routes
func (rr *RouteRegistry) GetAll() map[string]*RouteMetadata {
	result := make(map[string]*RouteMetadata)
	for k, v := range rr.routes {
		result[k] = v
	}
	return result
}

// GetByTag returns all routes with a specific tag
func (rr *RouteRegistry) GetByTag(tag string) []*RouteMetadata {
	var result []*RouteMetadata
	for _, metadata := range rr.routes {
		for _, t := range metadata.Tags {
			if t == tag {
				result = append(result, metadata)
				break
			}
		}
	}
	return result
}

// GetDeprecatedRoutes returns all deprecated routes
func (rr *RouteRegistry) GetDeprecatedRoutes() []*RouteMetadata {
	var result []*RouteMetadata
	for _, metadata := range rr.routes {
		if metadata.Deprecated {
			result = append(result, metadata)
		}
	}
	return result
}

// GetByAPIVersion returns all routes for a specific API version
func (rr *RouteRegistry) GetByAPIVersion(version string) []*RouteMetadata {
	var result []*RouteMetadata
	for _, metadata := range rr.routes {
		if metadata.APIVersion == version {
			result = append(result, metadata)
		}
	}
	return result
}

// GetByRole returns all routes accessible by a specific role
func (rr *RouteRegistry) GetByRole(role string) []*RouteMetadata {
	var result []*RouteMetadata
	for _, metadata := range rr.routes {
		for _, r := range metadata.AllowedRoles {
			if r == role {
				result = append(result, metadata)
				break
			}
		}
	}
	return result
}

// Count returns the total number of registered routes
func (rr *RouteRegistry) Count() int {
	return len(rr.routes)
}

// Clear clears all registered routes
func (rr *RouteRegistry) Clear() {
	rr.routes = make(map[string]*RouteMetadata)
}

// ToSwaggerTags converts route metadata to Swagger operation tags and descriptions
func (rm *RouteMetadata) ToSwaggerInfo() map[string]interface{} {
	info := map[string]interface{}{
		"path":    rm.Path,
		"method":  rm.Method,
		"summary": rm.Description,
		"tags":    rm.Tags,
	}

	if rm.Deprecated {
		info["deprecated"] = true
		if rm.ReplacedBy != "" {
			info["x-deprecated-use-instead"] = rm.ReplacedBy
		}
	}

	if rm.RateLimit != nil {
		info["x-rate-limit"] = map[string]interface{}{
			"default": rm.RateLimit.DefaultRequestsPerMinute,
			"burst":   rm.RateLimit.BurstAllowance,
		}
	}

	return info
}
