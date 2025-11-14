package utils

import (
	"strings"
)

// NormalizePathForLookup converts actual paths with numeric IDs to policy patterns
// Example: /api/v1/staff/qualification/630 -> /api/v1/staff/qualification/:id
// This ensures policy matching works correctly in Casbin
func NormalizePathForLookup(path string) string {
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")

	// Replace any numeric ID with generic :id placeholder
	// This works for all resources automatically without manual pattern mapping
	for i := 0; i < len(parts); i++ {
		if isNumeric(parts[i]) {
			parts[i] = ":id"
		}
	}

	return strings.Join(parts, "/")
}

// isNumeric checks if a string is numeric (ID)
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
