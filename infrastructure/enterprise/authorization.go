package enterprise

import (
	"log"

	"github.com/aruncs31s/azf/config"

	"github.com/aruncs31s/azf/constants"
	"github.com/casbin/casbin/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// EnterpriseAuth holds the enterprise authorization setup
// Type of *EnterpriseAuthorizationSetup
var EnterpriseAuth *EnterpriseAuthorizationSetup

// InitEnterpriseAuth initializes the enterprise authorization system
func IniAuthorization(
	db *gorm.DB,
	reddis *redis.Client,
	enforcer *casbin.Enforcer,
	logger *zap.Logger,

) error {
	setupOpts := &SetupOptions{
		Database:               db,
		Redis:                  reddis,
		PolicyFilePath:         config.CASBIN_POLICY_FILE,
		Environment:            config.GetEnvironment(),
		EnableAuditLogging:     config.AUDIT_LOGING,
		EnableRateLimit:        config.RATE_LIMITING,
		EnableDeprecationCheck: config.DEPRICATION_CHECK,
		GradualRolloutMode:     config.GetEnvironment() == constants.APP_SAGING,
		AllowMissingPolicies:   config.GetEnvironment() == constants.APP_DEVELOPMENT || config.ENABLE_NON_POLICY_ROUTES,
		ValidatePoliciesOnInit: config.GetEnvironment() != constants.APP_DEVELOPMENT,
		CasbinEnforcer:         enforcer,
		Logger:                 logger,
	}

	setup, err := NewEnterpriseAuthorizationSetup(setupOpts)
	if err != nil {
		log.Fatal("Failed to initialize authorization:", err)
		return err
	}
	EnterpriseAuth = setup
	return nil
}
