package config

// GetEnvironment returns the current environment
func GetEnvironment() string {
	env := GetEnvironmentVal()
	if env == "" {
		env = "development"
	}
	return env
}
