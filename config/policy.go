package config

const (
	POLICY_VERSION = 1
)

const (
	RATE_LIMITING            = true
	AUDIT_LOGING             = true
	DEPRICATION_CHECK        = true
	ENABLE_NON_POLICY_ROUTES = false
)

const (
	AUTH_MODE_SOFT_MIGRATION  = "SOFT_MIGRATION"
	AUTH_MODE_GRADUAL_ROLLOUT = "GRADUAL_ROLLOUT"
	AUTH_MODE_CASBIN          = "CASBIN_V2"
)
