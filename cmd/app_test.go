package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"salesforce-splunk-migration/utils"
)

func TestExecute(t *testing.T) {
	// Initialize logger for tests
	err := utils.InitializeGlobalLogger("test", "cmd", false)
	require.NoError(t, err)

	t.Run("Error_MissingConfigPath", func(t *testing.T) {
		// Unset VAULT_PATH to trigger error
		originalPath := os.Getenv("VAULT_PATH")
		os.Unsetenv("VAULT_PATH")
		defer func() {
			if originalPath != "" {
				os.Setenv("VAULT_PATH", originalPath)
			}
		}()

		err := Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load configuration")
	})

	t.Run("Error_InvalidConfigPath", func(t *testing.T) {
		os.Setenv("VAULT_PATH", "/nonexistent/path/credentials.json")
		defer os.Unsetenv("VAULT_PATH")

		err := Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load configuration")
	})

	t.Run("Error_InvalidJSON", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "credentials-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		// Invalid JSON
		_, err = tmpFile.Write([]byte(`{invalid json}`))
		require.NoError(t, err)
		tmpFile.Close()

		os.Setenv("VAULT_PATH", tmpFile.Name())
		defer os.Unsetenv("VAULT_PATH")

		err = Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load configuration")
	})

	t.Run("Error_MissingRequiredFields", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "credentials-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		// Missing required fields
		config := `{
			"SPLUNK_URL": "https://test.splunk.com:8089"
		}`
		_, err = tmpFile.Write([]byte(config))
		require.NoError(t, err)
		tmpFile.Close()

		os.Setenv("VAULT_PATH", tmpFile.Name())
		defer os.Unsetenv("VAULT_PATH")

		err = Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})

	// COMMENTED OUT: Makes actual network calls causing long timeouts
	// t.Run("Error_InvalidSplunkURL", func(t *testing.T) {
	// 	tmpFile, err := os.CreateTemp("", "credentials-*.json")
	// 	require.NoError(t, err)
	// 	defer os.Remove(tmpFile.Name())

	// 	config := `{
	// 		"SPLUNK_URL": "invalid-url",
	// 		"SPLUNK_USERNAME": "admin",
	// 		"SPLUNK_PASSWORD": "password",
	// 		"SPLUNK_INDEX_NAME": "test_index",
	// 		"SALESFORCE_ENDPOINT": "https://login.salesforce.com",
	// 		"SALESFORCE_CLIENT_ID": "test-client",
	// 		"SALESFORCE_CLIENT_SECRET": "test-secret",
	// 		"SALESFORCE_ACCOUNT_NAME": "test_account",
	// 		"DATA_INPUTS": [{
	// 			"name": "test_input",
	// 			"object": "Account",
	// 			"object_fields": "Id,Name"
	// 		}]
	// 	}`
	// 	_, err = tmpFile.Write([]byte(config))
	// 	require.NoError(t, err)
	// 	tmpFile.Close()

	// 	os.Setenv("VAULT_PATH", tmpFile.Name())
	// 	defer os.Unsetenv("VAULT_PATH")

	// 	err = Execute()
	// 	require.Error(t, err)
	// })

	// COMMENTED OUT: Makes actual network calls causing long timeouts
	// t.Run("Error_InvalidSalesforceEndpoint", func(t *testing.T) {
	// 	tmpFile, err := os.CreateTemp("", "credentials-*.json")
	// 	require.NoError(t, err)
	// 	defer os.Remove(tmpFile.Name())

	// 	config := `{
	// 		"SPLUNK_URL": "https://test.splunk.com:8089",
	// 		"SPLUNK_USERNAME": "admin",
	// 		"SPLUNK_PASSWORD": "password",
	// 		"SPLUNK_INDEX_NAME": "test_index",
	// 		"SALESFORCE_ENDPOINT": "invalid-endpoint",
	// 		"SALESFORCE_CLIENT_ID": "test-client",
	// 		"SALESFORCE_CLIENT_SECRET": "test-secret",
	// 		"SALESFORCE_ACCOUNT_NAME": "test_account",
	// 		"DATA_INPUTS": [{
	// 			"name": "test_input",
	// 			"object": "Account",
	// 			"object_fields": "Id,Name"
	// 		}]
	// 	}`
	// 	_, err = tmpFile.Write([]byte(config))
	// 	require.NoError(t, err)
	// 	tmpFile.Close()

	// 	os.Setenv("VAULT_PATH", tmpFile.Name())
	// 	defer os.Unsetenv("VAULT_PATH")

	// 	err = Execute()
	// 	require.Error(t, err)
	// 	// Should fail during config validation or early execution
	// 	assert.Error(t, err)
	// })

	t.Run("Error_EmptyDataInputs", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "credentials-*.json")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		config := `{
			"SPLUNK_URL": "https://test.splunk.com:8089",
			"SPLUNK_USERNAME": "admin",
			"SPLUNK_PASSWORD": "password",
			"SPLUNK_INDEX_NAME": "test_index",
			"SALESFORCE_ENDPOINT": "https://login.salesforce.com",
			"SALESFORCE_CLIENT_ID": "test-client",
			"SALESFORCE_CLIENT_SECRET": "test-secret",
			"SALESFORCE_ACCOUNT_NAME": "test_account",
			"DATA_INPUTS": []
		}`
		_, err = tmpFile.Write([]byte(config))
		require.NoError(t, err)
		tmpFile.Close()

		os.Setenv("VAULT_PATH", tmpFile.Name())
		defer os.Unsetenv("VAULT_PATH")

		err = Execute()
		require.Error(t, err)
	})

	// COMMENTED OUT: Makes actual network calls causing long timeouts
	// t.Run("Error_NetworkFailure", func(t *testing.T) {
	// 	tmpFile, err := os.CreateTemp("", "credentials-*.json")
	// 	require.NoError(t, err)
	// 	defer os.Remove(tmpFile.Name())

	// 	// Valid config but will fail on network call
	// 	config := `{
	// 		"SPLUNK_URL": "https://invalid-splunk-host-12345.com:8089",
	// 		"SPLUNK_USERNAME": "admin",
	// 		"SPLUNK_PASSWORD": "password",
	// 		"SPLUNK_INDEX_NAME": "test_index",
	// 		"SALESFORCE_ENDPOINT": "https://login.salesforce.com",
	// 		"SALESFORCE_CLIENT_ID": "test-client",
	// 		"SALESFORCE_CLIENT_SECRET": "test-secret",
	// 		"SALESFORCE_ACCOUNT_NAME": "test_account",
	// 		"DATA_INPUTS": [{
	// 			"name": "test_input",
	// 			"object": "Account",
	// 			"object_fields": "Id,Name"
	// 		}]
	// 	}`
	// 	_, err = tmpFile.Write([]byte(config))
	// 	require.NoError(t, err)
	// 	tmpFile.Close()

	// 	os.Setenv("VAULT_PATH", tmpFile.Name())
	// 	defer os.Unsetenv("VAULT_PATH")

	// 	// This should fail with network error or service creation error
	// 	err = Execute()
	// 	assert.Error(t, err)
	// 	// Should contain either network error or auth error
	// })
}
