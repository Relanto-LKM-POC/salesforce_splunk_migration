package cmd_test

// import (
// 	"context"
// 	"os"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"

// 	"salesforce-splunk-migration/cmd"
// )

// func TestExecute(t *testing.T) {
// 	t.Run("Error_ConfigurationLoadFailed", func(t *testing.T) {
// 		// Set invalid vault path
// 		os.Setenv("VAULT_PATH", "/invalid/path/config.json")
// 		defer os.Unsetenv("VAULT_PATH")

// 		err := cmd.Execute()
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "failed to load configuration")
// 	})

// 	t.Run("Error_ConfigurationValidationFailed", func(t *testing.T) {
// 		// Create a temporary invalid config file
// 		tempFile, err := os.CreateTemp("", "invalid-config-*.json")
// 		require.NoError(t, err)
// 		defer os.Remove(tempFile.Name())

// 		// Write invalid config (missing required fields)
// 		_, err = tempFile.WriteString(`{
// 			"SPLUNK_URL": "https://splunk.example.com:8089"
// 		}`)
// 		require.NoError(t, err)
// 		tempFile.Close()

// 		os.Setenv("VAULT_PATH", tempFile.Name())
// 		defer os.Unsetenv("VAULT_PATH")

// 		err = cmd.Execute()
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "configuration validation failed")
// 	})
// }

// func TestExecute_Integration(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration test in short mode")
// 	}

// 	t.Run("Success_ValidConfiguration", func(t *testing.T) {
// 		// Create a temporary valid config file
// 		tempFile, err := os.CreateTemp("", "valid-config-*.json")
// 		require.NoError(t, err)
// 		defer os.Remove(tempFile.Name())

// 		// Write valid config
// 		_, err = tempFile.WriteString(`{
// 			"SPLUNK_URL": "https://splunk.example.com:8089",
// 			"SPLUNK_USERNAME": "admin",
// 			"SPLUNK_PASSWORD": "password",
// 			"SALESFORCE_ENDPOINT": "https://login.salesforce.com",
// 			"SALESFORCE_CLIENT_ID": "test_client_id",
// 			"SALESFORCE_CLIENT_SECRET": "test_client_secret",
// 			"SALESFORCE_ACCOUNT_NAME": "test_account",
// 			"SPLUNK_INDEX_NAME": "test_index",
// 			"MIGRATION_CONCURRENT_REQUESTS": "5"
// 		}`)
// 		require.NoError(t, err)
// 		tempFile.Close()

// 		os.Setenv("VAULT_PATH", tempFile.Name())
// 		defer os.Unsetenv("VAULT_PATH")

// 		// This will fail at service creation or workflow execution
// 		// but validates config loading and validation work
// 		err = cmd.Execute()
// 		// We expect an error because we can't connect to real Splunk
// 		// but it should not be a config error
// 		if err != nil {
// 			assert.NotContains(t, err.Error(), "failed to load configuration")
// 			assert.NotContains(t, err.Error(), "configuration validation failed")
// 		}
// 	})
// }

// func TestExecute_WithContext(t *testing.T) {
// 	t.Run("Error_ContextCancellation", func(t *testing.T) {
// 		// Create a temporary valid config file
// 		tempFile, err := os.CreateTemp("", "config-*.json")
// 		require.NoError(t, err)
// 		defer os.Remove(tempFile.Name())

// 		_, err = tempFile.WriteString(`{
// 			"SPLUNK_URL": "https://splunk.example.com:8089",
// 			"SPLUNK_USERNAME": "admin",
// 			"SPLUNK_PASSWORD": "password",
// 			"SALESFORCE_ENDPOINT": "https://login.salesforce.com",
// 			"SALESFORCE_CLIENT_ID": "test_client_id",
// 			"SALESFORCE_CLIENT_SECRET": "test_client_secret",
// 			"SALESFORCE_ACCOUNT_NAME": "test_account",
// 			"SPLUNK_INDEX_NAME": "test_index"
// 		}`)
// 		require.NoError(t, err)
// 		tempFile.Close()

// 		os.Setenv("VAULT_PATH", tempFile.Name())
// 		defer os.Unsetenv("VAULT_PATH")

// 		// Create a cancelled context
// 		ctx, cancel := context.WithCancel(context.Background())
// 		cancel()
// 		_ = ctx

// 		// Execute will create its own context, but this tests timeout handling
// 		err = cmd.Execute()
// 		// Will fail at some point, but not due to our cancelled context
// 		require.Error(t, err)
// 	})
// }

// func TestExecute_ConfigurationVariations(t *testing.T) {
// 	t.Run("Error_MissingVaultPath", func(t *testing.T) {
// 		os.Unsetenv("VAULT_PATH")

// 		err := cmd.Execute()
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "failed to load configuration")
// 	})

// 	t.Run("Error_EmptyConfigFile", func(t *testing.T) {
// 		tempFile, err := os.CreateTemp("", "empty-config-*.json")
// 		require.NoError(t, err)
// 		defer os.Remove(tempFile.Name())
// 		tempFile.Close()

// 		os.Setenv("VAULT_PATH", tempFile.Name())
// 		defer os.Unsetenv("VAULT_PATH")

// 		err = cmd.Execute()
// 		require.Error(t, err)
// 	})

// 	t.Run("Error_MalformedJSON", func(t *testing.T) {
// 		tempFile, err := os.CreateTemp("", "malformed-*.json")
// 		require.NoError(t, err)
// 		defer os.Remove(tempFile.Name())

// 		_, err = tempFile.WriteString(`{invalid json`)
// 		require.NoError(t, err)
// 		tempFile.Close()

// 		os.Setenv("VAULT_PATH", tempFile.Name())
// 		defer os.Unsetenv("VAULT_PATH")

// 		err = cmd.Execute()
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "failed to load configuration")
// 	})
// }

// func TestExecute_ServiceCreation(t *testing.T) {
// 	t.Run("Error_InvalidSplunkURL", func(t *testing.T) {
// 		tempFile, err := os.CreateTemp("", "invalid-url-*.json")
// 		require.NoError(t, err)
// 		defer os.Remove(tempFile.Name())

// 		_, err = tempFile.WriteString(`{
// 			"SPLUNK_URL": "not-a-valid-url",
// 			"SPLUNK_USERNAME": "admin",
// 			"SPLUNK_PASSWORD": "password",
// 			"SALESFORCE_ENDPOINT": "https://login.salesforce.com",
// 			"SALESFORCE_CLIENT_ID": "test_client_id",
// 			"SALESFORCE_CLIENT_SECRET": "test_client_secret",
// 			"SALESFORCE_ACCOUNT_NAME": "test_account",
// 			"SPLUNK_INDEX_NAME": "test_index"
// 		}`)
// 		require.NoError(t, err)
// 		tempFile.Close()

// 		os.Setenv("VAULT_PATH", tempFile.Name())
// 		defer os.Unsetenv("VAULT_PATH")

// 		err = cmd.Execute()
// 		// May fail at validation or later stages
// 		require.Error(t, err)
// 	})
// }

// func TestExecute_ErrorHandling(t *testing.T) {
// 	t.Run("Error_PropagatesFromWorkflow", func(t *testing.T) {
// 		tempFile, err := os.CreateTemp("", "config-*.json")
// 		require.NoError(t, err)
// 		defer os.Remove(tempFile.Name())

// 		_, err = tempFile.WriteString(`{
// 			"SPLUNK_URL": "https://unreachable.example.com:8089",
// 			"SPLUNK_USERNAME": "admin",
// 			"SPLUNK_PASSWORD": "password",
// 			"SALESFORCE_ENDPOINT": "https://login.salesforce.com",
// 			"SALESFORCE_CLIENT_ID": "test",
// 			"SALESFORCE_CLIENT_SECRET": "test",
// 			"SALESFORCE_ACCOUNT_NAME": "test",
// 			"SPLUNK_INDEX_NAME": "test"
// 		}`)
// 		require.NoError(t, err)
// 		tempFile.Close()

// 		os.Setenv("VAULT_PATH", tempFile.Name())
// 		defer os.Unsetenv("VAULT_PATH")

// 		err = cmd.Execute()
// 		require.Error(t, err)
// 		// Should propagate error from workflow execution
// 	})
// }

// func TestExecute_DataInputConfiguration(t *testing.T) {
// 	t.Run("Error_MissingDataInputs", func(t *testing.T) {
// 		tempFile, err := os.CreateTemp("", "no-inputs-*.json")
// 		require.NoError(t, err)
// 		defer os.Remove(tempFile.Name())

// 		_, err = tempFile.WriteString(`{
// 			"SPLUNK_URL": "https://splunk.example.com:8089",
// 			"SPLUNK_USERNAME": "admin",
// 			"SPLUNK_PASSWORD": "password",
// 			"SALESFORCE_ENDPOINT": "https://login.salesforce.com",
// 			"SALESFORCE_CLIENT_ID": "test",
// 			"SALESFORCE_CLIENT_SECRET": "test",
// 			"SALESFORCE_ACCOUNT_NAME": "test",
// 			"SPLUNK_INDEX_NAME": "test",
// 			"DATA_INPUTS": []
// 		}`)
// 		require.NoError(t, err)
// 		tempFile.Close()

// 		os.Setenv("VAULT_PATH", tempFile.Name())
// 		defer os.Unsetenv("VAULT_PATH")

// 		err = cmd.Execute()
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "configuration validation failed")
// 	})

// 	t.Run("Error_InvalidDataInput", func(t *testing.T) {
// 		tempFile, err := os.CreateTemp("", "invalid-input-*.json")
// 		require.NoError(t, err)
// 		defer os.Remove(tempFile.Name())

// 		_, err = tempFile.WriteString(`{
// 			"SPLUNK_URL": "https://splunk.example.com:8089",
// 			"SPLUNK_USERNAME": "admin",
// 			"SPLUNK_PASSWORD": "password",
// 			"SALESFORCE_ENDPOINT": "https://login.salesforce.com",
// 			"SALESFORCE_CLIENT_ID": "test",
// 			"SALESFORCE_CLIENT_SECRET": "test",
// 			"SALESFORCE_ACCOUNT_NAME": "test",
// 			"SPLUNK_INDEX_NAME": "test",
// 			"DATA_INPUTS": [
// 				{
// 					"name": "Test"
// 				}
// 			]
// 		}`)
// 		require.NoError(t, err)
// 		tempFile.Close()

// 		os.Setenv("VAULT_PATH", tempFile.Name())
// 		defer os.Unsetenv("VAULT_PATH")

// 		err = cmd.Execute()
// 		require.Error(t, err)
// 	})
// }

// func TestExecute_EnvironmentVariables(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping environment variable test in short mode")
// 	}

// 	t.Run("Success_ReadsFromEnvironment", func(t *testing.T) {
// 		// Set environment variables
// 		os.Setenv("SPLUNK_URL", "https://splunk.example.com:8089")
// 		os.Setenv("SPLUNK_USERNAME", "admin")
// 		os.Setenv("SPLUNK_PASSWORD", "password")
// 		defer func() {
// 			os.Unsetenv("SPLUNK_URL")
// 			os.Unsetenv("SPLUNK_USERNAME")
// 			os.Unsetenv("SPLUNK_PASSWORD")
// 		}()

// 		// Still needs a config file with DATA_INPUTS
// 		tempFile, err := os.CreateTemp("", "env-config-*.json")
// 		require.NoError(t, err)
// 		defer os.Remove(tempFile.Name())

// 		_, err = tempFile.WriteString(`{
// 			"SALESFORCE_ENDPOINT": "https://login.salesforce.com",
// 			"SALESFORCE_CLIENT_ID": "test",
// 			"SALESFORCE_CLIENT_SECRET": "test",
// 			"SALESFORCE_ACCOUNT_NAME": "test",
// 			"DATA_INPUTS": [
// 				{
// 					"name": "test",
// 					"object": "Account",
// 					"object_fields": "Id,Name"
// 				}
// 			]
// 		}`)
// 		require.NoError(t, err)
// 		tempFile.Close()

// 		os.Setenv("VAULT_PATH", tempFile.Name())
// 		defer os.Unsetenv("VAULT_PATH")

// 		err = cmd.Execute()
// 		// Will fail at service creation but env vars should be read
// 		if err != nil {
// 			assert.NotContains(t, err.Error(), "SPLUNK_URL is required")
// 		}
// 	})
// }
