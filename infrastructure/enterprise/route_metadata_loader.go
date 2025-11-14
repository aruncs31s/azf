package enterprise

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aruncs31s/azf/config"
	"github.com/aruncs31s/azf/shared/logger"
	"go.uber.org/zap"
)

// EnterpriseRouteMetadataConfig holds the structure of the enterprise route metadata config
type EnterpriseRouteMetadataConfig struct {
	Routes []*RouteMetadata `json:"routes"`
}

func LoadEnterpriseRouteMetadata(configPath string) ([]*RouteMetadata, error) {
	// Default to enterprise_route_metadata.json if not specified
	//TODO: Move to config class
	if configPath == "" {
		configPath = filepath.Join("application", "routes", "enterprise_route_metadata.json")
	}

	// Check if environment variable overrides the config path
	if envPath := os.Getenv("ENTERPRISE_ROUTE_METADATA_PATH"); envPath != "" {
		configPath = envPath
	}

	// Read the configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.Warn("Enterprise route metadata config file not found, using compatibility mode",
			zap.String("path", configPath),
			zap.Error(err))
		return nil, err
	}

	// Parse the JSON
	var config EnterpriseRouteMetadataConfig
	if err := json.Unmarshal(data, &config); err != nil {
		logger.Error("Failed to parse enterprise route metadata config",
			zap.String("path", configPath),
			zap.Error(err))
		return nil, err
	}

	logger.Info("Loaded enterprise route metadata configuration",
		zap.String("path", configPath),
		zap.Int("routes", len(config.Routes)))

	return config.Routes, nil
}

// SaveEnterpriseRouteMetadata saves the route metadata back to the JSON file
func SaveEnterpriseRouteMetadata(routes []*RouteMetadata, configPath string) error {
	// Default to enterprise_route_metadata.json if not specified
	if configPath == "" {
		configPath = filepath.Join("application", "routes", "enterprise_route_metadata.json")
	}

	// Check if environment variable overrides the config path
	if envPath := os.Getenv("ENTERPRISE_ROUTE_METADATA_PATH"); envPath != "" {
		configPath = envPath
	}

	// Create config structure
	config := EnterpriseRouteMetadataConfig{
		Routes: routes,
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal route metadata to JSON",
			zap.Error(err))
		return err
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		logger.Error("Failed to write route metadata to file",
			zap.String("path", configPath),
			zap.Error(err))
		return err
	}

	logger.Info("Saved enterprise route metadata configuration",
		zap.String("path", configPath),
		zap.Int("routes", len(routes)))

	return nil
}

// UpdateCasbinPoliciesFromRoutes updates the Casbin policy CSV file based on route metadata
func UpdateCasbinPoliciesFromRoutes(routes []*RouteMetadata, policyFilePath string) error {
	if policyFilePath == "" {
		policyFilePath = config.CASBIN_POLICY_DEFAULT_PATH
	}

	// Read existing policy file to preserve role assignments (g lines)
	existingContent, err := os.ReadFile(policyFilePath)
	existingPolicies := ""
	if err == nil {
		lines := strings.Split(string(existingContent), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "g,") {
				// Keep role assignments
				existingPolicies += line + "\n"
			}
		}
	}

	// Generate new policies from routes
	newPolicies := ""
	for _, route := range routes {
		if !route.IsPublic && len(route.AllowedRoles) > 0 {
			for _, role := range route.AllowedRoles {
				// Create policy: p, role, path, method
				policy := fmt.Sprintf("p, %s, %s, %s\n", role, route.Path, route.Method)
				newPolicies += policy
			}
		}
	}

	// Combine existing role assignments with new policies
	fullContent := existingPolicies + newPolicies

	// Write back to file
	if err := os.WriteFile(policyFilePath, []byte(fullContent), 0644); err != nil {
		logger.Error("Failed to update Casbin policy file",
			zap.String("path", policyFilePath),
			zap.Error(err))
		return err
	}

	logger.Info("Updated Casbin policies from route metadata",
		zap.String("path", policyFilePath),
		zap.Int("routes_processed", len(routes)))

	return nil
}

// registerEnterpriseRouteMetadata registers all routes with enterprise authorization metadata
// Routes are loaded from enterprise_route_metadata.json configuration file
func RegisterEnterpriseRouteMetadata(setup *EnterpriseAuthorizationSetup) {
	// Load routes from configuration file
	routes, err := LoadEnterpriseRouteMetadata("")
	if err != nil {
		logger.Warn("Failed to load enterprise route metadata from config, enterprise authorization may be limited",
			zap.Error(err))
		return
	}

	// Register all routes
	if err := setup.RegisterRoutes(routes...); err != nil {
		logger.Error("Failed to register enterprise route metadata", zap.Error(err))
	}

	// Validate policies
	if report := setup.ValidateAllPolicies(); !report.IsValid {
		logger.Warn("Policy validation issues found",
			zap.Int("errors", len(report.Errors)),
			zap.Int("warnings", len(report.Warnings)),
			zap.Float64("coverage_percentage", report.SummaryStats.CoveragePercentage))
	} else {
		logger.Info("Policy validation passed",
			zap.Float64("coverage_percentage", report.SummaryStats.CoveragePercentage))
	}
}
