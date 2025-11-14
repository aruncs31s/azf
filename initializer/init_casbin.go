package initializer

import (
	"log"
	"sync"

	"github.com/aruncs31s/azf/config"
	"github.com/aruncs31s/azf/shared/logger"
	"github.com/casbin/casbin/v2"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"go.uber.org/zap"
)

var (
	CasbinEnforcer    *casbin.Enforcer
	casbinMutex       sync.RWMutex
	casbinInitialized bool
)

// InitCasbin initializes the Casbin enforcer with RBAC model and policy
// Thread-safe initialization with idempotency guard.
// Returns an error when initialization fails instead of terminating the process.
func InitCasbin() error {
	casbinMutex.Lock()
	defer casbinMutex.Unlock()

	// Already initialized, return early
	if casbinInitialized {
		// Use the logger wrapper to avoid direct access to the underlying Logger which may be nil
		logger.Debug("Casbin already initialized, skipping re-initialization")
		return nil
	}

	var err error

	// Create file adapter for the policy file
	adapter := fileadapter.NewAdapter(config.CASBIN_POLICY_FILE)

	// Create enforcer with model config and adapter
	CasbinEnforcer, err = casbin.NewEnforcer(config.CASBIN_MODEL_FILE, adapter)
	if err != nil {
		// Log the error and return it so callers can decide how to handle it
		logger.Error("Failed to create Casbin enforcer", zap.Error(err), zap.String("file", config.CASBIN_POLICY_FILE))
		return err
	}

	// Load policy from file
	err = CasbinEnforcer.LoadPolicy()
	if err != nil {
		// Log the error and return it
		logger.Error("Failed to load Casbin policy", zap.Error(err), zap.String("file", config.CASBIN_POLICY_FILE))
		return err
	}

	casbinInitialized = true

	// Informational log using wrapper
	logger.Info("Casbin initialized successfully", zap.String("file", config.CASBIN_POLICY_FILE))
	return nil
}

// GetCasbinEnforcer safely retrieves the initialized Casbin enforcer
// Returns nil if not yet initialized
func GetCasbinEnforcer() *casbin.Enforcer {
	casbinMutex.RLock()
	defer casbinMutex.RUnlock()
	return CasbinEnforcer
}

// IsCasbinInitialized checks if Casbin has been initialized
func IsCasbinInitialized() bool {
	casbinMutex.RLock()
	defer casbinMutex.RUnlock()
	return casbinInitialized
}

// AddRolePolicy adds a new role policy dynamically
// Example: AddRolePolicy("admin", "/api/v1/admin/staff", "POST")
func AddRolePolicy(role, resource, action string) bool {
	enforcer := GetCasbinEnforcer()
	if enforcer == nil {
		log.Printf("Casbin enforcer not initialized")
		return false
	}
	added, err := enforcer.AddPolicy(role, resource, action)
	if err != nil {
		log.Printf("Failed to add policy: %v", err)
		return false
	}
	return added
}

// RemoveRolePolicy removes a role policy dynamically
// Example: RemoveRolePolicy("admin", "/api/v1/admin/staff", "POST")
func RemoveRolePolicy(role, resource, action string) bool {
	enforcer := GetCasbinEnforcer()
	if enforcer == nil {
		log.Printf("Casbin enforcer not initialized")
		return false
	}
	removed, err := enforcer.RemovePolicy(role, resource, action)
	if err != nil {
		log.Printf("Failed to remove policy: %v", err)
		return false
	}
	return removed
}

// AssignRoleToUser assigns a role to a user (via their role string)
// Example: AssignRoleToUser("user123", "staff")
func AssignRoleToUser(user, role string) bool {
	enforcer := GetCasbinEnforcer()
	if enforcer == nil {
		log.Printf("Casbin enforcer not initialized")
		return false
	}
	added, err := enforcer.AddGroupingPolicy(user, role)
	if err != nil {
		log.Printf("Failed to assign role: %v", err)
		return false
	}
	return added
}

// RemoveRoleFromUser removes a role from a user
func RemoveRoleFromUser(user, role string) bool {
	enforcer := GetCasbinEnforcer()
	if enforcer == nil {
		log.Printf("Casbin enforcer not initialized")
		return false
	}
	removed, err := enforcer.RemoveGroupingPolicy(user, role)
	if err != nil {
		log.Printf("Failed to remove role: %v", err)
		return false
	}
	return removed
}

// CheckPermission checks if a user has permission to access a resource
// Returns true if user has permission, false otherwise
func CheckPermission(user, resource, action string) bool {
	enforcer := GetCasbinEnforcer()
	if enforcer == nil {
		log.Printf("Casbin enforcer not initialized")
		return false
	}
	allowed, err := enforcer.Enforce(user, resource, action)
	if err != nil {
		log.Printf("Error checking permission: %v", err)
		return false
	}
	return allowed
}

// GetRolesForUser gets all roles assigned to a user
func GetRolesForUser(user string) []string {
	enforcer := GetCasbinEnforcer()
	if enforcer == nil {
		log.Printf("Casbin enforcer not initialized")
		return []string{}
	}
	roles, err := enforcer.GetRolesForUser(user)
	if err != nil {
		log.Printf("Error getting roles: %v", err)
		return []string{}
	}
	return roles
}
