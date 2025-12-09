// Package workflows provides FlowGraph integration for migration
package workflows

import (
	"context"
	"fmt"
	"sync"
	"time"

	"salesforce-splunk-migration/services"
	"salesforce-splunk-migration/utils"

	"github.com/flowgraph/flowgraph/pkg/flowgraph"
)

// MigrationNodeProcessor is a custom FlowGraph node processor for migration nodes
type MigrationNodeProcessor struct {
	config           *utils.Config
	splunkService    services.SplunkServiceInterface
	dashboardService services.DashboardServiceInterface
	dataInputs       []utils.DataInput
	successCount     int
	failedCount      int
	failedInputs     []string
	mu               sync.RWMutex
	logger           utils.Logger
}

// NewMigrationNodeProcessor creates a new migration node processor
func NewMigrationNodeProcessor(config *utils.Config, splunkService services.SplunkServiceInterface, dashboardService services.DashboardServiceInterface) *MigrationNodeProcessor {
	return &MigrationNodeProcessor{
		config:           config,
		splunkService:    splunkService,
		dashboardService: dashboardService,
		failedInputs:     make([]string, 0),
		logger:           utils.GetLogger(),
	}
}

// Process executes a migration node based on its ID
func (p *MigrationNodeProcessor) Process(ctx context.Context, node *flowgraph.Node, input map[string]interface{}) (map[string]interface{}, error) {
	output := make(map[string]interface{})

	// Copy input to output
	for k, v := range input {
		output[k] = v
	}

	// Execute the appropriate migration node based on node ID
	var err error
	switch node.ID {
	case "authenticate":
		err = p.authenticateNode(ctx)
	case "check_salesforce_addon":
		err = p.checkSalesforceAddonNode(ctx)
	case "create_index":
		err = p.createIndexNode(ctx)
	case "create_account":
		err = p.createAccountNode(ctx)
	case "load_data_inputs":
		err = p.loadDataInputsNode(ctx)
	case "create_data_inputs":
		err = p.createDataInputsNode(ctx)
	case "verify_inputs":
		err = p.verifyInputsNode(ctx)
	case "create_dashboards":
		err = p.createDashboardsNode(ctx)
	default:
		p.logger.Error("Unknown migration node", utils.String("node_id", node.ID))
		return nil, fmt.Errorf("unknown migration node: %s", node.ID)
	}

	if err != nil {
		return nil, err
	}

	// Add step completion marker
	output["last_completed_step"] = node.ID
	output["timestamp"] = time.Now().Format(time.RFC3339)

	return output, nil
}

// CanProcess returns true if this processor can handle the given node type
func (p *MigrationNodeProcessor) CanProcess(nodeType flowgraph.NodeType) bool {
	// We handle all function nodes for migration
	return nodeType == flowgraph.NodeTypeFunction
}

// authenticateNode handles authentication with Splunk
func (p *MigrationNodeProcessor) authenticateNode(ctx context.Context) error {
	p.logger.Info("🔐 Node 1: Authenticating with Splunk...")

	if err := p.splunkService.Authenticate(ctx); err != nil {
		p.logger.Error("Authentication failed", utils.Err(err))
		return err
	}

	p.logger.Info("✅ Authentication successful")
	return nil
}

// checkSalesforceAddonNode checks if Splunk Add-on for Salesforce is installed
func (p *MigrationNodeProcessor) checkSalesforceAddonNode(ctx context.Context) error {
	p.logger.Info("🔌 Node 2: Checking Splunk Add-on for Salesforce...")

	if err := p.splunkService.CheckSalesforceAddon(ctx); err != nil {
		p.logger.Error("Splunk Add-on for Salesforce check failed", utils.Err(err))
		return err
	}

	p.logger.Info("✅ Splunk Add-on for Salesforce is installed and enabled")
	return nil
}

// createIndexNode handles index creation
func (p *MigrationNodeProcessor) createIndexNode(ctx context.Context) error {
	p.logger.Info("📊 Node 3: Verifying Splunk index...")

	// BYPASSED: Index creation skipped for Splunk Cloud
	// Splunk Cloud requires indexes to be created manually via UI or support ticket
	// The REST API cannot create indexes on indexer clusters

	// Check if index exists
	exists, err := p.splunkService.CheckIndexExists(ctx, p.config.Splunk.IndexName)
	if err != nil {
		p.logger.Warn("Could not verify index exists",
			utils.String("index_name", p.config.Splunk.IndexName),
			utils.Err(err))
		// Continue anyway - index might exist but not be visible via this endpoint
	}

	if exists {
		p.logger.Info("✅ Index verified successfully", utils.String("index_name", p.config.Splunk.IndexName))
	} else {
		p.logger.Warn("⚠️  Index not found - must be created manually via Splunk Web UI",
			utils.String("index_name", p.config.Splunk.IndexName),
			utils.String("action", "Create index at Settings → Indexes → New Index"))
		// Don't fail - allow workflow to continue
	}

	return nil
}

// createAccountNode handles Salesforce account creation
func (p *MigrationNodeProcessor) createAccountNode(ctx context.Context) error {
	p.logger.Info("🔗 Node 4: Creating Salesforce account in Splunk...")

	// Check if account already exists
	exists, err := p.splunkService.CheckSalesforceAccountExists(ctx)
	if err != nil {
		// Only log warning for actual errors (404 is handled gracefully by CheckSalesforceAccountExists)
		p.logger.Warn("Could not check if Salesforce account exists, will attempt to create",
			utils.String("account", p.config.Salesforce.AccountName),
			utils.Err(err))
		exists = false
	}

	if exists {
		// Account exists, update it
		p.logger.Info("Salesforce account exists, updating...",
			utils.String("account", p.config.Salesforce.AccountName))

		if err := p.splunkService.UpdateSalesforceAccount(ctx); err != nil {
			p.logger.Error("Failed to update Salesforce account", utils.Err(err))
			return err
		}
		p.logger.Info("✅ Salesforce account updated successfully")
	} else {
		// Account doesn't exist, create it
		if err := p.splunkService.CreateSalesforceAccount(ctx); err != nil {
			p.logger.Error("Failed to create Salesforce account", utils.Err(err))
			return err
		}
		p.logger.Info("✅ Salesforce account created successfully")
	}

	return nil
}

// loadDataInputsNode loads data inputs from configuration
func (p *MigrationNodeProcessor) loadDataInputsNode(ctx context.Context) error {
	dataInputs, err := p.config.GetDataInputs()
	if err != nil {
		p.logger.Error("Failed to load data inputs", utils.Err(err))
		return err
	}

	p.dataInputs = dataInputs

	p.logger.Info("📥 Node 5: Loaded data inputs for creation",
		utils.Int("count", len(dataInputs)))
	return nil
}

// createDataInputsNode creates data inputs in parallel
func (p *MigrationNodeProcessor) createDataInputsNode(ctx context.Context) error {
	if len(p.dataInputs) == 0 {
		p.logger.Warn("⚠️  No data inputs configured. Skipping...")
		return nil
	}

	maxParallelism := p.config.Migration.ConcurrentRequests
	p.logger.Info("🔄 Node 6: Creating data inputs in parallel",
		utils.Int("count", len(p.dataInputs)),
		utils.Int("max_workers", maxParallelism))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxParallelism)
	startTime := time.Now()

	for i, input := range p.dataInputs {
		wg.Add(1)
		go func(idx int, inp utils.DataInput) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Check if data input already exists
			exists, err := p.splunkService.CheckDataInputExists(ctx, inp.Name)
			if err != nil {
				// Only log warning for actual errors (404 is handled gracefully by CheckDataInputExists)
				p.logger.Warn("Could not check if data input exists, will attempt to create",
					utils.String("name", inp.Name),
					utils.Err(err))
				exists = false
			}

			if exists {
				// Data input exists, update it
				p.logger.Info("Data input exists, updating...",
					utils.String("name", inp.Name),
					utils.String("object", inp.Object))

				if err := p.splunkService.UpdateDataInput(ctx, &inp); err != nil {
					p.logger.Error("Failed to update data input",
						utils.String("name", inp.Name),
						utils.String("object", inp.Object),
						utils.Err(err))
					p.incrementFailed(inp.Name)
				} else {
					p.logger.Info("Data input updated successfully",
						utils.String("name", inp.Name),
						utils.String("object", inp.Object))
					p.incrementSuccess()
				}
			} else {
				// Data input doesn't exist, create it
				if err := p.splunkService.CreateDataInput(ctx, &inp); err != nil {
					p.logger.Error("Failed to create data input",
						utils.String("name", inp.Name),
						utils.String("object", inp.Object),
						utils.Err(err))
					p.incrementFailed(inp.Name)
				} else {
					p.logger.Info("Data input created successfully",
						utils.String("name", inp.Name),
						utils.String("object", inp.Object))
					p.incrementSuccess()
				}
			}
		}(i, input)
	}

	wg.Wait()

	duration := time.Since(startTime)
	success, failed := p.getCounters()

	if failed > 0 {
		p.logger.Warn("❌ Data inputs creation completed with errors",
			utils.Int("success", success),
			utils.Int("failed", failed),
			utils.Duration("duration", duration))
		return fmt.Errorf("%d data inputs failed to create", failed)
	}

	p.logger.Info("✅ All data inputs created successfully",
		utils.Int("count", success),
		utils.Duration("duration", duration))

	return nil
}

// verifyInputsNode verifies created data inputs
func (p *MigrationNodeProcessor) verifyInputsNode(ctx context.Context) error {
	p.logger.Info("🔍 Node 7: Verifying created data inputs...")

	existingInputs, err := p.splunkService.ListDataInputs(ctx)
	if err != nil {
		p.logger.Warn("Could not list data inputs", utils.Err(err))
		return nil
	}

	createdInputNames := make(map[string]bool)
	for _, input := range p.dataInputs {
		createdInputNames[input.Name] = true
	}

	var ourInputs []string
	for _, name := range existingInputs {
		if createdInputNames[name] {
			ourInputs = append(ourInputs, name)
		}
	}

	if len(ourInputs) > 0 {
		p.logger.Info("✅ Verified data inputs created",
			utils.Int("count", len(ourInputs)))
		for _, name := range ourInputs {
			p.logger.Debug("Verified input", utils.String("name", name))
		}
	} else {
		p.logger.Warn("None of the configured data inputs were found in Splunk")
	}

	return nil
}

// createDashboardsNode creates dashboards from XML files
func (p *MigrationNodeProcessor) createDashboardsNode(ctx context.Context) error {
	dashboardDir := p.config.Migration.DashboardDirectory

	p.logger.Info("📊 Node 8: Creating Splunk dashboards...",
		utils.String("directory", dashboardDir))

	if dashboardDir == "" {
		p.logger.Warn("⚠️  Dashboard directory not configured. Skipping dashboard creation...")
		return nil
	}

	// Check if directory exists
	if exists, err := utils.FileExists(dashboardDir); err != nil || !exists {
		p.logger.Warn("⚠️  Dashboard directory not found. Skipping dashboard creation...",
			utils.String("directory", dashboardDir))
		return nil
	}

	if err := p.dashboardService.CreateDashboardsFromDirectory(ctx, dashboardDir); err != nil {
		p.logger.Error("Failed to create dashboards",
			utils.String("directory", dashboardDir),
			utils.Err(err))
		return err
	}

	p.logger.Info("✅ Dashboards created successfully")
	return nil
}

// Helper methods for thread-safe counter management
func (p *MigrationNodeProcessor) incrementSuccess() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.successCount++
}

func (p *MigrationNodeProcessor) incrementFailed(inputName string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failedCount++
	p.failedInputs = append(p.failedInputs, inputName)
}

func (p *MigrationNodeProcessor) getCounters() (success, failed int) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.successCount, p.failedCount
}

// GetCounters returns the final counters
func (p *MigrationNodeProcessor) GetCounters() (success, failed int) {
	return p.getCounters()
}
