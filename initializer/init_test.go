package initializer_test

import (
	"os"
	"testing"

	"github.com/aruncs31s/azf/initializer"
	"github.com/aruncs31s/azf/shared/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	// initialize shared logger for tests
	logger.InitLogger()

	// Ensure config directory exists and write minimal Casbin model + policy used by tests.
	_ = os.MkdirAll("config", 0755)

	model := `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`
	_ = os.WriteFile("config/casbin_rbac_model.conf", []byte(model), 0644)

	policy := `p, staff, /api/v1/staff/profile, GET
	g, user, staff
	`
	_ = os.WriteFile("config/casbin_rbac_policy.csv", []byte(policy), 0644)

	// Run tests
	code := m.Run()

	// Cleanup test-created files
	_ = os.Remove("config/casbin_rbac_model.conf")
	_ = os.Remove("config/casbin_rbac_policy.csv")

	os.Exit(code)
}

// TestInitLocalDB_WithProvidedDB ensures that when a non-nil *gorm.DB is passed to InitLocalDB,
// the function returns early and does not replace the global initializer.DB.
func TestInitLocalDB_WithProvidedDB(t *testing.T) {
	t.Parallel()

	// create an in-memory sqlite DB
	memDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	// preserve previous global DB and restore at the end
	prev := initializer.DB
	defer func() { initializer.DB = prev }()

	// ensure global DB is nil for test isolation
	initializer.DB = nil

	// call InitLocalDB with a provided DB - per current implementation this should return early
	initializer.InitLocalDB(memDB)

	if initializer.DB != nil {
		t.Fatalf("expected initializer.DB to remain nil when provided a non-nil db, got non-nil")
	}
}

// TestInitLocalDB_CreatesLocalDB_WhenNil verifies that calling InitLocalDB with nil creates
// a local sqlite database at the configured path and sets initializer.DB.
func TestInitLocalDB_CreatesLocalDB_WhenNil(t *testing.T) {
	t.Parallel()

	// preserve previous global DB and restore at the end
	prev := initializer.DB
	defer func() { initializer.DB = prev }()

	// remove any existing file so we start fresh
	dbPath := "/tmp/AZF_auth_z.db"
	_ = os.Remove(dbPath)

	initializer.DB = nil

	initializer.InitLocalDB(nil)

	if initializer.DB == nil {
		t.Fatalf("expected initializer.DB to be set after InitLocalDB(nil)")
	}

	// Try to close underlying sql DB if possible
	sqlDB, err := initializer.DB.DB()
	if err == nil && sqlDB != nil {
		_ = sqlDB.Close()
	}

	// cleanup created DB file
	_ = os.Remove(dbPath)
}

// TestInitCasbin_PolicyAndRoleOps tests Casbin initialization and some dynamic policy/role operations.
func TestInitCasbin_PolicyAndRoleOps(t *testing.T) {
	t.Parallel()

	// InitCasbin is designed to be idempotent; call it and expect no error.
	if err := initializer.InitCasbin(); err != nil {
		t.Fatalf("InitCasbin returned error: %v", err)
	}

	en := initializer.GetCasbinEnforcer()
	if en == nil {
		t.Fatalf("expected Casbin enforcer to be non-nil after InitCasbin")
	}

	// There is an existing mapping in the provided policy: user -> staff and staff has access to /api/v1/staff/profile GET
	allowed := initializer.CheckPermission("user", "/api/v1/staff/profile", "GET")
	if !allowed {
		t.Fatalf("expected 'user' to have permission for /api/v1/staff/profile GET based on bundled policy")
	}

	// Test adding a new role policy and assigning it to a user
	role := "testrole"
	resource := "/some/test"
	action := "GET"

	added := initializer.AddRolePolicy(role, resource, action)
	if !added {
		t.Fatalf("expected AddRolePolicy to return true when adding a new policy")
	}

	assigned := initializer.AssignRoleToUser("alice", role)
	if !assigned {
		// try to remove the policy before failing to keep state clean
		_ = initializer.RemoveRolePolicy(role, resource, action)
		t.Fatalf("expected AssignRoleToUser to return true")
	}

	// Now alice should have permission via the assigned role
	if !initializer.CheckPermission("alice", resource, action) {
		t.Fatalf("expected 'alice' to have permission for %s %s via role %s", action, resource, role)
	}

	// Roles for user should include the newly assigned role
	roles := initializer.GetRolesForUser("alice")
	found := false
	for _, r := range roles {
		if r == role {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected GetRolesForUser to include %q, got: %v", role, roles)
	}

	// Clean up: remove policy and role assignment
	removed := initializer.RemoveRolePolicy(role, resource, action)
	if !removed {
		t.Fatalf("expected RemoveRolePolicy to return true")
	}
	removedRole := initializer.RemoveRoleFromUser("alice", role)
	if !removedRole {
		t.Fatalf("expected RemoveRoleFromUser to return true")
	}

	// After removal the permission should no longer hold
	if initializer.CheckPermission("alice", resource, action) {
		t.Fatalf("expected permission to be revoked after removing policy and role assignment")
	}
}
