// Package workflows provides FlowGraph integration for migration
package workflows_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"salesforce-splunk-migration/internal/workflows"
	"salesforce-splunk-migration/mocks"
	"salesforce-splunk-migration/utils"

	"github.com/flowgraph/flowgraph/pkg/flowgraph"
)

func TestMigrationNodeProcessor_Process(t *testing.T) {
	config := &utils.Config{
		Splunk: utils.SplunkConfig{
			IndexName: "test_index",
		},
		Salesforce: utils.SalesforceConfig{
			AccountName: "test_account",
		},
		Migration: utils.MigrationConfig{
			ConcurrentRequests: 5,
		},
		Extensions: map[string]interface{}{},
	}

	t.Run("Success_Authenticate", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{
			AuthenticateFunc: func(ctx context.Context) error {
				return nil
			},
		}

		processor := workflows.NewMigrationNodeProcessor(config, mockService)
		node := &flowgraph.Node{
			ID:   "authenticate",
			Name: "Authenticate with Splunk",
			Type: flowgraph.NodeTypeFunction,
		}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.Equal(t, 1, mockService.AuthenticateCalls)
		assert.Equal(t, "authenticate", output["last_completed_step"])
	})

	t.Run("Error_Authenticate", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{
			AuthenticateFunc: func(ctx context.Context) error {
				return fmt.Errorf("authentication failed")
			},
		}

		processor := workflows.NewMigrationNodeProcessor(config, mockService)
		node := &flowgraph.Node{ID: "authenticate"}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
		assert.Nil(t, output)
		assert.Equal(t, 1, mockService.AuthenticateCalls)
	})

	t.Run("Success_CheckSalesforceAddon", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{
			CheckSalesforceAddonFunc: func(ctx context.Context) error {
				return nil
			},
		}

		processor := workflows.NewMigrationNodeProcessor(config, mockService)
		node := &flowgraph.Node{
			ID:   "check_salesforce_addon",
			Name: "Check Salesforce Addon",
			Type: flowgraph.NodeTypeFunction,
		}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.Equal(t, 1, mockService.CheckSalesforceAddonCalls)
		assert.Equal(t, "check_salesforce_addon", output["last_completed_step"])
	})

	t.Run("Error_CheckSalesforceAddon", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{
			CheckSalesforceAddonFunc: func(ctx context.Context) error {
				return fmt.Errorf("addon not found")
			},
		}

		processor := workflows.NewMigrationNodeProcessor(config, mockService)
		node := &flowgraph.Node{ID: "check_salesforce_addon"}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "addon not found")
		assert.Nil(t, output)
		assert.Equal(t, 1, mockService.CheckSalesforceAddonCalls)
	})

	t.Run("Success_CreateIndex", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{
			CreateIndexFunc: func(ctx context.Context, indexName string) error {
				return nil
			},
		}

		processor := workflows.NewMigrationNodeProcessor(config, mockService)
		node := &flowgraph.Node{
			ID:   "create_index",
			Name: "Create Index",
			Type: flowgraph.NodeTypeFunction,
		}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.Equal(t, 1, mockService.CreateIndexCalls)
		assert.Equal(t, "create_index", output["last_completed_step"])
	})

	t.Run("Error_CreateIndex", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{
			CreateIndexFunc: func(ctx context.Context, indexName string) error {
				return fmt.Errorf("index creation failed")
			},
		}

		processor := workflows.NewMigrationNodeProcessor(config, mockService)
		node := &flowgraph.Node{ID: "create_index"}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "index creation failed")
		assert.Nil(t, output)
		assert.Equal(t, 1, mockService.CreateIndexCalls)
	})

	t.Run("Success_CreateAccount", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{
			CreateSalesforceAccountFunc: func(ctx context.Context) error {
				return nil
			},
		}

		processor := workflows.NewMigrationNodeProcessor(config, mockService)
		node := &flowgraph.Node{
			ID:   "create_account",
			Name: "Create Salesforce Account",
			Type: flowgraph.NodeTypeFunction,
		}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.Equal(t, 1, mockService.CreateSalesforceAccountCalls)
		assert.Equal(t, "create_account", output["last_completed_step"])
	})

	t.Run("Error_CreateAccount", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{
			CreateSalesforceAccountFunc: func(ctx context.Context) error {
				return fmt.Errorf("account creation failed")
			},
		}

		processor := workflows.NewMigrationNodeProcessor(config, mockService)
		node := &flowgraph.Node{ID: "create_account"}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "account creation failed")
		assert.Nil(t, output)
		assert.Equal(t, 1, mockService.CreateSalesforceAccountCalls)
	})

	t.Run("Success_LoadDataInputs", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{}
		configWithInputs := &utils.Config{
			Splunk: utils.SplunkConfig{
				IndexName: "test_index",
			},
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
			Migration: utils.MigrationConfig{
				ConcurrentRequests: 5,
			},
			Extensions: map[string]interface{}{
				"DATA_INPUTS": []interface{}{
					map[string]interface{}{"name": "test_input", "object": "Account", "index": "test_index"},
				},
			},
		}

		processor := workflows.NewMigrationNodeProcessor(configWithInputs, mockService)
		node := &flowgraph.Node{
			ID:   "load_data_inputs",
			Name: "Load Data Inputs",
			Type: flowgraph.NodeTypeFunction,
		}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.Equal(t, "load_data_inputs", output["last_completed_step"])
	})

	t.Run("Success_LoadDataInputs_NoInputs", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{}
		configNoInputs := &utils.Config{
			Splunk: utils.SplunkConfig{
				IndexName: "test_index",
			},
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
			Migration: utils.MigrationConfig{
				ConcurrentRequests: 5,
			},
			Extensions: map[string]interface{}{},
		}
		processor := workflows.NewMigrationNodeProcessor(configNoInputs, mockService)
		node := &flowgraph.Node{ID: "load_data_inputs"}

		_, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.Error(t, err) // Should error when DATA_INPUTS not found
		assert.Contains(t, err.Error(), "DATA_INPUTS not found")
	})

	t.Run("Success_CreateDataInputs_WithInputs", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{
			CreateDataInputFunc: func(ctx context.Context, input *utils.DataInput) error {
				return nil
			},
		}
		configWithInputs := &utils.Config{
			Splunk: utils.SplunkConfig{
				IndexName: "test_index",
			},
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
			Migration: utils.MigrationConfig{
				ConcurrentRequests: 5,
			},
			Extensions: map[string]interface{}{
				"DATA_INPUTS": []interface{}{
					map[string]interface{}{"name": "test_input", "object": "Account", "index": "test_index"},
				},
			},
		}

		processor := workflows.NewMigrationNodeProcessor(configWithInputs, mockService)
		// Need to load inputs first
		loadNode := &flowgraph.Node{ID: "load_data_inputs"}
		_, _ = processor.Process(context.Background(), loadNode, make(map[string]interface{}))

		node := &flowgraph.Node{
			ID:   "create_data_inputs",
			Name: "Create Data Inputs",
			Type: flowgraph.NodeTypeFunction,
		}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.Equal(t, 1, mockService.CreateDataInputCalls)
		assert.Equal(t, "create_data_inputs", output["last_completed_step"])
	})

	t.Run("Success_CreateDataInputs_NoInputs", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{}
		processor := workflows.NewMigrationNodeProcessor(config, mockService)
		node := &flowgraph.Node{ID: "create_data_inputs"}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.Equal(t, 0, mockService.CreateDataInputCalls)
	})

	t.Run("Error_CreateDataInputs", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{
			CreateDataInputFunc: func(ctx context.Context, input *utils.DataInput) error {
				return fmt.Errorf("failed to create input")
			},
		}
		configWithInputs := &utils.Config{
			Splunk: utils.SplunkConfig{
				IndexName: "test_index",
			},
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
			Migration: utils.MigrationConfig{
				ConcurrentRequests: 5,
			},
			Extensions: map[string]interface{}{
				"DATA_INPUTS": []interface{}{
					map[string]interface{}{"name": "test_input", "object": "Account", "index": "test_index"},
				},
			},
		}

		processor := workflows.NewMigrationNodeProcessor(configWithInputs, mockService)
		// Load inputs first
		loadNode := &flowgraph.Node{ID: "load_data_inputs"}
		_, _ = processor.Process(context.Background(), loadNode, make(map[string]interface{}))

		node := &flowgraph.Node{ID: "create_data_inputs"}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "data inputs failed")
		assert.Nil(t, output)
		assert.Equal(t, 1, mockService.CreateDataInputCalls)
	})

	t.Run("Success_VerifyInputs", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{
			ListDataInputsFunc: func(ctx context.Context) ([]string, error) {
				return []string{"test_input"}, nil
			},
		}
		configWithInputs := &utils.Config{
			Splunk: utils.SplunkConfig{
				IndexName: "test_index",
			},
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
			Migration: utils.MigrationConfig{
				ConcurrentRequests: 5,
			},
			Extensions: map[string]interface{}{
				"DATA_INPUTS": []interface{}{
					map[string]interface{}{"name": "test_input", "object": "Account", "index": "test_index"},
				},
			},
		}

		processor := workflows.NewMigrationNodeProcessor(configWithInputs, mockService)
		// Load inputs first
		loadInputsNode := &flowgraph.Node{ID: "load_data_inputs"}
		_, _ = processor.Process(context.Background(), loadInputsNode, make(map[string]interface{}))

		node := &flowgraph.Node{
			ID:   "verify_inputs",
			Name: "Verify Data Inputs",
			Type: flowgraph.NodeTypeFunction,
		}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.Equal(t, 1, mockService.ListDataInputsCalls)
		assert.Equal(t, "verify_inputs", output["last_completed_step"])
	})

	t.Run("Error_VerifyInputs", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{
			ListDataInputsFunc: func(ctx context.Context) ([]string, error) {
				return nil, fmt.Errorf("failed to list inputs")
			},
		}
		configWithInputs := &utils.Config{
			Splunk: utils.SplunkConfig{
				IndexName: "test_index",
			},
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
			Migration: utils.MigrationConfig{
				ConcurrentRequests: 5,
			},
			Extensions: map[string]interface{}{
				"DATA_INPUTS": []interface{}{
					map[string]interface{}{"name": "test_input", "object": "Account", "index": "test_index"},
				},
			},
		}

		processor := workflows.NewMigrationNodeProcessor(configWithInputs, mockService)
		// Load inputs first
		loadInputsNode2 := &flowgraph.Node{ID: "load_data_inputs"}
		_, _ = processor.Process(context.Background(), loadInputsNode2, make(map[string]interface{}))

		node := &flowgraph.Node{ID: "verify_inputs"}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.NoError(t, err) // Verification failure is logged but doesn't error
		require.NotNil(t, output)
		assert.Equal(t, 1, mockService.ListDataInputsCalls)
	})

	t.Run("Error_UnknownNode", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{}
		processor := workflows.NewMigrationNodeProcessor(config, mockService)
		node := &flowgraph.Node{ID: "unknown_node"}

		output, err := processor.Process(context.Background(), node, make(map[string]interface{}))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown migration node")
		assert.Nil(t, output)
	})

	t.Run("Success_InputPassthrough", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{
			AuthenticateFunc: func(ctx context.Context) error {
				return nil
			},
		}
		processor := workflows.NewMigrationNodeProcessor(config, mockService)
		node := &flowgraph.Node{ID: "authenticate"}

		input := map[string]interface{}{
			"test_key": "test_value",
		}

		output, err := processor.Process(context.Background(), node, input)

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.Equal(t, "test_value", output["test_key"])
	})
}

func TestNewMigrationNodeProcessor(t *testing.T) {
	t.Run("Success_CreatesProcessor", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{}
		config := &utils.Config{
			Splunk: utils.SplunkConfig{
				IndexName: "test_index",
			},
		}

		processor := workflows.NewMigrationNodeProcessor(config, mockService)

		require.NotNil(t, processor)
	})

	t.Run("Success_NilConfig", func(t *testing.T) {
		mockService := &mocks.MockSplunkService{}

		processor := workflows.NewMigrationNodeProcessor(nil, mockService)

		require.NotNil(t, processor)
	})

	t.Run("Success_NilService", func(t *testing.T) {
		config := &utils.Config{}

		processor := workflows.NewMigrationNodeProcessor(config, nil)

		require.NotNil(t, processor)
	})
}
