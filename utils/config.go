// Package utils provides configuration loading and management
package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// Config holds the complete application configuration
type Config struct {
	Splunk     SplunkConfig           `env:"SPLUNK"`
	Salesforce SalesforceConfig       `env:"SALESFORCE"`
	Migration  MigrationConfig        `env:"MIGRATION"`
	Extensions map[string]interface{} `json:"-"` // Store DATA_INPUTS and other dynamic config
}

// SplunkConfig holds Splunk-specific configuration
type SplunkConfig struct {
	URL                string `env:"SPLUNK_URL"`
	Username           string `env:"SPLUNK_USERNAME"`
	Password           string `env:"SPLUNK_PASSWORD"`
	TokenName          string `env:"SPLUNK_TOKEN_NAME"`     // Token name for /authorization/tokens endpoint
	TokenAudience      string `env:"SPLUNK_TOKEN_AUDIENCE"` // Token audience (e.g., "Automation")
	SkipSSLVerify      bool   `env:"SPLUNK_SKIP_SSL_VERIFY"`
	DefaultIndex       string `env:"SPLUNK_DEFAULT_INDEX"`
	IndexName          string `env:"SPLUNK_INDEX_NAME"`
	MaxTotalDataSizeMB int    `env:"SPLUNK_MAX_TOTAL_DATA_SIZE_MB"`
	RequestTimeout     int    `env:"SPLUNK_REQUEST_TIMEOUT"`
	MaxRetries         int    `env:"SPLUNK_MAX_RETRIES"`
	RetryDelay         int    `env:"SPLUNK_RETRY_DELAY"`
}

// SalesforceConfig holds Salesforce-specific configuration
type SalesforceConfig struct {
	Endpoint     string `env:"SALESFORCE_ENDPOINT"`
	APIVersion   string `env:"SALESFORCE_API_VERSION"`
	AuthType     string `env:"SALESFORCE_AUTH_TYPE"`
	ClientID     string `env:"SALESFORCE_CLIENT_ID"`
	ClientSecret string `env:"SALESFORCE_CLIENT_SECRET"`
	AccountName  string `env:"SALESFORCE_ACCOUNT_NAME"`
}

// MigrationConfig holds migration-specific settings
type MigrationConfig struct {
	DashboardDirectory string `env:"MIGRATION_DASHBOARD_DIRECTORY"`
	ConcurrentRequests int    `env:"MIGRATION_CONCURRENT_REQUESTS"`
	LogLevel           string `env:"MIGRATION_LOG_LEVEL"`
}

// DataInput represents a Salesforce data input configuration
type DataInput struct {
	Name         string `json:"name"`
	Object       string `json:"object"`
	ObjectFields string `json:"object_fields"`
	OrderBy      string `json:"order_by"`
	StartDate    string `json:"start_date"`
	Interval     int    `json:"interval"`
	Delay        int    `json:"delay"`
	Index        string `json:"index"`
}

// Loader handles environment and file-based configuration loading
type Loader struct {
	values map[string]string
}

// SetValue sets a value in the loader (for testing)
func (l *Loader) SetValue(key, value string) {
	if l.values == nil {
		l.values = make(map[string]string)
	}
	l.values[key] = value
}

// GetValues returns all values (for testing)
func (l *Loader) GetValues() map[string]string {
	return l.values
}

// Load populates a struct from the loaded values
func (l *Loader) Load(structPtr interface{}) error {
	val := reflect.ValueOf(structPtr)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("input must be a pointer to a struct")
	}
	dataMap := BuildMap(val.Elem().Type(), l.values)
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result: structPtr,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			StrToNumeric,
			StrToBool,
		),
	})
	if err != nil {
		return err
	}
	if err := decoder.Decode(dataMap); err != nil {
		return fmt.Errorf("failed to populate struct: %w", err)
	}
	return nil
}

// LoadConfig loads configuration using the loader pattern
func LoadConfig(filePath string) (*Config, error) {
	if filePath == "" {
		filePath = "credentials.json"
	}

	// Create loader
	loader, err := CreateLoader(filePath)
	if err != nil {
		return nil, err
	}

	// Load structured config
	config := &Config{
		Extensions: make(map[string]interface{}),
	}
	if err := loader.Load(config); err != nil {
		return nil, err
	}

	// Set defaults
	if config.Splunk.DefaultIndex == "" {
		config.Splunk.DefaultIndex = "salesforce_testing"
	}
	if config.Splunk.IndexName == "" {
		config.Splunk.IndexName = config.Splunk.DefaultIndex
	}
	if config.Splunk.RequestTimeout == 0 {
		config.Splunk.RequestTimeout = 30
	}
	if config.Splunk.MaxRetries == 0 {
		config.Splunk.MaxRetries = 3
	}
	if config.Splunk.RetryDelay == 0 {
		config.Splunk.RetryDelay = 5
	}
	if config.Salesforce.APIVersion == "" {
		config.Salesforce.APIVersion = "64.0"
	}
	if config.Salesforce.AuthType == "" {
		config.Salesforce.AuthType = "oauth_client_credentials"
	}
	if config.Migration.DashboardDirectory == "" {
		config.Migration.DashboardDirectory = "resources/dashboards"
	}
	if config.Migration.ConcurrentRequests == 0 {
		config.Migration.ConcurrentRequests = 3
	}
	if config.Migration.LogLevel == "" {
		config.Migration.LogLevel = "info"
	}

	// Load extensions (DATA_INPUTS, etc.)
	if err := LoadExtensions(filePath, config); err != nil {
		return nil, err
	}

	return config, nil
}

// CreateLoader creates a new loader from file and environment
func CreateLoader(credentialsPath string) (*Loader, error) {
	values := make(map[string]string)

	// Load from file
	if credentialsPath != "" {
		content, err := os.ReadFile(credentialsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read credentials file: %w", err)
		}

		// Parse as generic map first to handle mixed types
		var rawData map[string]interface{}
		if err := json.Unmarshal(content, &rawData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal credentials file: %w", err)
		}

		// Convert only string values to the values map (skip arrays and objects)
		for key, value := range rawData {
			if strVal, ok := value.(string); ok {
				values[key] = strVal
			} else if numVal, ok := value.(float64); ok {
				// Convert numbers to strings
				values[key] = strconv.FormatFloat(numVal, 'f', -1, 64)
			} else if boolVal, ok := value.(bool); ok {
				// Convert booleans to strings
				values[key] = strconv.FormatBool(boolVal)
			}
			// Skip arrays and nested objects - they'll be handled by loadExtensions
		}
	}

	// Merge with environment variables (env vars take precedence)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			values[pair[0]] = pair[1]
		}
	}

	return &Loader{values: values}, nil
}

// LoadExtensions loads dynamic configuration like DATA_INPUTS into Extensions map
func LoadExtensions(filePath string, config *Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file for extensions: %w", err)
	}

	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return fmt.Errorf("failed to parse config file for extensions: %w", err)
	}

	// Load DATA_INPUTS array
	if dataInputsRaw, ok := rawConfig["DATA_INPUTS"]; ok {
		config.Extensions["DATA_INPUTS"] = dataInputsRaw
	}

	// Load any other dynamic extensions as needed
	for key, value := range rawConfig {
		// Skip known structured fields
		if !IsStructuredField(key) {
			config.Extensions[key] = value
		}
	}

	return nil
}

// IsStructuredField checks if a key belongs to structured config
func IsStructuredField(key string) bool {
	structuredPrefixes := []string{
		"APP_",
		"SPLUNK_",
		"SALESFORCE_",
		"MIGRATION_",
	}
	for _, prefix := range structuredPrefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

// BuildMap recursively builds a map for mapstructure from struct tags
func BuildMap(structType reflect.Type, values map[string]string) map[string]interface{} {
	dataMap := make(map[string]interface{})
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.Type.Kind() == reflect.Struct {
			nestedMap := BuildMap(field.Type, values)
			if len(nestedMap) > 0 {
				dataMap[field.Name] = nestedMap
			}
			continue
		}
		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}
		if value, found := values[envTag]; found {
			dataMap[field.Name] = value
		}
	}
	return dataMap
}

// StrToNumeric converts string to numeric types
func StrToNumeric(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String {
		return data, nil
	}

	s := data.(string)
	if s == "" {
		switch t.Kind() {
		case reflect.Int:
			return 0, nil
		case reflect.Int64:
			return int64(0), nil
		case reflect.Float64:
			return float64(0), nil
		case reflect.Bool:
			return false, nil
		}
		return data, nil
	}

	switch t.Kind() {
	case reflect.Int:
		return strconv.Atoi(s)
	case reflect.Int64:
		return strconv.ParseInt(s, 10, 64)
	case reflect.Float64:
		return strconv.ParseFloat(s, 64)
	}

	return data, nil
}

// StrToBool converts string to bool
func StrToBool(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t.Kind() != reflect.Bool {
		return data, nil
	}

	s := data.(string)
	if s == "" {
		return false, nil
	}

	return strconv.ParseBool(s)
}

// SetExtension safely sets an extension value
func (c *Config) SetExtension(key string, value interface{}) {
	if c.Extensions == nil {
		c.Extensions = make(map[string]interface{})
	}
	c.Extensions[key] = value
}

// GetDataInputs retrieves and parses DATA_INPUTS from extensions
func (c *Config) GetDataInputs() ([]DataInput, error) {
	dataInputsRaw, exists := c.Extensions["DATA_INPUTS"]
	if !exists {
		return nil, fmt.Errorf("DATA_INPUTS not found in configuration")
	}

	// Convert to []DataInput
	var inputs []DataInput
	if dataInputsArray, ok := dataInputsRaw.([]interface{}); ok {
		for i, inputRaw := range dataInputsArray {
			if inputMap, ok := inputRaw.(map[string]interface{}); ok {
				input := DataInput{
					Name:         getStringFromMap(inputMap, "name", ""),
					Object:       getStringFromMap(inputMap, "object", ""),
					ObjectFields: getStringFromMap(inputMap, "object_fields", ""),
					OrderBy:      getStringFromMap(inputMap, "order_by", "LastModifiedDate"),
					StartDate:    getStringFromMap(inputMap, "start_date", "2024-01-01T00:00:00.000Z"),
					Interval:     getIntFromMap(inputMap, "interval", 300),
					Delay:        getIntFromMap(inputMap, "delay", 60),
					Index:        getStringFromMap(inputMap, "index", ""),
				}

				// Use default index if not specified
				if input.Index == "" {
					input.Index = c.Splunk.DefaultIndex
				}

				if input.Name == "" || input.Object == "" {
					return nil, fmt.Errorf("data input [%d] missing required fields (name or object)", i)
				}

				inputs = append(inputs, input)
			} else {
				return nil, fmt.Errorf("data input [%d] is not a valid object", i)
			}
		}
	}

	return inputs, nil
}

// Helper functions for map extraction
func getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func getIntFromMap(m map[string]interface{}, key string, defaultValue int) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case string:
			if intVal, err := strconv.Atoi(v); err == nil {
				return intVal
			}
		}
	}
	return defaultValue
}

// Validate checks if all required configuration values are present
func (c *Config) Validate() error {
	// Validate Splunk configuration
	if c.Splunk.URL == "" {
		return fmt.Errorf("SPLUNK_URL is required")
	}
	if c.Splunk.Username == "" || c.Splunk.Password == "" {
		return fmt.Errorf("splunk authentication required: provide SPLUNK_USERNAME and SPLUNK_PASSWORD")
	}

	// Validate Salesforce configuration
	if c.Salesforce.Endpoint == "" {
		return fmt.Errorf("SALESFORCE_ENDPOINT is required")
	}
	if c.Salesforce.ClientID == "" {
		return fmt.Errorf("SALESFORCE_CLIENT_ID is required")
	}
	if c.Salesforce.ClientSecret == "" {
		return fmt.Errorf("SALESFORCE_CLIENT_SECRET is required")
	}
	if c.Salesforce.AccountName == "" {
		return fmt.Errorf("SALESFORCE_ACCOUNT_NAME is required")
	}

	// Validate Data Inputs
	dataInputs, err := c.GetDataInputs()
	if err != nil {
		return fmt.Errorf("failed to load data inputs: %w", err)
	}

	if len(dataInputs) == 0 {
		return fmt.Errorf("at least one data input configuration is required")
	}

	for i, input := range dataInputs {
		if input.Name == "" {
			return fmt.Errorf("data input [%d] name is required", i)
		}
		if input.Object == "" {
			return fmt.Errorf("data input [%d] object is required", i)
		}
		if input.ObjectFields == "" {
			return fmt.Errorf("data input [%d] object_fields is required", i)
		}
	}

	return nil
}
