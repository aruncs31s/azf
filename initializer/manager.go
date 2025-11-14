package initializer

import (
	"errors"
	"log"
	"os"
	"sync"

	"github.com/aruncs31s/azf/config"
	"github.com/aruncs31s/azf/domain/api_usage"
	"github.com/aruncs31s/azf/infrastructure/persistence"
	"github.com/aruncs31s/azf/shared/logger"
	"github.com/casbin/casbin/v2"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Manager encapsulates initialization state (DB, Casbin enforcer, logger) to avoid package-level globals.
// It provides thread-safe operations for Casbin and lifecycle control for DB and enforcer.
type Manager struct {
	mu          sync.RWMutex
	DB          *gorm.DB
	Enforcer    *casbin.Enforcer
	logger      *zap.Logger
	initialized bool
}

// NewManager constructs a Manager. Any of the args may be nil to let the Manager initialize them later.
// - db: optional pre-created *gorm.DB to use (if non-nil the Manager will adopt it)
// - enforcer: optional pre-created *casbin.Enforcer to use
// - l: optional *zap.Logger; if nil the Manager will use package shared logger
func NewManager(db *gorm.DB, enforcer *casbin.Enforcer, l *zap.Logger) *Manager {
	if l == nil {
		// Ensure the shared logger is initialized and use it
		logger.InitLogger()
		l = logger.GetLogger()
	}
	return &Manager{
		DB:       db,
		Enforcer: enforcer,
		logger:   l,
	}
}

// InitLocalDB ensures the Manager has a working *gorm.DB.
// If a non-nil tempDB is passed, it will be used as-is. Otherwise the Manager attempts to
// open a file-backed sqlite DB at /tmp/AZF_auth_z.db, falling back to an in-memory DB if the file cannot be created.
func (m *Manager) InitLocalDB(tempDB *gorm.DB) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If a DB already exists on the manager, do nothing
	if m.DB != nil {
		m.logger.Debug("Local DB already initialized on manager")
		return nil
	}

	// Use provided DB if available
	if tempDB != nil {
		m.DB = tempDB
		m.logger.Debug("Using provided DB in Manager")
		// run migrations
		if err := migrateTable(m.DB); err != nil {
			m.logger.Error("migration failed on provided DB", zap.Error(err))
			return err
		}
		return nil
	}

	// Try to create/open file-backed DB
	localDBPath := "/tmp/AZF_auth_z.db"
	// best-effort ensure directory exists
	_ = os.MkdirAll("/tmp", 0755)

	db, err := gorm.Open(sqlite.Open(localDBPath), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		// fallback to in-memory DB
		m.logger.Error("failed to open file sqlite DB, falling back to in-memory", zap.Error(err))
		memDB, memErr := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
			SkipDefaultTransaction: true,
		})
		if memErr != nil {
			m.logger.Error("failed to open in-memory sqlite fallback", zap.Error(memErr))
			return memErr
		}
		m.DB = memDB
		if err := migrateTable(m.DB); err != nil {
			m.logger.Error("migration failed on in-memory DB", zap.Error(err))
			return err
		}
		m.logger.Info("initialized in-memory local DB")
		return nil
	}

	m.DB = db
	if err := migrateTable(m.DB); err != nil {
		m.logger.Error("migration failed on file-backed DB", zap.Error(err))
		return err
	}
	m.logger.Info("initialized file-backed local DB", zap.String("path", localDBPath))
	return nil
}

// migrateTable runs AutoMigrate for API usage tables on the provided DB.
// kept as a helper to keep initialization logic together.
func migrateTable(db *gorm.DB) error {
	if db == nil {
		return errors.New("migrateTable: db is nil")
	}
	if err := db.AutoMigrate(
		api_usage.APIUsageStats{},
		api_usage.APIUsageLog{},
		&persistence.UserModel{},
	); err != nil {
		return err
	}
	return nil
}

// InitCasbin initializes the casbin enforcer if it does not already exist on the Manager.
// It accepts optional model and policy paths; if either is empty, defaults from config are used.
func (m *Manager) InitCasbin(modelPath, policyPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Enforcer != nil && m.initialized {
		m.logger.Debug("Casbin already initialized on manager")
		return nil
	}

	if modelPath == "" {
		modelPath = config.CASBIN_MODEL_FILE
	}
	if policyPath == "" {
		policyPath = config.CASBIN_POLICY_FILE
	}

	adapter := fileadapter.NewAdapter(policyPath)
	enf, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		m.logger.Error("failed to create Casbin enforcer", zap.Error(err), zap.String("policy", policyPath), zap.String("model", modelPath))
		return err
	}

	if err := enf.LoadPolicy(); err != nil {
		m.logger.Error("failed to load Casbin policy", zap.Error(err), zap.String("policy", policyPath))
		return err
	}

	m.Enforcer = enf
	m.initialized = true
	m.logger.Info("Casbin initialized on manager", zap.String("policy", policyPath))
	return nil
}

// GetEnforcer returns the manager's Casbin enforcer. May be nil if not initialized.
func (m *Manager) GetEnforcer() *casbin.Enforcer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Enforcer
}

// AddRolePolicy wraps Enforcer.AddPolicy in a safe way.
func (m *Manager) AddRolePolicy(role, resource, action string) (bool, error) {
	m.mu.RLock()
	enf := m.Enforcer
	m.mu.RUnlock()

	if enf == nil {
		return false, errors.New("enforcer not initialized")
	}
	added, err := enf.AddPolicy(role, resource, action)
	return added, err
}

// RemoveRolePolicy wraps Enforcer.RemovePolicy in a safe way.
func (m *Manager) RemoveRolePolicy(role, resource, action string) (bool, error) {
	m.mu.RLock()
	enf := m.Enforcer
	m.mu.RUnlock()

	if enf == nil {
		return false, errors.New("enforcer not initialized")
	}
	removed, err := enf.RemovePolicy(role, resource, action)
	return removed, err
}

// AssignRoleToUser adds a grouping policy (assigns role to user).
func (m *Manager) AssignRoleToUser(user, role string) (bool, error) {
	m.mu.RLock()
	enf := m.Enforcer
	m.mu.RUnlock()

	if enf == nil {
		return false, errors.New("enforcer not initialized")
	}
	added, err := enf.AddGroupingPolicy(user, role)
	return added, err
}

// RemoveRoleFromUser removes a grouping policy.
func (m *Manager) RemoveRoleFromUser(user, role string) (bool, error) {
	m.mu.RLock()
	enf := m.Enforcer
	m.mu.RUnlock()

	if enf == nil {
		return false, errors.New("enforcer not initialized")
	}
	removed, err := enf.RemoveGroupingPolicy(user, role)
	return removed, err
}

// CheckPermission performs an enforcement check (user, resource, action).
func (m *Manager) CheckPermission(user, resource, action string) (bool, error) {
	m.mu.RLock()
	enf := m.Enforcer
	m.mu.RUnlock()

	if enf == nil {
		return false, errors.New("enforcer not initialized")
	}
	allowed, err := enf.Enforce(user, resource, action)
	if err != nil {
		return false, err
	}
	return allowed, nil
}

// GetRolesForUser returns roles assigned to a user.
func (m *Manager) GetRolesForUser(user string) ([]string, error) {
	m.mu.RLock()
	enf := m.Enforcer
	m.mu.RUnlock()

	if enf == nil {
		return nil, errors.New("enforcer not initialized")
	}
	roles, err := enf.GetRolesForUser(user)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// Close gracefully releases resources controlled by the Manager (e.g., DB).
// It will attempt to close the underlying sql.DB if available.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	if m.DB != nil {
		sqlDB, err := m.DB.DB()
		if err == nil && sqlDB != nil {
			if err2 := sqlDB.Close(); err2 != nil {
				m.logger.Error("error closing DB", zap.Error(err2))
				lastErr = err2
			}
		}
		m.DB = nil
	}

	// Casbin enforcer does not require an explicit close; just nil it out for GC.
	m.Enforcer = nil
	m.initialized = false

	return lastErr
}

// Ensure compatibility helpers: these helpers make it easier to migrate existing call sites.
// They are convenience wrappers that create a Manager, initialize components and return it.

// NewAndInitManager constructs a Manager and initializes DB and Casbin (using defaults).
// If a pre-existing casbin enforcer is provided it will be used instead of initializing from files.
func NewAndInitManager(tempDB *gorm.DB, casbinEnforcer *casbin.Enforcer, zapLogger *zap.Logger) (*Manager, error) {
	m := NewManager(tempDB, casbinEnforcer, zapLogger)

	// initialize DB (will use provided tempDB if non-nil)
	if err := m.InitLocalDB(tempDB); err != nil {
		// continue even if DB init fails (caller can decide), but log it
		log.Printf("warning: InitLocalDB returned error: %v", err)
	}

	// If a casbin enforcer was supplied at construction, mark initialized.
	if casbinEnforcer != nil {
		m.mu.Lock()
		m.Enforcer = casbinEnforcer
		m.initialized = true
		m.mu.Unlock()
		return m, nil
	}

	// otherwise initialize from default config paths
	if err := m.InitCasbin("", ""); err != nil {
		// return the error so callers can decide how to proceed
		return m, err
	}
	return m, nil
}
