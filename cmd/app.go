package cmd

import (
	"context"
	"fmt"

	"salesforce-splunk-migration/internal/workflows"
	"salesforce-splunk-migration/services"
	"salesforce-splunk-migration/utils"
)

// Execute runs the main migration workflow using FlowGraph orchestration
func Execute() error {
	// Load configuration
	config, err := utils.LoadConfig("credentials.json")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	fmt.Println("✅ Configuration loaded and validated")

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
		fmt.Printf("\n⚠️  Migration completed with errors: %d/%d inputs failed\n", failed, success+failed)
	}

	return nil
}
