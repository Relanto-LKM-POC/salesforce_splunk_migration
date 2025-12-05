// Package services implements business logic for Splunk API interactions
package services

import (
	"context"
	"fmt"
	"time"

	"salesforce-splunk-migration/models"
	"salesforce-splunk-migration/utils"
)

// SplunkService handles all Splunk API operations
type SplunkService struct {
	config     *utils.Config
	httpClient utils.HTTPClientInterface
	authToken  string
}

// NewSplunkService creates a new Splunk service instance with connection pooling
func NewSplunkService(config *utils.Config) (*SplunkService, error) {
	return NewSplunkServiceWithClient(config, nil)
}

// NewSplunkServiceWithClient creates a new Splunk service with custom HTTP client (for testing)
func NewSplunkServiceWithClient(config *utils.Config, httpClient utils.HTTPClientInterface) (*SplunkService, error) {
	if httpClient == nil {
		// Create HTTP client with connection pooling and retry configuration
		httpClient = utils.NewHTTPClient(utils.HTTPClientConfig{
			BaseURL: config.Splunk.URL,
			Timeout: time.Duration(config.Splunk.RequestTimeout) * time.Second,
			Headers: map[string]string{
				"User-Agent": "Salesforce-Splunk-Migration/1.0",
			},
			RetryConfig: utils.RetryConfig{
				MaxRetries: config.Splunk.MaxRetries,
				RetryDelay: time.Duration(config.Splunk.RetryDelay) * time.Second,
				BackoffExp: 2.0,
			},
			SkipSSLVerify:   config.Splunk.SkipSSLVerify,
			MaxIdleConns:    100,
			MaxConnsPerHost: 100,
		})
	}

	return &SplunkService{
		config:     config,
		httpClient: httpClient,
	}, nil
}

// Authenticate authenticates with Splunk and obtains a session token
func (s *SplunkService) Authenticate(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	formData := map[string]string{
		"username":    s.config.Splunk.Username,
		"password":    s.config.Splunk.Password,
		"output_mode": "json",
	}

	resp, err := s.httpClient.PostForm(ctx, "/services/auth/login", formData, nil)
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, resp.String())
	}

	// Parse response to extract session key
	var authResp models.AuthResponse

	if err := resp.JSON(&authResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	s.authToken = authResp.SessionKey
	return nil
}

// CheckSalesforceAddon checks if Splunk Add-on for Salesforce is installed
func (s *SplunkService) CheckSalesforceAddon(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Splunk %s", s.authToken),
	}

	// List all installed apps
	resp, err := s.httpClient.Get(ctx, "/services/apps/local?output_mode=json", headers)
	if err != nil {
		return fmt.Errorf("failed to list installed apps: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to list installed apps: status %d - %s", resp.StatusCode, resp.String())
	}

	var result struct {
		Entry []struct {
			Name    string `json:"name"`
			Content struct {
				Label    string `json:"label"`
				Version  string `json:"version"`
				Disabled bool   `json:"disabled"`
			} `json:"content"`
		} `json:"entry"`
	}

	if err := resp.JSON(&result); err != nil {
		return fmt.Errorf("failed to parse apps list: %w", err)
	}

	// Check if Splunk_TA_salesforce is installed and enabled
	for _, entry := range result.Entry {
		if entry.Name == "Splunk_TA_salesforce" {
			if entry.Content.Disabled {
				return fmt.Errorf("Splunk Add-on for Salesforce is installed but disabled (version: %s)", entry.Content.Version)
			}
			return nil // App found and enabled
		}
	}

	return fmt.Errorf("Splunk Add-on for Salesforce (Splunk_TA_salesforce) is not installed. Please install it from Splunkbase before proceeding")
}

// CreateIndex creates a new Splunk index
func (s *SplunkService) CreateIndex(ctx context.Context, indexName string) error {
	if indexName == "" {
		return fmt.Errorf("index name cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	formData := map[string]string{
		"name":        indexName,
		"datatype":    "event",
		"output_mode": "json",
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Splunk %s", s.authToken),
	}

	resp, err := s.httpClient.PostForm(ctx, "/services/data/indexes", formData, headers)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// The communication layer already handles 409 and 500 "already exists" responses
	if !resp.IsSuccess() && resp.StatusCode != 409 && resp.StatusCode != 500 {
		return fmt.Errorf("failed to create index: status %d - %s", resp.StatusCode, resp.String())
	}

	return s.checkResponseMessages(resp)
}

// CreateSalesforceAccount creates a Salesforce account in Splunk
func (s *SplunkService) CreateSalesforceAccount(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	formData := map[string]string{
		"name":             s.config.Salesforce.AccountName,
		"endpoint":         s.config.Salesforce.Endpoint,
		"sfdc_api_version": s.config.Salesforce.APIVersion,
		"auth_type":        s.config.Salesforce.AuthType,
		"output_mode":      "json",
	}

	// Add OAuth client credentials
	formData["client_id_oauth_credentials"] = s.config.Salesforce.ClientID
	formData["client_secret_oauth_credentials"] = s.config.Salesforce.ClientSecret

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Splunk %s", s.authToken),
	}

	resp, err := s.httpClient.PostForm(ctx, "/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account", formData, headers)
	if err != nil {
		return fmt.Errorf("failed to create Salesforce account: %w", err)
	}

	// The communication layer already handles 409 and 500 "already exists" responses
	// Just check if response is successful or handled
	if !resp.IsSuccess() && resp.StatusCode != 409 && resp.StatusCode != 500 {
		return fmt.Errorf("failed to create Salesforce account: status %d - %s", resp.StatusCode, resp.String())
	}

	return s.checkResponseMessages(resp)
}

// CreateDataInput creates a Salesforce object data input in Splunk
func (s *SplunkService) CreateDataInput(ctx context.Context, input *utils.DataInput) error {
	if input == nil {
		return fmt.Errorf("data input cannot be nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	formData := map[string]string{
		"name":          input.Name,
		"account":       s.config.Salesforce.AccountName,
		"object":        input.Object,
		"object_fields": input.ObjectFields,
		"order_by":      input.OrderBy,
		"start_date":    input.StartDate,
		"interval":      fmt.Sprintf("%d", input.Interval),
		"delay":         fmt.Sprintf("%d", input.Delay),
		"index":         input.Index,
		"output_mode":   "json",
	}

	// Use default index if not specified
	if formData["index"] == "" {
		formData["index"] = s.config.Splunk.DefaultIndex
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Splunk %s", s.authToken),
	}

	resp, err := s.httpClient.PostForm(ctx, "/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object", formData, headers)
	if err != nil {
		return fmt.Errorf("failed to create data input: %w", err)
	}

	// The communication layer already handles 409 and 500 "already exists" responses
	if !resp.IsSuccess() && resp.StatusCode != 409 && resp.StatusCode != 500 {
		return fmt.Errorf("failed to create data input: status %d - %s", resp.StatusCode, resp.String())
	}

	return s.checkResponseMessages(resp)
}

// checkResponseMessages checks Splunk API response for messages and errors
func (s *SplunkService) checkResponseMessages(resp *utils.HTTPResponse) error {
	var splunkResp models.SplunkResponse
	if err := resp.JSON(&splunkResp); err != nil {
		// If we can't parse as JSON, but request was successful, just return
		if resp.IsSuccess() {
			return nil
		}
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error messages
	if len(splunkResp.Messages) > 0 {
		for _, msg := range splunkResp.Messages {
			if msg.Type == "ERROR" {
				// Don't treat "already exists" as error
				return nil
			}
		}
	}

	return nil
}

// ListDataInputs lists all existing Salesforce object data inputs
func (s *SplunkService) ListDataInputs(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Splunk %s", s.authToken),
	}

	resp, err := s.httpClient.Get(ctx, "/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object?output_mode=json", headers)
	if err != nil {
		return nil, fmt.Errorf("failed to list data inputs: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("failed to list data inputs: status %d - %s", resp.StatusCode, resp.String())
	}

	var result struct {
		Entry []struct {
			Name string `json:"name"`
		} `json:"entry"`
	}

	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var names []string
	for _, entry := range result.Entry {
		names = append(names, entry.Name)
	}

	return names, nil
}
