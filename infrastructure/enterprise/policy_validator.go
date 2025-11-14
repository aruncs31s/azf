package enterprise

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aruncs31s/azf/utils"
)

// PolicyValidationReport holds results of policy validation
type PolicyValidationReport struct {
	IsValid            bool
	Errors             []string
	Warnings           []string
	DeadPolicies       []*DeadPolicy
	Conflicts          []*PolicyConflict
	UnregisteredRoutes []*UnregisteredRoute
	SummaryStats       *PolicySummary
}

// DeadPolicy represents a policy that has no corresponding route
type DeadPolicy struct {
	Role     string // e.g., "admin"
	Resource string // e.g., "/api/v1/staff/profile"
	Action   string // e.g., "GET"
	Reason   string // Why it's considered dead
}

// PolicyConflict represents conflicting policies
type PolicyConflict struct {
	Type        string // "DUPLICATE", "CONTRADICTING", "OVERLAPPING"
	Policy1     *PolicyPattern
	Policy2     *PolicyPattern
	Description string
}

// UnregisteredRoute represents a route without a policy
type UnregisteredRoute struct {
	Method       string
	Path         string
	AllowedRoles []string
	Reason       string
}

// PolicySummary contains policy statistics
type PolicySummary struct {
	TotalPolicies      int
	TotalRoles         int
	TotalResources     int
	TotalRoutes        int
	CoveredRoutes      int
	UncoveredRoutes    int
	CoveragePercentage float64
	DeprecatedRoutes   int
	PublicRoutes       int
}

// PolicyPattern represents a Casbin policy
type PolicyPattern struct {
	Role     string
	Resource string
	Action   string
}

type PolicyValidator interface {
	AddPolicy(role, resource, action string)
	AddPolicies(policies []*PolicyPattern)
	Validate() *PolicyValidationReport
}

// PolicyValidator validates Casbin policies against registered routes
type policyValidator struct {
	registry      *RouteRegistry
	policies      []*PolicyPattern
	routeMetadata map[string]*RouteMetadata
}

// NewPolicyValidator creates a new policy validator
func NewPolicyValidator(registry *RouteRegistry) PolicyValidator {
	return &policyValidator{
		registry:      registry,
		policies:      make([]*PolicyPattern, 0),
		routeMetadata: make(map[string]*RouteMetadata),
	}
}

// AddPolicy adds a policy for validation
func (pv *policyValidator) AddPolicy(role, resource, action string) {
	pv.policies = append(pv.policies, &PolicyPattern{
		Role:     role,
		Resource: resource,
		Action:   action,
	})
}

// AddPolicies adds multiple policies
func (pv *policyValidator) AddPolicies(policies []*PolicyPattern) {
	pv.policies = append(pv.policies, policies...)
}

// Validate performs comprehensive policy validation
func (pv *policyValidator) Validate() *PolicyValidationReport {
	report := &PolicyValidationReport{
		IsValid:            true,
		Errors:             make([]string, 0),
		Warnings:           make([]string, 0),
		DeadPolicies:       make([]*DeadPolicy, 0),
		Conflicts:          make([]*PolicyConflict, 0),
		UnregisteredRoutes: make([]*UnregisteredRoute, 0),
		SummaryStats:       &PolicySummary{},
	}

	// Check for policy syntax errors
	pv.validatePolicySyntax(report)

	// Check for dead policies (policies with no corresponding routes)
	pv.findDeadPolicies(report)

	// Check for unregistered routes (routes with no policies)
	pv.findUnregisteredRoutes(report)

	// Check for conflicts
	pv.findConflicts(report)

	// Generate statistics
	pv.generateSummary(report)

	// Set overall validity
	report.IsValid = len(report.Errors) == 0

	return report
}

// validatePolicySyntax checks for syntax errors in policies
func (pv *policyValidator) validatePolicySyntax(report *PolicyValidationReport) {
	seenPolicies := make(map[string]bool)

	for i, policy := range pv.policies {
		// Check required fields
		if policy.Role == "" {
			report.Errors = append(report.Errors, fmt.Sprintf("Policy %d: Role cannot be empty", i))
			continue
		}
		if policy.Resource == "" {
			report.Errors = append(report.Errors, fmt.Sprintf("Policy %d: Resource cannot be empty", i))
			continue
		}
		if policy.Action == "" {
			report.Errors = append(report.Errors, fmt.Sprintf("Policy %d: Action cannot be empty", i))
			continue
		}

		// Check for invalid HTTP method
		validMethods := map[string]bool{
			"GET": true, "POST": true, "PUT": true, "DELETE": true,
			"PATCH": true, "OPTIONS": true, "HEAD": true,
		}
		if !validMethods[policy.Action] {
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("Policy %d: Invalid HTTP method '%s' for %s %s", i, policy.Action, policy.Action, policy.Resource))
		}

		// Check for duplicate policies
		key := fmt.Sprintf("%s:%s:%s", policy.Role, policy.Resource, policy.Action)
		if seenPolicies[key] {
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("Policy %d: Duplicate policy detected: %s %s for role %s", i, policy.Action, policy.Resource, policy.Role))
		}
		seenPolicies[key] = true

		// Check path format
		if !strings.HasPrefix(policy.Resource, "/") {
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("Policy %d: Resource path should start with '/': %s", i, policy.Resource))
		}
	}
}

// findDeadPolicies identifies policies with no corresponding routes
func (pv *policyValidator) findDeadPolicies(report *PolicyValidationReport) {
	for _, policy := range pv.policies {
		// Try to find a matching route
		found := false
		normalizedPolicyPath := utils.NormalizePathForLookup(policy.Resource)

		// Check exact matches
		if metadata, exists := pv.registry.Get(policy.Resource, policy.Action); exists {
			found = true
			// Check if role is allowed
			roleAllowed := false
			for _, role := range metadata.AllowedRoles {
				if role == policy.Role {
					roleAllowed = true
					break
				}
			}
			if !roleAllowed && !metadata.IsPublic {
				report.Warnings = append(report.Warnings,
					fmt.Sprintf("Policy for role '%s' on %s %s but route only allows roles: %v",
						policy.Role, policy.Action, policy.Resource, metadata.AllowedRoles))
			}
		}

		// Check pattern matches (e.g., :id patterns)
		if !found {
			for _, metadata := range pv.registry.GetAll() {
				if pv.pathsMatch(normalizedPolicyPath, utils.NormalizePathForLookup(metadata.Path)) &&
					strings.EqualFold(policy.Action, metadata.Method) {
					found = true
					break
				}
			}
		}

		if !found {
			report.DeadPolicies = append(report.DeadPolicies, &DeadPolicy{
				Role:     policy.Role,
				Resource: policy.Resource,
				Action:   policy.Action,
				Reason:   "No corresponding route registered",
			})
			report.Warnings = append(report.Warnings,
				fmt.Sprintf("Dead policy: %s %s for role '%s' has no corresponding registered route",
					policy.Action, policy.Resource, policy.Role))
		}
	}
}

// findUnregisteredRoutes identifies routes without policies
func (pv *policyValidator) findUnregisteredRoutes(report *PolicyValidationReport) {
	for _, metadata := range pv.registry.GetAll() {
		if metadata.IsPublic {
			// Public routes don't need policies
			continue
		}

		found := false
		for _, policy := range pv.policies {
			if pv.pathsMatch(policy.Resource, metadata.Path) &&
				strings.EqualFold(policy.Action, metadata.Method) {
				found = true
				break
			}
		}

		if !found {
			report.UnregisteredRoutes = append(report.UnregisteredRoutes, &UnregisteredRoute{
				Method:       metadata.Method,
				Path:         metadata.Path,
				AllowedRoles: metadata.AllowedRoles,
				Reason:       fmt.Sprintf("Route %s %s has no Casbin policy", metadata.Method, metadata.Path),
			})
			report.Errors = append(report.Errors,
				fmt.Sprintf("Missing policy: %s %s (allowed roles: %v)",
					metadata.Method, metadata.Path, metadata.AllowedRoles))
		}
	}
}

// findConflicts identifies conflicting policies
func (pv *policyValidator) findConflicts(report *PolicyValidationReport) {
	// Check for duplicate policies
	seenPolicies := make(map[string]int)
	for i, policy := range pv.policies {
		key := fmt.Sprintf("%s:%s:%s", policy.Role, policy.Resource, policy.Action)
		if idx, seen := seenPolicies[key]; seen {
			report.Conflicts = append(report.Conflicts, &PolicyConflict{
				Type:        "DUPLICATE",
				Policy1:     pv.policies[idx],
				Policy2:     policy,
				Description: fmt.Sprintf("Duplicate policies for %s %s (role: %s)", policy.Action, policy.Resource, policy.Role),
			})
		}
		seenPolicies[key] = i
	}

	// Check for overlapping policies
	for i := 0; i < len(pv.policies); i++ {
		for j := i + 1; j < len(pv.policies); j++ {
			p1 := pv.policies[i]
			p2 := pv.policies[j]

			if p1.Role == p2.Role && p1.Action == p2.Action {
				// Check if paths overlap (e.g., /api/v1/staff and /api/v1/staff/:id)
				if pv.pathsOverlap(p1.Resource, p2.Resource) {
					report.Conflicts = append(report.Conflicts, &PolicyConflict{
						Type:        "OVERLAPPING",
						Policy1:     p1,
						Policy2:     p2,
						Description: fmt.Sprintf("Overlapping paths for %s %s", p1.Action, p1.Resource),
					})
				}
			}
		}
	}
}

// generateSummary generates policy statistics
func (pv *policyValidator) generateSummary(report *PolicyValidationReport) {
	stats := report.SummaryStats

	// Count unique resources, roles, and actions
	roles := make(map[string]bool)
	resources := make(map[string]bool)

	for _, policy := range pv.policies {
		roles[policy.Role] = true
		resources[policy.Resource] = true
	}

	stats.TotalPolicies = len(pv.policies)
	stats.TotalRoles = len(roles)
	stats.TotalResources = len(resources)
	stats.TotalRoutes = pv.registry.Count()
	stats.CoveredRoutes = stats.TotalRoutes - len(report.UnregisteredRoutes)
	stats.UncoveredRoutes = len(report.UnregisteredRoutes)

	if stats.TotalRoutes > 0 {
		stats.CoveragePercentage = float64(stats.CoveredRoutes) / float64(stats.TotalRoutes) * 100
	}

	// Count deprecated and public routes
	for _, metadata := range pv.registry.GetAll() {
		if metadata.Deprecated {
			stats.DeprecatedRoutes++
		}
		if metadata.IsPublic {
			stats.PublicRoutes++
		}
	}
}

// pathsMatch checks if two paths match (considering :id patterns)
func (pv *policyValidator) pathsMatch(path1, path2 string) bool {
	// Exact match
	if path1 == path2 {
		return true
	}

	// Normalize both paths
	norm1 := utils.NormalizePathForLookup(path1)
	norm2 := utils.NormalizePathForLookup(path2)

	return norm1 == norm2
}

// pathsOverlap checks if two paths have overlapping coverage
func (pv *policyValidator) pathsOverlap(path1, path2 string) bool {
	// Simple overlap check: if one is a prefix of the other
	p1Parts := strings.Split(strings.TrimPrefix(path1, "/"), "/")
	p2Parts := strings.Split(strings.TrimPrefix(path2, "/"), "/")

	minLen := len(p1Parts)
	if len(p2Parts) < minLen {
		minLen = len(p2Parts)
	}

	// Check if all common parts match
	for i := 0; i < minLen; i++ {
		if !pv.partsMatch(p1Parts[i], p2Parts[i]) {
			return false
		}
	}

	return true
}

// partsMatch checks if two path parts match (considering :id patterns)
func (pv *policyValidator) partsMatch(part1, part2 string) bool {
	if part1 == part2 {
		return true
	}
	// Check if either is a parameter
	if strings.HasPrefix(part1, ":") || strings.HasPrefix(part2, ":") {
		return true
	}
	// Check if either is numeric (potential ID)
	if isNumeric(part1) || isNumeric(part2) {
		return true
	}
	return false
}

// NormalizePathForPolicy converts a concrete path to policy pattern
func NormalizePathForPolicy(path string) string {
	// Replace numeric IDs with :id placeholder
	re := regexp.MustCompile(`/\d+(/|$)`)
	normalized := re.ReplaceAllString(path, "/:id$1")
	return normalized
}

// isNumeric checks if a string is numeric
func isNumeric(s string) bool {
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

// String returns a formatted report string
func (report *PolicyValidationReport) String() string {
	var sb strings.Builder

	sb.WriteString("=== POLICY VALIDATION REPORT ===\n")
	sb.WriteString(fmt.Sprintf("Valid: %v\n", report.IsValid))
	sb.WriteString("\n--- STATISTICS ---\n")
	if report.SummaryStats != nil {
		stats := report.SummaryStats
		sb.WriteString(fmt.Sprintf("Total Policies: %d\n", stats.TotalPolicies))
		sb.WriteString(fmt.Sprintf("Total Roles: %d\n", stats.TotalRoles))
		sb.WriteString(fmt.Sprintf("Total Resources: %d\n", stats.TotalResources))
		sb.WriteString(fmt.Sprintf("Total Routes: %d\n", stats.TotalRoutes))
		sb.WriteString(fmt.Sprintf("Covered Routes: %d/%d (%.1f%%)\n", stats.CoveredRoutes, stats.TotalRoutes, stats.CoveragePercentage))
		sb.WriteString(fmt.Sprintf("Deprecated Routes: %d\n", stats.DeprecatedRoutes))
		sb.WriteString(fmt.Sprintf("Public Routes: %d\n", stats.PublicRoutes))
	}

	if len(report.Errors) > 0 {
		sb.WriteString("\n--- ERRORS ---\n")
		for _, err := range report.Errors {
			sb.WriteString(fmt.Sprintf("❌ %s\n", err))
		}
	}

	if len(report.Warnings) > 0 {
		sb.WriteString("\n--- WARNINGS ---\n")
		for _, warn := range report.Warnings {
			sb.WriteString(fmt.Sprintf("⚠️  %s\n", warn))
		}
	}

	if len(report.DeadPolicies) > 0 {
		sb.WriteString("\n--- DEAD POLICIES ---\n")
		for _, dp := range report.DeadPolicies {
			sb.WriteString(fmt.Sprintf("🪦 %s %s (role: %s) - %s\n", dp.Action, dp.Resource, dp.Role, dp.Reason))
		}
	}

	if len(report.UnregisteredRoutes) > 0 {
		sb.WriteString("\n--- UNREGISTERED ROUTES ---\n")
		for _, ur := range report.UnregisteredRoutes {
			sb.WriteString(fmt.Sprintf("⚠️  %s %s (roles: %v) - %s\n", ur.Method, ur.Path, ur.AllowedRoles, ur.Reason))
		}
	}

	if len(report.Conflicts) > 0 {
		sb.WriteString("\n--- CONFLICTS ---\n")
		for _, conflict := range report.Conflicts {
			sb.WriteString(fmt.Sprintf("⚡ [%s] %s\n", conflict.Type, conflict.Description))
		}
	}

	return sb.String()
}
