package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// JWT configuration
	JWT JWTConfig

	// Database configuration
	Database DBConfig

	// Rate limiting configuration
	RateLimit RateLimitCfg

	// Server configuration
	Server ServerConfig

	// Environment (development, staging, production)
	Environment string

	// Casbin configuration
	Casbin CasbinConfig
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
}

// DBConfig holds database-related configuration for the new config system
type DBConfig struct {
	Driver          string
	DSN             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// RateLimitConfig holds rate limiting configuration
type RateLimitCfg struct {
	Enabled           bool
	RequestsPerSecond float64
	Burst             int
	UseRedis          bool
	RedisURL          string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host            string
	Port            int
	BaseURL         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	AllowedOrigins  []string
}

// CasbinConfig holds Casbin configuration
type CasbinConfig struct {
	ModelPath  string
	PolicyPath string
}

// Validation errors
var (
	ErrJWTSecretRequired   = errors.New("JWT_SECRET environment variable is required")
	ErrJWTSecretTooShort   = errors.New("JWT_SECRET must be at least 32 characters for security")
	ErrJWTSecretWeak       = errors.New("JWT_SECRET should be at least 64 characters for enhanced security")
	ErrInvalidPort         = errors.New("invalid server port")
	ErrDatabaseDSNRequired = errors.New("database DSN is required")
)

// LoadConfig loads configuration from environment variables with validation
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// Load environment
	cfg.Environment = getEnvOrDefault("ENVIRONMENT", "development")

	// Load JWT config
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, ErrJWTSecretRequired
	}
	if len(jwtSecret) < 32 {
		return nil, ErrJWTSecretTooShort
	}

	cfg.JWT = JWTConfig{
		Secret:             jwtSecret,
		AccessTokenExpiry:  getDurationOrDefault("JWT_ACCESS_EXPIRY", 15*time.Minute),
		RefreshTokenExpiry: getDurationOrDefault("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
		Issuer:             getEnvOrDefault("JWT_ISSUER", "azf"),
	}

	// Load database config
	cfg.Database = DBConfig{
		Driver:          getEnvOrDefault("DB_DRIVER", "sqlite"),
		DSN:             getEnvOrDefault("DB_DSN", "tmp/AZF_auth_z.db"),
		MaxIdleConns:    getIntOrDefault("DB_MAX_IDLE_CONNS", 10),
		MaxOpenConns:    getIntOrDefault("DB_MAX_OPEN_CONNS", 100),
		ConnMaxLifetime: getDurationOrDefault("DB_CONN_MAX_LIFETIME", time.Hour),
		ConnMaxIdleTime: getDurationOrDefault("DB_CONN_MAX_IDLE_TIME", 30*time.Minute),
	}

	// Load rate limit config
	cfg.RateLimit = RateLimitCfg{
		Enabled:           getBoolOrDefault("RATE_LIMIT_ENABLED", true),
		RequestsPerSecond: getFloatOrDefault("RATE_LIMIT_RPS", 10.0),
		Burst:             getIntOrDefault("RATE_LIMIT_BURST", 20),
		UseRedis:          getBoolOrDefault("RATE_LIMIT_USE_REDIS", false),
		RedisURL:          getEnvOrDefault("REDIS_URL", ""),
	}

	// Load server config
	cfg.Server = ServerConfig{
		Host:            getEnvOrDefault("SERVER_HOST", "0.0.0.0"),
		Port:            getIntOrDefault("SERVER_PORT", 8080),
		BaseURL:         getEnvOrDefault("BASE_URL", "http://localhost:8080"),
		ReadTimeout:     getDurationOrDefault("SERVER_READ_TIMEOUT", 30*time.Second),
		WriteTimeout:    getDurationOrDefault("SERVER_WRITE_TIMEOUT", 30*time.Second),
		ShutdownTimeout: getDurationOrDefault("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
		AllowedOrigins:  getSliceOrDefault("ALLOWED_ORIGINS", []string{"http://localhost:8080"}),
	}

	// Load Casbin config
	cfg.Casbin = CasbinConfig{
		ModelPath:  getEnvOrDefault("CASBIN_MODEL", CASBIN_MODEL_FILE),
		PolicyPath: getEnvOrDefault("CASBIN_POLICY", CASBIN_POLICY_FILE),
	}

	return cfg, nil
}

// MustLoadConfig loads configuration and panics on error
func MustLoadConfig() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}
	return cfg
}

// LoadConfigWithDefaults loads config with sensible defaults for development
// This should NOT be used in production
func LoadConfigWithDefaults() *Config {
	return &Config{
		Environment: "development",
		JWT: JWTConfig{
			Secret:             getEnvOrDefault("JWT_SECRET", "development-secret-key-do-not-use-in-production-min-32-chars"),
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 7 * 24 * time.Hour,
			Issuer:             "azf-dev",
		},
		Database: DBConfig{
			Driver:          "sqlite",
			DSN:             "tmp/AZF_auth_z.db",
			MaxIdleConns:    10,
			MaxOpenConns:    100,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: 30 * time.Minute,
		},
		RateLimit: RateLimitCfg{
			Enabled:           true,
			RequestsPerSecond: 10.0,
			Burst:             20,
		},
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			BaseURL:         "http://localhost:8080",
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			AllowedOrigins:  []string{"*"},
		},
		Casbin: CasbinConfig{
			ModelPath:  CASBIN_MODEL_FILE,
			PolicyPath: CASBIN_POLICY_FILE,
		},
	}
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return ErrJWTSecretRequired
	}
	if len(c.JWT.Secret) < 32 {
		return ErrJWTSecretTooShort
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return ErrInvalidPort
	}
	return nil
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getFloatOrDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getSliceOrDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		var result []string
		for _, v := range splitAndTrim(value, ",") {
			if v != "" {
				result = append(result, v)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

func splitAndTrim(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, trim(s[start:i]))
			start = i + len(sep)
		}
	}
	result = append(result, trim(s[start:]))
	return result
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
