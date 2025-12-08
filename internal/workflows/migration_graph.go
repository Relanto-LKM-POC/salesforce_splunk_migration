// Package workflows defines FlowGraph-based migration workflows
package workflows

import (
	"context"
	"fmt"
	"time"

	"salesforce-splunk-migration/services"
	"salesforce-splunk-migration/utils"

	"github.com/flowgraph/flowgraph/pkg/flowgraph"
)

// MigrationGraph represents the complete migration workflow using FlowGraph
type MigrationGraph struct {
	runtime   *flowgraph.Runtime
	graph     *flowgraph.Graph
	processor *MigrationNodeProcessor
	startTime time.Time
	endTime   time.Time
	logger    utils.Logger
}

// NewMigrationGraph creates a new FlowGraph-based migration workflow
func NewMigrationGraph(config *utils.Config, splunkService services.SplunkServiceInterface, dashboardService services.DashboardServiceInterface) (*MigrationGraph, error) {
	// Create custom node processor for migration nodes
	processor := NewMigrationNodeProcessor(config, splunkService, dashboardService)

	// Create FlowGraph runtime with custom processor
	runtime := flowgraph.NewRuntimeWithNodeProcessor(processor)

	// Build the migration graph
	migrationGraph, err := buildMigrationGraph()
	if err != nil {
		return nil, err
	}

	// Save graph to runtime
	ctx := context.Background()
	if err := runtime.SaveGraph(ctx, migrationGraph); err != nil {
		return nil, err
	}

	return &MigrationGraph{
		runtime:   runtime,
		graph:     migrationGraph,
		processor: processor,
		logger:    utils.GetLogger(),
	}, nil
}

// buildMigrationGraph constructs the FlowGraph structure for migration
func buildMigrationGraph() (*flowgraph.Graph, error) {
	g := &flowgraph.Graph{
		ID:         "salesforce-splunk-migration",
		Name:       "Salesforce to Splunk Migration",
		EntryPoint: "authenticate",
		Nodes:      make(map[string]*flowgraph.Node),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Define migration nodes
	nodes := []*flowgraph.Node{
		{
			ID:        "authenticate",
			Name:      "Authenticate with Splunk",
			Type:      flowgraph.NodeTypeFunction,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "check_salesforce_addon",
			Name:      "Check Splunk Add-on for Salesforce",
			Type:      flowgraph.NodeTypeFunction,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "create_index",
			Name:      "Create Splunk Index",
			Type:      flowgraph.NodeTypeFunction,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "create_account",
			Name:      "Create Salesforce Account",
			Type:      flowgraph.NodeTypeFunction,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "load_data_inputs",
			Name:      "Load Data Inputs",
			Type:      flowgraph.NodeTypeFunction,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "create_data_inputs",
			Name:      "Create Data Inputs",
			Type:      flowgraph.NodeTypeFunction,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "verify_inputs",
			Name:      "Verify Data Inputs",
			Type:      flowgraph.NodeTypeFunction,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "create_dashboards",
			Name:      "Create Dashboards",
			Type:      flowgraph.NodeTypeFunction,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Add nodes to graph
	for _, node := range nodes {
		if err := g.AddNode(node); err != nil {
			return nil, err
		}
	}

	// Define edges (sequential workflow)
	edges := []*flowgraph.Edge{
		{Source: "authenticate", Target: "check_salesforce_addon"},
		{Source: "check_salesforce_addon", Target: "create_index"},
		{Source: "create_index", Target: "create_account"},
		{Source: "create_account", Target: "load_data_inputs"},
		{Source: "load_data_inputs", Target: "create_data_inputs"},
		{Source: "create_data_inputs", Target: "verify_inputs"},
		{Source: "verify_inputs", Target: "create_dashboards"},
	}

	// Add edges to graph
	for _, edge := range edges {
		if err := g.AddEdge(edge); err != nil {
			return nil, err
		}
	}

	return g, nil
}

// Execute runs the migration workflow using FlowGraph
func (mg *MigrationGraph) Execute(ctx context.Context) error {
	mg.logger.Info("🚀 Starting Salesforce to Splunk Migration with FlowGraph...")

	mg.startTime = time.Now()

	defer func() {
		if r := recover(); r != nil {
			mg.logger.Error("Panic occurred during migration", utils.String("panic", fmt.Sprintf("%v", r)))
			panic(r)
		}
	}()

	// Execute the graph using FlowGraph runtime
	req := &flowgraph.ExecutionRequest{
		GraphID:  mg.graph.ID,
		ThreadID: "migration-thread-1",
		Input:    make(map[string]interface{}),
		Config: flowgraph.ExecutionConfig{
			MaxSteps:        100,
			Timeout:         30 * time.Minute,
			CheckpointEvery: 1,
			ValidateGraph:   true,
		},
	}

	response, err := mg.runtime.Execute(ctx, req)
	if err != nil {
		mg.logger.Error("Migration execution failed", utils.Err(err))
		return err
	}

	mg.endTime = time.Now()
	duration := mg.endTime.Sub(mg.startTime)

	if response.Status == "completed" {
		mg.logger.Info("🎉 Migration completed successfully",
			utils.Duration("total_time", duration))
	} else {
		mg.logger.Warn("Migration ended with non-completed status",
			utils.String("status", string(response.Status)),
			utils.Duration("total_time", duration))
	}

	return nil
}

// GetState returns counters for backwards compatibility
func (mg *MigrationGraph) GetState() *MigrationState {
	success, failed := mg.processor.GetCounters()
	return &MigrationState{
		SuccessCount: success,
		FailedCount:  failed,
	}
}

// MigrationState provides backwards compatibility for counter access
type MigrationState struct {
	SuccessCount int
	FailedCount  int
}

// GetCounters returns success and failed counts
func (ms *MigrationState) GetCounters() (success, failed int) {
	return ms.SuccessCount, ms.FailedCount
}
