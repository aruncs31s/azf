package routes

import (
	"github.com/aruncs31s/azf/infrastructure/enterprise"
	"github.com/aruncs31s/azf/shared/logger"
	"go.uber.org/zap"
)

// registerEnterpriseRouteMetadata registers all routes with enterprise authorization metadata
// Routes are loaded from enterprise_route_metadata.json configuration file
func registerEnterpriseRouteMetadata(setup *enterprise.EnterpriseAuthorizationSetup) {
	// Load routes from configuration file
	routes, err := enterprise.LoadEnterpriseRouteMetadata("")
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
