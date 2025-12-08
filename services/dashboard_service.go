package services

import (
	"context"
	"fmt"

	"salesforce-splunk-migration/utils"
	dashboardModels "splunk-dashboards/models"
	dashboardPkg "splunk-dashboards/utils/dashboard"
)

// DashboardServiceInterface defines the interface for dashboard operations
type DashboardServiceInterface interface {
	CreateDashboardsFromDirectory(ctx context.Context, dashboardDir string) error
}

// DashboardService handles dashboard creation operations
type DashboardService struct {
	config           *utils.Config
	splunkService    SplunkServiceInterface
	dashboardManager *dashboardPkg.DashboardManager
	logger           utils.Logger
}

// NewDashboardService creates a new dashboard service instance
func NewDashboardService(config *utils.Config, splunkService SplunkServiceInterface) (*DashboardService, error) {
	logger := utils.GetLogger()

	// Get authentication token from splunk service
	authToken := splunkService.GetAuthToken()

	// Convert config to dashboard package format with token
	dashboardConfig := dashboardModels.NewAppConfigFromCreds(
		config.Splunk.URL,
		config.Splunk.Username,
		config.Splunk.Password,
		config.Splunk.SkipSSLVerify,
		config.Migration.DashboardDirectory,
	)
	dashboardConfig.SplunkToken = authToken

	// Create dashboard manager
	dashboardManager, err := dashboardPkg.NewDashboardManager(dashboardConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dashboard manager: %w", err)
	}

	return &DashboardService{
		config:           config,
		splunkService:    splunkService,
		dashboardManager: dashboardManager,
		logger:           logger,
	}, nil
}

// CreateDashboardsFromDirectory creates all dashboards from XML files in the configured directory
func (ds *DashboardService) CreateDashboardsFromDirectory(ctx context.Context, dashboardDir string) error {
	ds.logger.Info("Creating dashboards from directory",
		utils.String("directory", dashboardDir))

	// Update the dashboard manager with the current auth token
	// (in case it was refreshed after initial service creation)
	ds.dashboardManager.UpdateAuthToken(ds.splunkService.GetAuthToken())

	err := ds.dashboardManager.CreateDashboardsFromDirectory(ctx, dashboardDir)
	if err != nil {
		ds.logger.Error("Failed to create dashboards",
			utils.String("directory", dashboardDir),
			utils.Err(err))
		return err
	}

	ds.logger.Info("✅ All dashboards created successfully",
		utils.String("directory", dashboardDir))
	return nil
}
