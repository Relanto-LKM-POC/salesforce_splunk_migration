package cmd

import (
	"context"
	"fmt"

	"salesforce-splunk-migration/internal/workflows"
	"salesforce-splunk-migration/services"
	"salesforce-splunk-migration/utils"
)

const VAULT_PATH = "credentials.json"

// Execute runs the main migration workflow using FlowGraph orchestration
func Execute() error {
	logger := utils.GetLogger()

	// Load configuration
	config, err := utils.LoadConfig(VAULT_PATH)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	logger.Info("✅ Configuration loaded and validated")

	// Create Splunk service
	splunkService, err := services.NewSplunkService(config)
	if err != nil {
		return fmt.Errorf("failed to create Splunk service: %w", err)
	}

	// Create FlowGraph-based migration workflow
	migrationGraph, err := workflows.NewMigrationGraph(config, splunkService)
	if err != nil {
		return fmt.Errorf("failed to create migration graph: %w", err)
	}

	// Execute the workflow with state management
	ctx := context.Background()
	if err := migrationGraph.Execute(ctx); err != nil {
		return fmt.Errorf("migration workflow failed: %w", err)
	}

	// Get final state for reporting
	state := migrationGraph.GetState()
	success, failed := state.GetCounters()

	if failed > 0 {
		logger.Warn("Migration completed with errors",
			utils.Int("failed", failed),
			utils.Int("total", success+failed))
	}

	return nil
}
