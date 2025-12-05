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

// MigrationNodeProcessor is a custom FlowGraph node processor for migration steps
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
}

// NewMigrationNodeProcessor creates a new migration node processor
func NewMigrationNodeProcessor(config *utils.Config, splunkService *services.SplunkService) *MigrationNodeProcessor {
	executionID := state.GenerateExecutionID()
	return &MigrationNodeProcessor{
		config:        config,
		splunkService: splunkService,
		stateManager:  state.NewMigrationStateManager(executionID),
		failedInputs:  make([]string, 0),
	}
}

// Process executes a migration node based on its ID
func (p *MigrationNodeProcessor) Process(ctx context.Context, node *flowgraph.Node, input map[string]interface{}) (map[string]interface{}, error) {
	output := make(map[string]interface{})

	// Copy input to output
	for k, v := range input {
		output[k] = v
	}

	// Execute the appropriate migration step based on node ID
	var err error
	switch node.ID {
	case "authenticate":
		err = p.authenticateStep(ctx)
	case "create_index":
		err = p.createIndexStep(ctx)
	case "create_account":
		err = p.createAccountStep(ctx)
	case "load_data_inputs":
		err = p.loadDataInputsStep(ctx)
	case "create_data_inputs":
		err = p.createDataInputsStep(ctx)
	case "verify_inputs":
		err = p.verifyInputsStep(ctx)
	default:
		return nil, fmt.Errorf("unknown migration step: %s", node.ID)
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

// authenticateStep handles authentication with Splunk
func (p *MigrationNodeProcessor) authenticateStep(ctx context.Context) error {
	fmt.Println("\n🔐 Step 1: Authenticating with Splunk...")
	p.stateManager.StartStep("authenticate")

	if err := p.splunkService.Authenticate(); err != nil {
		p.stateManager.FailStep("authenticate", err)
		return fmt.Errorf("authentication failed: %w", err)
	}

	p.stateManager.CompleteStep("authenticate")
	fmt.Println("✅ Authentication successful")
	return nil
}

// createIndexStep handles index creation
func (p *MigrationNodeProcessor) createIndexStep(ctx context.Context) error {
	fmt.Println("\n📊 Step 2: Creating Splunk index...")
	p.stateManager.StartStep("create_index")

	if err := p.splunkService.CreateIndex(p.config.Splunk.IndexName); err != nil {
		p.stateManager.FailStep("create_index", err)
		return fmt.Errorf("failed to create index %s: %w", p.config.Splunk.IndexName, err)
	}

	p.stateManager.CompleteStep("create_index")
	fmt.Printf("✅ Index created: %s\n", p.config.Splunk.IndexName)
	return nil
}

// createAccountStep handles Salesforce account creation
func (p *MigrationNodeProcessor) createAccountStep(ctx context.Context) error {
	fmt.Println("\n🔗 Step 3: Creating Salesforce account in Splunk...")
	p.stateManager.StartStep("create_account")

	if err := p.splunkService.CreateSalesforceAccount(); err != nil {
		p.stateManager.FailStep("create_account", err)
		return fmt.Errorf("failed to create Salesforce account: %w", err)
	}

	p.stateManager.CompleteStep("create_account")
	fmt.Println("✅ Salesforce account created")
	return nil
}

// loadDataInputsStep loads data inputs from configuration
func (p *MigrationNodeProcessor) loadDataInputsStep(ctx context.Context) error {
	p.stateManager.StartStep("load_data_inputs")

	dataInputs, err := p.config.GetDataInputs()
	if err != nil {
		p.stateManager.FailStep("load_data_inputs", err)
		return fmt.Errorf("failed to load data inputs: %w", err)
	}

	p.dataInputs = dataInputs
	p.inputProgress = state.NewDataInputProgress(dataInputs)
	p.stateManager.SetStepMetadata("load_data_inputs", "total_inputs", len(dataInputs))
	p.stateManager.CompleteStep("load_data_inputs")

	fmt.Printf("\n📥 Step 4: Loaded %d data inputs for creation\n", len(dataInputs))
	return nil
}

// createDataInputsStep creates data inputs in parallel
func (p *MigrationNodeProcessor) createDataInputsStep(ctx context.Context) error {
	if len(p.dataInputs) == 0 {
		fmt.Println("⚠️  No data inputs configured. Skipping...")
		return nil
	}

	p.stateManager.StartStep("create_data_inputs")

	maxParallelism := p.config.Migration.ConcurrentRequests
	fmt.Printf("\n🔄 Step 5: Creating %d data inputs in parallel (max %d workers)...\n",
		len(p.dataInputs), maxParallelism)

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

			fmt.Printf("\n[%d/%d] Processing data input: %s (Object: %s)\n",
				idx+1, len(p.dataInputs), inp.Name, inp.Object)

			if err := p.splunkService.CreateDataInput(&inp); err != nil {
				fmt.Printf("  ❌ Failed: %v\n", err)
				p.incrementFailed(inp.Name)
				p.inputProgress.FailInput(inp.Name, err)
			} else {
				fmt.Printf("  ✅ Created successfully\n")
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

	fmt.Printf("\n⏱️  Parallel execution completed in %v\n", duration.Round(time.Millisecond))
	fmt.Printf("📊 Summary: %d/%d data inputs created successfully", success, len(p.dataInputs))

	if failed > 0 {
		fmt.Printf(", %d failed\n", failed)
		fmt.Printf("❌ Failed inputs: %v\n", p.failedInputs)
		return fmt.Errorf("%d data inputs failed to create", failed)
	}
	fmt.Println()

	return nil
}

// verifyInputsStep verifies created data inputs
func (p *MigrationNodeProcessor) verifyInputsStep(ctx context.Context) error {
	fmt.Println("\n🔍 Step 6: Verifying created data inputs...")
	p.stateManager.StartStep("verify_inputs")

	existingInputs, err := p.splunkService.ListDataInputs()
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not list data inputs: %v\n", err)
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
		fmt.Printf("✅ Verified %d data inputs created by this code:\n", len(ourInputs))
		for _, name := range ourInputs {
			fmt.Printf("  - %s\n", name)
		}
	} else {
		fmt.Println("⚠️  Warning: None of the configured data inputs were found in Splunk")
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
