package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"salesforce-splunk-migration/internal/workflows"
	"salesforce-splunk-migration/services"
	"salesforce-splunk-migration/utils"
)

func Execute() error {
	logger := utils.GetLogger()

	config, err := utils.LoadConfig(os.Getenv("VAULT_PATH"))
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	logger.Info("✅ Configuration loaded and validated")

	splunkService, err := services.NewSplunkService(config)
	if err != nil {
		return fmt.Errorf("failed to create Splunk service: %w", err)
	}

	dashboardService, err := services.NewDashboardService(config, splunkService)
	if err != nil {
		return fmt.Errorf("failed to create Dashboard service: %w", err)
	}

	migrationGraph, err := workflows.NewMigrationGraph(config, splunkService, dashboardService)
	if err != nil {
		return fmt.Errorf("failed to create migration graph: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := migrationGraph.Execute(ctx); err != nil {
		return fmt.Errorf("migration workflow failed: %w", err)
	}

	state := migrationGraph.GetState()
	success, failed := state.GetCounters()

	if failed > 0 {
		logger.Warn("Migration completed with errors",
			utils.Int("failed", failed),
			utils.Int("total", success+failed))
	}

	return nil
}
