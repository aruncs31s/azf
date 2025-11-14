package config
package config

import (
    // Because the , cwd changes when using the module outside
	_ "embed"  // Import for embedding

)

// Embed the default policy and model files
//go:embed casbin_rbac_policy.csv
var DefaultPolicy []byte

//go:embed casbin_rbac_model.conf
var DefaultModel []byte

const (
    CASBIN_MODEL_FILE          = "config/casbin_rbac_model.conf"  // Keep for backward compatibility, but we'll use embedded for defaults
    CASBIN_POLICY_FILE         = "config/casbin_rbac_policy.csv"
    CASBIN_POLICY_DEFAULT_PATH = "config/casbin_rbac_policy.csv"
)
