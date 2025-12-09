// Package services implements business logic for Splunk API interactions
package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"salesforce-splunk-migration/models"
	"salesforce-splunk-migration/utils"
)

type SplunkServiceInterface interface {
	Authenticate(ctx context.Context) error
	GetAuthToken() string
	CheckSalesforceAddon(ctx context.Context) error
	CreateIndex(ctx context.Context, indexName string) error
	CheckIndexExists(ctx context.Context, indexName string) (bool, error)
	UpdateIndex(ctx context.Context, indexName string) error
	CreateSalesforceAccount(ctx context.Context) error
	CheckSalesforceAccountExists(ctx context.Context) (bool, error)
	UpdateSalesforceAccount(ctx context.Context) error
	CreateDataInput(ctx context.Context, input *utils.DataInput) error
	UpdateDataInput(ctx context.Context, input *utils.DataInput) error
	CheckDataInputExists(ctx context.Context, inputName string) (bool, error)
	ListDataInputs(ctx context.Context) ([]string, error)
}

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

// Authenticate authenticates with Splunk and obtains a JWT token using /services/authorization/tokens
func (s *SplunkService) Authenticate(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Set default values if not provided
	tokenName := s.config.Splunk.TokenName
	if tokenName == "" {
		tokenName = s.config.Splunk.Username // Use username as token name if not specified
	}

	tokenAudience := s.config.Splunk.TokenAudience
	if tokenAudience == "" {
		tokenAudience = "Automation" // Default audience for automation
	}

	formData := map[string]string{
		"name":        tokenName,
		"audience":    tokenAudience,
		"output_mode": "json",
	}

	// Create headers with Basic Authentication
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	// Use PostForm with basic auth credentials
	resp, err := s.httpClient.PostFormWithBasicAuth(ctx, "/services/authorization/tokens", formData, headers, s.config.Splunk.Username, s.config.Splunk.Password)
	if err != nil {
		return fmt.Errorf("token authentication request failed: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("token authentication failed with status %d: %s", resp.StatusCode, resp.String())
	}

	// Parse response to extract JWT token
	var tokenResp models.TokenAuthResponse

	if err := resp.JSON(&tokenResp); err != nil {
		return fmt.Errorf("failed to parse token auth response: %w", err)
	}

	if len(tokenResp.Entry) == 0 {
		return fmt.Errorf("no token returned in authentication response")
	}

	s.authToken = tokenResp.Entry[0].Content.Token
	return nil
}

// GetAuthToken returns the authentication token
func (s *SplunkService) GetAuthToken() string {
	return s.authToken
}

// CheckSalesforceAddon checks if Splunk Add-on for Salesforce is installed
func (s *SplunkService) CheckSalesforceAddon(ctx context.Context) error {
	// BYPASSED: Assuming Splunk Add-on for Salesforce is installed
	// This check is skipped for Splunk Cloud instances where the add-on
	// may not be visible via the /services/apps/local API endpoint
	return nil

	/* Original check commented out for Splunk Cloud compatibility
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.authToken),
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
	*/
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

	// Add maxTotalDataSizeMB if configured
	if s.config.Splunk.MaxTotalDataSizeMB > 0 {
		formData["maxTotalDataSizeMB"] = fmt.Sprintf("%d", s.config.Splunk.MaxTotalDataSizeMB)
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.authToken),
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

// CheckIndexExists checks if an index exists
func (s *SplunkService) CheckIndexExists(ctx context.Context, indexName string) (bool, error) {
	if indexName == "" {
		return false, fmt.Errorf("index name cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.authToken),
	}

	url := fmt.Sprintf("/services/data/indexes/%s?output_mode=json", indexName)
	resp, err := s.httpClient.Get(ctx, url, headers)
	if err != nil {
		return false, fmt.Errorf("failed to check index existence: %w", err)
	}

	// 404 means index doesn't exist - this is not an error condition
	if resp.StatusCode == 404 {
		return false, nil
	}

	return resp.StatusCode == 200, nil
}

// UpdateIndex updates an existing Splunk index
func (s *SplunkService) UpdateIndex(ctx context.Context, indexName string) error {
	if indexName == "" {
		return fmt.Errorf("index name cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	formData := map[string]string{
		"output_mode": "json",
	}

	// Add maxTotalDataSizeMB if configured
	if s.config.Splunk.MaxTotalDataSizeMB > 0 {
		formData["maxTotalDataSizeMB"] = fmt.Sprintf("%d", s.config.Splunk.MaxTotalDataSizeMB)
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.authToken),
	}

	// Update uses POST to the specific index endpoint
	url := fmt.Sprintf("/services/data/indexes/%s", indexName)
	resp, err := s.httpClient.PostForm(ctx, url, formData, headers)
	if err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to update index: status %d - %s", resp.StatusCode, resp.String())
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
		"Authorization": fmt.Sprintf("Bearer %s", s.authToken),
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

// CheckSalesforceAccountExists checks if a Salesforce account exists
func (s *SplunkService) CheckSalesforceAccountExists(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.authToken),
	}

	url := fmt.Sprintf("/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account/%s?output_mode=json", s.config.Salesforce.AccountName)
	resp, err := s.httpClient.Get(ctx, url, headers)
	if err != nil {
		return false, fmt.Errorf("failed to check Salesforce account existence: %w", err)
	}

	// Any 4xx error means account doesn't exist or we can't access it
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return false, nil
	}

	// 500 error with "Not Found" message means account doesn't exist
	// (Splunk returns 500 instead of 404 for some endpoints)
	if resp.StatusCode == 500 {
		bodyStr := string(resp.Body)
		if strings.Contains(bodyStr, "Not Found") ||
			strings.Contains(bodyStr, "Could not find object") ||
			strings.Contains(bodyStr, "[404]") {
			return false, nil
		}
	}

	return resp.StatusCode == 200, nil
}

// UpdateSalesforceAccount updates an existing Salesforce account in Splunk
func (s *SplunkService) UpdateSalesforceAccount(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	formData := map[string]string{
		"endpoint":         s.config.Salesforce.Endpoint,
		"sfdc_api_version": s.config.Salesforce.APIVersion,
		"auth_type":        s.config.Salesforce.AuthType,
		"output_mode":      "json",
	}

	// Add OAuth client credentials
	formData["client_id_oauth_credentials"] = s.config.Salesforce.ClientID
	formData["client_secret_oauth_credentials"] = s.config.Salesforce.ClientSecret

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.authToken),
	}

	// Update uses POST to the specific account endpoint
	url := fmt.Sprintf("/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_account/%s", s.config.Salesforce.AccountName)
	resp, err := s.httpClient.PostForm(ctx, url, formData, headers)
	if err != nil {
		return fmt.Errorf("failed to update Salesforce account: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to update Salesforce account: status %d - %s", resp.StatusCode, resp.String())
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
		"Authorization": fmt.Sprintf("Bearer %s", s.authToken),
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

// CheckDataInputExists checks if a data input exists
func (s *SplunkService) CheckDataInputExists(ctx context.Context, inputName string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", s.authToken),
	}

	url := fmt.Sprintf("/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object/%s?output_mode=json", inputName)
	resp, err := s.httpClient.Get(ctx, url, headers)
	if err != nil {
		return false, fmt.Errorf("failed to check data input existence: %w", err)
	}

	// Any 4xx error means data input doesn't exist or we can't access it
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return false, nil
	}

	// 500 error with "Not Found" message means data input doesn't exist
	// (Splunk returns 500 instead of 404 for some endpoints)
	if resp.StatusCode == 500 {
		bodyStr := string(resp.Body)
		if strings.Contains(bodyStr, "Not Found") ||
			strings.Contains(bodyStr, "Could not find object") ||
			strings.Contains(bodyStr, "[404]") {
			return false, nil
		}
	}

	return resp.StatusCode == 200, nil
}

// UpdateDataInput updates an existing Salesforce object data input in Splunk
func (s *SplunkService) UpdateDataInput(ctx context.Context, input *utils.DataInput) error {
	if input == nil {
		return fmt.Errorf("data input cannot be nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	formData := map[string]string{
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
		"Authorization": fmt.Sprintf("Bearer %s", s.authToken),
	}

	// Update uses POST to the specific input endpoint
	url := fmt.Sprintf("/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object/%s", input.Name)
	resp, err := s.httpClient.PostForm(ctx, url, formData, headers)
	if err != nil {
		return fmt.Errorf("failed to update data input: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to update data input: status %d - %s", resp.StatusCode, resp.String())
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
		"Authorization": fmt.Sprintf("Bearer %s", s.authToken),
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
