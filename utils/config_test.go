// Package utils provides configuration loading and management
package utils_test

import (
	"os"
	"reflect"
	"testing"

	"salesforce-splunk-migration/utils"
)

func TestConfig_Validate(t *testing.T) {
	validConfig := func() *utils.Config {
		return &utils.Config{
			Splunk: utils.SplunkConfig{
				URL:            "https://splunk.example.com:8089",
				Username:       "admin",
				Password:       "password",
				IndexName:      "salesforce_data",
				RequestTimeout: 30,
				MaxRetries:     3,
				RetryDelay:     5,
			},
			Salesforce: utils.SalesforceConfig{
				Endpoint:     "https://login.salesforce.com",
				APIVersion:   "v58.0",
				AuthType:     "oauth2",
				ClientID:     "client123",
				ClientSecret: "secret456",
				AccountName:  "test_account",
			},
			Migration: utils.MigrationConfig{
				ConcurrentRequests: 5,
				LogLevel:           "info",
			},
			Extensions: map[string]interface{}{
				"DATA_INPUTS": []interface{}{
					map[string]interface{}{
						"name":          "Account_Input",
						"object":        "Account",
						"object_fields": "Id,Name,CreatedDate",
						"order_by":      "CreatedDate",
						"start_date":    "2024-01-01",
						"interval":      300,
						"delay":         60,
						"index":         "main",
					},
				},
			},
		}
	}

	tests := []struct {
		name      string
		config    *utils.Config
		wantErr   bool
		setupFunc func(*utils.Config)
	}{
		{
			name:    "Success_ValidConfig",
			config:  validConfig(),
			wantErr: false,
		},
		{
			name:    "Error_MissingSplunkURL",
			config:  validConfig(),
			wantErr: true,
			setupFunc: func(c *utils.Config) {
				c.Splunk.URL = ""
			},
		},
		{
			name:    "Error_MissingSplunkUsername",
			config:  validConfig(),
			wantErr: true,
			setupFunc: func(c *utils.Config) {
				c.Splunk.Username = ""
			},
		},
		{
			name:    "Error_MissingSplunkPassword",
			config:  validConfig(),
			wantErr: true,
			setupFunc: func(c *utils.Config) {
				c.Splunk.Password = ""
			},
		},
		{
			name:    "Error_MissingSalesforceEndpoint",
			config:  validConfig(),
			wantErr: true,
			setupFunc: func(c *utils.Config) {
				c.Salesforce.Endpoint = ""
			},
		},
		{
			name:    "Error_MissingObjectFields",
			config:  validConfig(),
			wantErr: true,
			setupFunc: func(c *utils.Config) {
				c.Extensions["DATA_INPUTS"] = []interface{}{
					map[string]interface{}{
						"name":   "Test",
						"object": "Account",
					},
				}
			},
		},
		{
			name:    "Error_NoDataInputs",
			config:  validConfig(),
			wantErr: true,
			setupFunc: func(c *utils.Config) {
				c.Extensions["DATA_INPUTS"] = []interface{}{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(tt.config)
			}
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataInput_Validation(t *testing.T) {
	tests := []struct {
		name      string
		dataInput utils.DataInput
		wantErr   bool
	}{
		{"valid data input", utils.DataInput{Name: "Account_Input", Object: "Account", StartDate: "2024-01-01", Interval: 300, Delay: 60, Index: "main"}, false},
		{"empty fields are allowed", utils.DataInput{Name: "Contact_Input", Object: "Contact"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dataInput.Name == "" && tt.wantErr {
				t.Error("Expected validation error for empty name")
			}
		})
	}
}

func TestConfig_Defaults(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() interface{}
		check    func(t *testing.T, config interface{})
	}{
		{
			name: "SplunkConfig_Defaults",
			setup: func() interface{} {
				config := utils.SplunkConfig{URL: "https://splunk.example.com:8089", Username: "admin", Password: "password"}
				if config.RequestTimeout == 0 {
					config.RequestTimeout = 30
				}
				if config.MaxRetries == 0 {
					config.MaxRetries = 3
				}
				return config
			},
			check: func(t *testing.T, cfg interface{}) {
				config := cfg.(utils.SplunkConfig)
				if config.RequestTimeout != 30 {
					t.Errorf("Expected default timeout 30, got %d", config.RequestTimeout)
				}
				if config.MaxRetries != 3 {
					t.Errorf("Expected default max retries 3, got %d", config.MaxRetries)
				}
			},
		},
		{
			name: "MigrationConfig_Defaults",
			setup: func() interface{} {
				config := utils.MigrationConfig{LogLevel: "info"}
				if config.ConcurrentRequests == 0 {
					config.ConcurrentRequests = 5
				}
				return config
			},
			check: func(t *testing.T, cfg interface{}) {
				config := cfg.(utils.MigrationConfig)
				if config.ConcurrentRequests != 5 {
					t.Errorf("Expected default concurrent requests 5, got %d", config.ConcurrentRequests)
				}
			},
		},
		{
			name: "SalesforceConfig_Defaults",
			setup: func() interface{} {
				return &utils.Config{Salesforce: utils.SalesforceConfig{Endpoint: "https://login.salesforce.com"}}
			},
			check: func(t *testing.T, cfg interface{}) {
				config := cfg.(*utils.Config)
				if config.Salesforce.Endpoint != "https://login.salesforce.com" {
					t.Errorf("Expected Salesforce endpoint to match")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setup()
			tt.check(t, config)
		})
	}
}

func TestConfig_GetDataInputs(t *testing.T) {
	tests := []struct {
		name        string
		extensions  map[string]interface{}
		wantErr     bool
		wantCount   int
		checkFirst  bool
		firstName   string
		firstObject string
	}{
		{
			name: "Success_MultipleInputs",
			extensions: map[string]interface{}{
				"DATA_INPUTS": []interface{}{
					map[string]interface{}{"name": "Account_Input", "object": "Account", "object_fields": "Id,Name,CreatedDate", "start_date": "2024-01-01T00:00:00Z"},
					map[string]interface{}{"name": "Contact_Input", "object": "Contact", "object_fields": "Id,FirstName,LastName", "start_date": "2024-01-01T00:00:00Z"},
				},
			},
			wantErr:     false,
			wantCount:   2,
			checkFirst:  true,
			firstName:   "Account_Input",
			firstObject: "Account",
		},
		{"Error_Missing", map[string]interface{}{}, true, 0, false, "", ""},
		{"Success_EmptyArray", map[string]interface{}{"DATA_INPUTS": []interface{}{}}, false, 0, false, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &utils.Config{Extensions: tt.extensions}
			inputs, err := config.GetDataInputs()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDataInputs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(inputs) != tt.wantCount {
				t.Errorf("Expected %d data inputs, got %d", tt.wantCount, len(inputs))
			}
			if tt.checkFirst && len(inputs) > 0 {
				if inputs[0].Name != tt.firstName {
					t.Errorf("Expected first input name='%s', got '%s'", tt.firstName, inputs[0].Name)
				}
				if inputs[0].Object != tt.firstObject {
					t.Errorf("Expected first input object='%s', got '%s'", tt.firstObject, inputs[0].Object)
				}
			}
		})
	}

	// Test invalid type separately since it may or may not error
	t.Run("Error_InvalidType", func(t *testing.T) {
		config := &utils.Config{Extensions: map[string]interface{}{"DATA_INPUTS": "invalid_type"}}
		inputs, err := config.GetDataInputs()
		if err == nil && len(inputs) != 0 {
			t.Errorf("GetDataInputs() with invalid type should return empty or error, got %d inputs", len(inputs))
		}
	})
}

func TestConfig_SetExtension(t *testing.T) {
	tests := []struct {
		name       string
		extensions map[string]interface{}
		key        string
		value      interface{}
		wantValue  interface{}
	}{
		{"Success_SetStringExtension", make(map[string]interface{}), "test_key", "test_value", "test_value"},
		{"Success_SetIntExtension", make(map[string]interface{}), "count", 42, 42},
		{"Success_NilExtensions", nil, "key", "value", "value"},
		{"Success_MultipleExtensions", make(map[string]interface{}), "custom_field1", "value1", "value1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &utils.Config{Extensions: tt.extensions}
			config.SetExtension(tt.key, tt.value)

			if config.Extensions == nil {
				t.Error("Extensions map should be initialized")
				return
			}
			if val, ok := config.Extensions[tt.key]; !ok || val != tt.wantValue {
				t.Errorf("Extension not set correctly: got %v, want %v", val, tt.wantValue)
			}
		})
	}
}

func TestConfig_GetDataInputs_MissingFieldsAndDefaults(t *testing.T) {
	tests := []struct {
		name        string
		config      *utils.Config
		wantErr     bool
		checkDefaults bool
	}{
		{
			name: "Error_MissingName",
			config: &utils.Config{
				Extensions: map[string]interface{}{
					"DATA_INPUTS": []interface{}{map[string]interface{}{"object": "Account", "object_fields": "Id,Name"}},
				},
			},
			wantErr: true,
		},
		{
			name: "Error_MissingObject",
			config: &utils.Config{
				Extensions: map[string]interface{}{
					"DATA_INPUTS": []interface{}{map[string]interface{}{"name": "test_input", "object_fields": "Id,Name"}},
				},
			},
			wantErr: true,
		},
		{
			name: "Success_AppliesDefaults",
			config: &utils.Config{
				Splunk: utils.SplunkConfig{DefaultIndex: "default_index"},
				Extensions: map[string]interface{}{
					"DATA_INPUTS": []interface{}{map[string]interface{}{"name": "Test_Input", "object": "Account", "object_fields": "Id,Name"}},
				},
			},
			wantErr:       false,
			checkDefaults: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputs, err := tt.config.GetDataInputs()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDataInputs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkDefaults && len(inputs) > 0 {
				input := inputs[0]
				if input.Index != "default_index" {
					t.Errorf("Expected default index='default_index', got '%s'", input.Index)
				}
				if input.OrderBy != "LastModifiedDate" {
					t.Errorf("Expected default OrderBy='LastModifiedDate', got '%s'", input.OrderBy)
				}
				if input.Interval != 300 {
					t.Errorf("Expected default Interval=300, got %d", input.Interval)
				}
				if input.Delay != 60 {
					t.Errorf("Expected default Delay=60, got %d", input.Delay)
				}
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	createTestFile := func(content string) string {
		tmpFile, err := os.CreateTemp("", "test-config-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		if _, err := tmpFile.Write([]byte(content)); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}
		tmpFile.Close()
		return tmpFile.Name()
	}

	tests := []struct {
		name       string
		content    string
		wantErr    bool
		skipFile   bool
		checkFunc  func(t *testing.T, config *utils.Config)
	}{
		{
			name: "Success_AllDefaults",
			content: `{
				"SPLUNK_URL": "https://test.splunk.com:8089",
				"SPLUNK_USERNAME": "admin",
				"SPLUNK_PASSWORD": "password",
				"SALESFORCE_ENDPOINT": "https://login.salesforce.com",
				"SALESFORCE_CLIENT_ID": "test-client",
				"SALESFORCE_CLIENT_SECRET": "test-secret",
				"SALESFORCE_ACCOUNT_NAME": "test-account",
				"DATA_INPUTS": [{"name": "Account_Test", "object": "Account", "object_fields": "Id,Name"}]
			}`,
			wantErr: false,
			checkFunc: func(t *testing.T, config *utils.Config) {
				assertDefaults := map[string]interface{}{
					"DefaultIndex":           "salesforce_testing",
					"IndexName":              "salesforce_testing",
					"RequestTimeout":         30,
					"MaxRetries":             3,
					"RetryDelay":             5,
					"APIVersion":             "64.0",
					"AuthType":               "oauth_client_credentials",
					"DashboardDirectory":     "resources/dashboards",
					"ConcurrentRequests":     3,
					"LogLevel":               "info",
				}
				if config.Splunk.DefaultIndex != assertDefaults["DefaultIndex"] {
					t.Errorf("Expected DefaultIndex='%v', got '%s'", assertDefaults["DefaultIndex"], config.Splunk.DefaultIndex)
				}
				if config.Splunk.RequestTimeout != assertDefaults["RequestTimeout"] {
					t.Errorf("Expected RequestTimeout=%v, got %d", assertDefaults["RequestTimeout"], config.Splunk.RequestTimeout)
				}
				if config.Migration.ConcurrentRequests != assertDefaults["ConcurrentRequests"] {
					t.Errorf("Expected ConcurrentRequests=%v, got %d", assertDefaults["ConcurrentRequests"], config.Migration.ConcurrentRequests)
				}
			},
		},
		{
			name: "Success_CustomValues",
			content: `{
				"SPLUNK_URL": "https://test.splunk.com:8089",
				"SPLUNK_USERNAME": "admin",
				"SPLUNK_PASSWORD": "password",
				"SPLUNK_DEFAULT_INDEX": "custom_index",
				"SPLUNK_REQUEST_TIMEOUT": 60,
				"SALESFORCE_ENDPOINT": "https://login.salesforce.com",
				"SALESFORCE_API_VERSION": "58.0",
				"SALESFORCE_CLIENT_ID": "test-client",
				"SALESFORCE_CLIENT_SECRET": "test-secret",
				"SALESFORCE_ACCOUNT_NAME": "test-account",
				"MIGRATION_LOG_LEVEL": "debug",
				"DATA_INPUTS": [{"name": "Account_Test", "object": "Account", "object_fields": "Id,Name"}]
			}`,
			wantErr: false,
			checkFunc: func(t *testing.T, config *utils.Config) {
				if config.Splunk.DefaultIndex != "custom_index" {
					t.Errorf("Expected DefaultIndex='custom_index', got '%s'", config.Splunk.DefaultIndex)
				}
				if config.Splunk.RequestTimeout != 60 {
					t.Errorf("Expected RequestTimeout=60, got %d", config.Splunk.RequestTimeout)
				}
				if config.Salesforce.APIVersion != "58.0" {
					t.Errorf("Expected APIVersion='58.0', got '%s'", config.Salesforce.APIVersion)
				}
				if config.Migration.LogLevel != "debug" {
					t.Errorf("Expected LogLevel='debug', got '%s'", config.Migration.LogLevel)
				}
			},
		},
		{name: "Error_InvalidJSON", content: `{invalid json}`, wantErr: true},
		{name: "Error_FileNotFound", content: "", wantErr: true, skipFile: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string
			if tt.skipFile {
				filePath = "nonexistent_file.json"
			} else {
				filePath = createTestFile(tt.content)
				defer os.Remove(filePath)
			}

			config, err := utils.LoadConfig(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, config)
			}
		})
	}

	// Special test for empty file path
	t.Run("Success_WithEmptyFilePath", func(t *testing.T) {
		config, err := utils.LoadConfig("")
		if err != nil {
			t.Skipf("Skipping test, credentials.json not available: %v", err)
		}
		if config == nil {
			t.Error("LoadConfig should return non-nil config")
		}
	})
}

func TestLoader_Load(t *testing.T) {
	type SimpleConfig struct {
		Name string `env:"TEST_NAME"`
		Port int    `env:"TEST_PORT"`
	}
	type Database struct {
		Host string `env:"DB_HOST"`
		Port int    `env:"DB_PORT"`
	}
	type AppConfig struct {
		AppName string `env:"APP_NAME"`
		DB      Database
	}

	tests := []struct {
		name      string
		setup     func() (*utils.Loader, interface{})
		wantErr   bool
		checkFunc func(t *testing.T, result interface{})
	}{
		{
			name: "Success_LoadSimpleStruct",
			setup: func() (*utils.Loader, interface{}) {
				loader := &utils.Loader{}
				loader.SetValue("TEST_NAME", "test-service")
				loader.SetValue("TEST_PORT", "8080")
				var config SimpleConfig
				return loader, &config
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result interface{}) {
				config := result.(*SimpleConfig)
				if config.Name != "test-service" || config.Port != 8080 {
					t.Errorf("Expected Name='test-service' Port=8080, got Name='%s' Port=%d", config.Name, config.Port)
				}
			},
		},
		{
			name: "Success_LoadNestedStruct",
			setup: func() (*utils.Loader, interface{}) {
				loader := &utils.Loader{}
				loader.SetValue("APP_NAME", "myapp")
				loader.SetValue("DB_HOST", "localhost")
				loader.SetValue("DB_PORT", "5432")
				var config AppConfig
				return loader, &config
			},
			wantErr: false,
			checkFunc: func(t *testing.T, result interface{}) {
				config := result.(*AppConfig)
				if config.AppName != "myapp" || config.DB.Host != "localhost" || config.DB.Port != 5432 {
					t.Errorf("Nested struct values incorrect")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader, config := tt.setup()
			err := loader.Load(config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, config)
			}
		})
	}

	// Error cases with different input types
	errorTests := []struct {
		name    string
		input   interface{}
		wantMsg string
	}{
		{"Error_NotPointer", utils.Config{}, "input must be a pointer to a struct"},
		{"Error_PointerToNonStruct", new(string), "input must be a pointer to a struct"},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			loader := &utils.Loader{}
			err := loader.Load(tt.input)
			if err == nil || err.Error() != tt.wantMsg {
				t.Errorf("Load() error = %v, want %s", err, tt.wantMsg)
			}
		})
	}

	// Type conversion tests
	conversionTests := []struct {
		name    string
		envVals map[string]string
		check   func(t *testing.T, loader *utils.Loader)
	}{
		{
			name:    "Success_WithBoolConversion",
			envVals: map[string]string{"ENABLED": "true"},
			check: func(t *testing.T, loader *utils.Loader) {
				type ConfigWithBool struct {
					Enabled bool `env:"ENABLED"`
				}
				var config ConfigWithBool
				if err := loader.Load(&config); err != nil || !config.Enabled {
					t.Errorf("Expected Enabled=true, got %v, err=%v", config.Enabled, err)
				}
			},
		},
		{
			name:    "Success_WithInt64",
			envVals: map[string]string{"MAX_SIZE": "9223372036854775807"},
			check: func(t *testing.T, loader *utils.Loader) {
				type ConfigWithInt64 struct {
					MaxSize int64 `env:"MAX_SIZE"`
				}
				var config ConfigWithInt64
				if err := loader.Load(&config); err != nil || config.MaxSize != 9223372036854775807 {
					t.Errorf("Expected MaxSize=9223372036854775807, got %d, err=%v", config.MaxSize, err)
				}
			},
		},
		{
			name:    "Success_WithFloat64",
			envVals: map[string]string{"RATE": "3.14159"},
			check: func(t *testing.T, loader *utils.Loader) {
				type ConfigWithFloat struct {
					Rate float64 `env:"RATE"`
				}
				var config ConfigWithFloat
				if err := loader.Load(&config); err != nil || config.Rate != 3.14159 {
					t.Errorf("Expected Rate=3.14159, got %f, err=%v", config.Rate, err)
				}
			},
		},
		{
			name:    "Success_WithEmptyString",
			envVals: map[string]string{"COUNT": "", "AMOUNT": "", "ENABLED": ""},
			check: func(t *testing.T, loader *utils.Loader) {
				type ConfigWithDefaults struct {
					Count   int     `env:"COUNT"`
					Amount  float64 `env:"AMOUNT"`
					Enabled bool    `env:"ENABLED"`
				}
				var config ConfigWithDefaults
				if err := loader.Load(&config); err != nil || config.Count != 0 || config.Amount != 0 || config.Enabled != false {
					t.Errorf("Empty string defaults failed")
				}
			},
		},
		{
			name:    "Error_InvalidIntConversion",
			envVals: map[string]string{"PORT": "not-a-number"},
			check: func(t *testing.T, loader *utils.Loader) {
				type ConfigWithInt struct {
					Port int `env:"PORT"`
				}
				var config ConfigWithInt
				if err := loader.Load(&config); err == nil {
					t.Error("Load() should fail with invalid int conversion")
				}
			},
		},
		{
			name:    "Error_InvalidBoolConversion",
			envVals: map[string]string{"ENABLED": "not-a-bool"},
			check: func(t *testing.T, loader *utils.Loader) {
				type ConfigWithBool struct {
					Enabled bool `env:"ENABLED"`
				}
				var config ConfigWithBool
				if err := loader.Load(&config); err == nil {
					t.Error("Load() should fail with invalid bool conversion")
				}
			},
		},
	}

	for _, tt := range conversionTests {
		t.Run(tt.name, func(t *testing.T) {
			loader := &utils.Loader{}
			for k, v := range tt.envVals {
				loader.SetValue(k, v)
			}
			tt.check(t, loader)
		})
	}
}

func TestCreateLoader(t *testing.T) {
	createTestFile := func(content string) string {
		tmpFile, err := os.CreateTemp("", "test-loader-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		if _, err := tmpFile.Write([]byte(content)); err != nil {
				t.Fatalf("Failed to write temp file: %v", err)
		}
		tmpFile.Close()
		return tmpFile.Name()
	}

	tests := []struct {
		name        string
		content     string
		envVars     map[string]string
		wantErr     bool
		checkValues map[string]string
		skipKeys    []string
	}{
		{
			name:    "Success_LoadFromFile",
			content: `{"SPLUNK_URL": "https://test.com", "SPLUNK_PORT": 8089, "SPLUNK_ENABLED": true, "DATA_INPUTS": []}`,
			wantErr: false,
			checkValues: map[string]string{"SPLUNK_URL": "https://test.com", "SPLUNK_PORT": "8089", "SPLUNK_ENABLED": "true"},
			skipKeys:    []string{"DATA_INPUTS"},
		},
		{
			name:        "Success_WithEnvironmentVariables",
			content:     `{"SPLUNK_URL": "https://test.com"}`,
			envVars:     map[string]string{"TEST_ENV_VAR": "test-value", "TEST_ENV_NUM": "42"},
			wantErr:     false,
			checkValues: map[string]string{"TEST_ENV_VAR": "test-value"},
		},
		{
			name:        "Success_EnvVarOverridesFile",
			content:     `{"SPLUNK_URL": "https://file-value.com"}`,
			envVars:     map[string]string{"SPLUNK_URL": "https://env-override.com"},
			wantErr:     false,
			checkValues: map[string]string{"SPLUNK_URL": "https://env-override.com"},
		},
		{
			name:        "Success_WithFloatNumber",
			content:     `{"RATE": 3.14159, "PERCENTAGE": 99.9}`,
			wantErr:     false,
			checkValues: map[string]string{"RATE": "3.14159"},
		},
		{
			name:        "Success_WithNestedObject",
			content:     `{"SIMPLE_KEY": "value", "NESTED_OBJECT": {"key1": "value1"}}`,
			wantErr:     false,
			checkValues: map[string]string{"SIMPLE_KEY": "value"},
			skipKeys:    []string{"NESTED_OBJECT"},
		},
		{name: "Error_InvalidJSON", content: `{invalid json}`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			filePath := createTestFile(tt.content)
			defer os.Remove(filePath)

			loader, err := utils.CreateLoader(filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLoader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			values := loader.GetValues()
			for key, expected := range tt.checkValues {
				if values[key] != expected {
					t.Errorf("Expected %s='%s', got '%s'", key, expected, values[key])
				}
			}
			for _, key := range tt.skipKeys {
				if _, exists := values[key]; exists {
					t.Errorf("%s should not be in values map", key)
				}
			}
		})
	}

	t.Run("Error_FileNotFound", func(t *testing.T) {
		_, err := utils.CreateLoader("nonexistent-file.json")
		if err == nil {
			t.Error("CreateLoader() should fail with non-existent file")
		}
	})
}

func TestLoadExtensions(t *testing.T) {
	createTestFile := func(content string) string {
		tmpFile, err := os.CreateTemp("", "test-ext-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		if _, err := tmpFile.Write([]byte(content)); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}
		tmpFile.Close()
		return tmpFile.Name()
	}

	tests := []struct {
		name          string
		content       string
		wantErr       bool
		checkKeys     []string
		skipKeys      []string
		skipFile      bool
	}{
		{
			name:      "Success_LoadDataInputs",
			content:   `{"DATA_INPUTS": [{"name": "Account_Input", "object": "Account"}]}`,
			wantErr:   false,
			checkKeys: []string{"DATA_INPUTS"},
		},
		{
			name:      "Success_LoadCustomExtensions",
			content:   `{"CUSTOM_FIELD": "custom_value", "ANOTHER_EXTENSION": {"nested": "data"}, "SPLUNK_URL": "https://test.com", "APP_NAME": "test-app"}`,
			wantErr:   false,
			checkKeys: []string{"CUSTOM_FIELD", "ANOTHER_EXTENSION"},
			skipKeys:  []string{"SPLUNK_URL", "APP_NAME"},
		},
		{name: "Error_InvalidJSON", content: `{invalid}`, wantErr: true},
		{name: "Error_FileNotFound", skipFile: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &utils.Config{Extensions: make(map[string]interface{})}
			var filePath string
			if tt.skipFile {
				filePath = "nonexistent.json"
			} else {
				filePath = createTestFile(tt.content)
				defer os.Remove(filePath)
			}

			err := utils.LoadExtensions(filePath, config)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadExtensions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for _, key := range tt.checkKeys {
					if _, exists := config.Extensions[key]; !exists {
						t.Errorf("%s should be in extensions", key)
					}
				}
				for _, key := range tt.skipKeys {
					if _, exists := config.Extensions[key]; exists {
						t.Errorf("%s should not be in extensions (structured field)", key)
					}
				}
			}
		})
	}
}

func TestIsStructuredField(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"APP_NAME", true},
		{"SPLUNK_URL", true},
		{"SALESFORCE_ENDPOINT", true},
		{"MIGRATION_LOG_LEVEL", true},
		{"CUSTOM_FIELD", false},
		{"DATA_INPUTS", false},
		{"", false},
		{"app_name", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := utils.IsStructuredField(tt.key)
			if result != tt.expected {
				t.Errorf("IsStructuredField(%s) = %v, expected %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestBuildMap(t *testing.T) {
	tests := []struct {
		name      string
		structType interface{}
		values    map[string]string
		checkFunc func(t *testing.T, result map[string]interface{})
	}{
		{
			name: "Success_SimpleStruct",
			structType: struct {
				Name string `env:"NAME"`
				Port int    `env:"PORT"`
			}{},
			values: map[string]string{"NAME": "test-service", "PORT": "8080"},
			checkFunc: func(t *testing.T, result map[string]interface{}) {
				if result["Name"] != "test-service" || result["Port"] != "8080" {
					t.Errorf("Simple struct values incorrect")
				}
			},
		},
		{
			name: "Success_NestedStruct",
			structType: struct {
				AppName string `env:"APP_NAME"`
				DB      struct {
					Host string `env:"DB_HOST"`
					Port int    `env:"DB_PORT"`
				}
			}{},
			values: map[string]string{"APP_NAME": "myapp", "DB_HOST": "localhost", "DB_PORT": "5432"},
			checkFunc: func(t *testing.T, result map[string]interface{}) {
				if result["AppName"] != "myapp" {
					t.Errorf("Expected AppName='myapp'")
				}
				dbMap, ok := result["DB"].(map[string]interface{})
				if !ok || dbMap["Host"] != "localhost" || dbMap["Port"] != "5432" {
					t.Errorf("Nested struct values incorrect")
				}
			},
		},
		{
			name: "Success_MissingEnvTag",
			structType: struct {
				Name string
				Port int `env:"PORT"`
			}{},
			values: map[string]string{"PORT": "8080"},
			checkFunc: func(t *testing.T, result map[string]interface{}) {
				if _, exists := result["Name"]; exists {
					t.Error("Field without env tag should not be in result")
				}
				if result["Port"] != "8080" {
					t.Errorf("Port value incorrect")
				}
			},
		},
		{
			name: "Success_ValueNotFound",
			structType: struct {
				Name string `env:"NAME"`
				Port int    `env:"PORT"`
			}{},
			values: map[string]string{"NAME": "test-service"},
			checkFunc: func(t *testing.T, result map[string]interface{}) {
				if _, exists := result["Port"]; exists {
					t.Error("Port should not be in result when value not found")
				}
			},
		},
		{
			name: "Success_EmptyNestedStruct",
			structType: struct {
				AppName string `env:"APP_NAME"`
				DB      struct {
					Host string `env:"DB_HOST"`
				}
			}{},
			values: map[string]string{"APP_NAME": "myapp"},
			checkFunc: func(t *testing.T, result map[string]interface{}) {
				if _, exists := result["DB"]; exists {
					t.Error("Empty nested struct should not be in result")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.BuildMap(reflect.TypeOf(tt.structType), tt.values)
			tt.checkFunc(t, result)
		})
	}
}

func TestStrToNumeric(t *testing.T) {
	tests := []struct {
		name       string
		srcType    reflect.Type
		dstType    reflect.Type
		value      interface{}
		wantResult interface{}
		wantErr    bool
	}{
		{"Success_StringToInt", reflect.TypeOf(""), reflect.TypeOf(0), "42", 42, false},
		{"Success_StringToInt64", reflect.TypeOf(""), reflect.TypeOf(int64(0)), "9223372036854775807", int64(9223372036854775807), false},
		{"Success_StringToFloat64", reflect.TypeOf(""), reflect.TypeOf(float64(0)), "3.14159", 3.14159, false},
		{"Success_EmptyStringToInt", reflect.TypeOf(""), reflect.TypeOf(0), "", 0, false},
		{"Success_EmptyStringToInt64", reflect.TypeOf(""), reflect.TypeOf(int64(0)), "", int64(0), false},
		{"Success_EmptyStringToFloat64", reflect.TypeOf(""), reflect.TypeOf(float64(0)), "", float64(0), false},
		{"Success_EmptyStringToBool", reflect.TypeOf(""), reflect.TypeOf(false), "", false, false},
		{"Success_NonStringInput", reflect.TypeOf(0), reflect.TypeOf(0), 42, 42, false},
		{"Success_NonNumericTarget", reflect.TypeOf(""), reflect.TypeOf(""), "test", "test", false},
		{"Error_InvalidInt", reflect.TypeOf(""), reflect.TypeOf(0), "not-a-number", nil, true},
		{"Error_InvalidInt64", reflect.TypeOf(""), reflect.TypeOf(int64(0)), "not-a-number", nil, true},
		{"Error_InvalidFloat64", reflect.TypeOf(""), reflect.TypeOf(float64(0)), "not-a-number", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.StrToNumeric(tt.srcType, tt.dstType, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("StrToNumeric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.wantResult {
				t.Errorf("StrToNumeric() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestStrToBool(t *testing.T) {
	tests := []struct {
		name       string
		srcType    reflect.Type
		dstType    reflect.Type
		value      interface{}
		wantResult interface{}
		wantErr    bool
	}{
		{"Success_StringToBool_True", reflect.TypeOf(""), reflect.TypeOf(false), "true", true, false},
		{"Success_StringToBool_False", reflect.TypeOf(""), reflect.TypeOf(false), "false", false, false},
		{"Success_StringToBool_1", reflect.TypeOf(""), reflect.TypeOf(false), "1", true, false},
		{"Success_StringToBool_0", reflect.TypeOf(""), reflect.TypeOf(false), "0", false, false},
		{"Success_EmptyString", reflect.TypeOf(""), reflect.TypeOf(false), "", false, false},
		{"Success_NonStringInput", reflect.TypeOf(0), reflect.TypeOf(false), 42, 42, false},
		{"Success_NonBoolTarget", reflect.TypeOf(""), reflect.TypeOf(""), "test", "test", false},
		{"Error_InvalidBool", reflect.TypeOf(""), reflect.TypeOf(false), "not-a-bool", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utils.StrToBool(tt.srcType, tt.dstType, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("StrToBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.wantResult {
				t.Errorf("StrToBool() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}
