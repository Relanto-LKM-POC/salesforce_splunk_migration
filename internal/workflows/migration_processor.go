// Package workflows provides FlowGraph integration for migration
package workflows

import (
	"context"
	"fmt"
	"sync"
	"time"

	"salesforce-splunk-migration/internal/state"
	"salesforce-splunk-migration/services"
	"salesforce-splunk-migration/utils"

	"github.com/flowgraph/flowgraph/pkg/flowgraph"
)

// MigrationNodeProcessor is a custom FlowGraph node processor for migration nodes
type MigrationNodeProcessor struct {
	config        *utils.Config
	splunkService *services.SplunkService
	stateManager  *state.MigrationStateManager
	dataInputs    []utils.DataInput
	inputProgress *state.DataInputProgress
	successCount  int
	failedCount   int
	failedInputs  []string
	mu            sync.RWMutex
	logger        utils.Logger
}

// NewMigrationNodeProcessor creates a new migration node processor
func NewMigrationNodeProcessor(config *utils.Config, splunkService *services.SplunkService) *MigrationNodeProcessor {
	executionID := state.GenerateExecutionID()
	return &MigrationNodeProcessor{
		config:        config,
		splunkService: splunkService,
		stateManager:  state.NewMigrationStateManager(executionID),
		failedInputs:  make([]string, 0),
		logger:        utils.GetLogger(),
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
	p.stateManager.StartStep("authenticate")

	if err := p.splunkService.Authenticate(); err != nil {
		p.stateManager.FailStep("authenticate", err)
		p.logger.Error("Authentication failed", utils.Err(err))
		return err
	}

	p.stateManager.CompleteStep("authenticate")
	p.logger.Info("✅ Authentication successful")
	return nil
}

// checkSalesforceAddonNode checks if Splunk Add-on for Salesforce is installed
func (p *MigrationNodeProcessor) checkSalesforceAddonNode(ctx context.Context) error {
	p.logger.Info("🔌 Node 2: Checking Splunk Add-on for Salesforce...")
	p.stateManager.StartStep("check_salesforce_addon")

	if err := p.splunkService.CheckSalesforceAddon(); err != nil {
		p.stateManager.FailStep("check_salesforce_addon", err)
		p.logger.Error("Splunk Add-on for Salesforce check failed", utils.Err(err))
		return err
	}

	p.stateManager.CompleteStep("check_salesforce_addon")
	p.logger.Info("✅ Splunk Add-on for Salesforce is installed and enabled")
	return nil
}

// createIndexNode handles index creation
func (p *MigrationNodeProcessor) createIndexNode(ctx context.Context) error {
	p.logger.Info("📊 Node 3: Creating Splunk index...")
	p.stateManager.StartStep("create_index")

	if err := p.splunkService.CreateIndex(p.config.Splunk.IndexName); err != nil {
		p.stateManager.FailStep("create_index", err)
		p.logger.Error("Failed to create index",
			utils.String("index_name", p.config.Splunk.IndexName),
			utils.Err(err))
		return err
	}

	p.stateManager.CompleteStep("create_index")
	p.logger.Info("✅ Index created", utils.String("index_name", p.config.Splunk.IndexName))
	return nil
}

// createAccountNode handles Salesforce account creation
func (p *MigrationNodeProcessor) createAccountNode(ctx context.Context) error {
	p.logger.Info("🔗 Node 4: Creating Salesforce account in Splunk...")
	p.stateManager.StartStep("create_account")

	if err := p.splunkService.CreateSalesforceAccount(); err != nil {
		p.stateManager.FailStep("create_account", err)
		p.logger.Error("Failed to create Salesforce account", utils.Err(err))
		return err
	}

	p.stateManager.CompleteStep("create_account")
	p.logger.Info("✅ Salesforce account created")
	return nil
}

// loadDataInputsNode loads data inputs from configuration
func (p *MigrationNodeProcessor) loadDataInputsNode(ctx context.Context) error {
	p.stateManager.StartStep("load_data_inputs")

	dataInputs, err := p.config.GetDataInputs()
	if err != nil {
		p.stateManager.FailStep("load_data_inputs", err)
		p.logger.Error("Failed to load data inputs", utils.Err(err))
		return err
	}

	p.dataInputs = dataInputs
	p.inputProgress = state.NewDataInputProgress(dataInputs)
	p.stateManager.SetStepMetadata("load_data_inputs", "total_inputs", len(dataInputs))
	p.stateManager.CompleteStep("load_data_inputs")

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

	p.stateManager.StartStep("create_data_inputs")

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

			p.inputProgress.StartInput(inp.Name)

			if err := p.splunkService.CreateDataInput(&inp); err != nil {
				p.logger.Error("Failed to create data input",
					utils.String("name", inp.Name),
					utils.String("object", inp.Object),
					utils.Err(err))
				p.incrementFailed(inp.Name)
				p.inputProgress.FailInput(inp.Name, err)
			} else {
				p.logger.Info("Data input created successfully",
					utils.String("name", inp.Name),
					utils.String("object", inp.Object))
				p.incrementSuccess()
				p.inputProgress.CompleteInput(inp.Name)
			}
		}(i, input)
	}

	wg.Wait()

	duration := time.Since(startTime)
	success, failed := p.getCounters()

	p.stateManager.SetStepMetadata("create_data_inputs", "success_count", success)
	p.stateManager.SetStepMetadata("create_data_inputs", "failed_count", failed)
	p.stateManager.SetStepMetadata("create_data_inputs", "duration", duration)
	p.stateManager.SetStepMetadata("create_data_inputs", "parallelism", maxParallelism)
	p.stateManager.CompleteStep("create_data_inputs")

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
	p.stateManager.StartStep("verify_inputs")

	existingInputs, err := p.splunkService.ListDataInputs()
	if err != nil {
		p.logger.Warn("Could not list data inputs", utils.Err(err))
		p.stateManager.CompleteStep("verify_inputs")
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

	p.stateManager.CompleteStep("verify_inputs")
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

// GetStateManager returns the state manager for reporting
func (p *MigrationNodeProcessor) GetStateManager() *state.MigrationStateManager {
	return p.stateManager
}

// GetCounters returns the final counters
func (p *MigrationNodeProcessor) GetCounters() (success, failed int) {
	return p.getCounters()
}
