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
	httpClient *utils.HTTPClient
	authToken  string
}

// NewSplunkService creates a new Splunk service instance with connection pooling
func NewSplunkService(config *utils.Config) (*SplunkService, error) {
	// Create HTTP client with connection pooling and retry configuration
	httpClient := utils.NewHTTPClient(utils.HTTPClientConfig{
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

	return &SplunkService{
		config:     config,
		httpClient: httpClient,
		authToken:  config.Splunk.AuthToken,
	}, nil
}

// Authenticate authenticates with Splunk and obtains a session token
func (s *SplunkService) Authenticate() error {
	// If auth token is already provided, use it
	if s.authToken != "" {
		fmt.Println("Using provided auth token")
		return nil
	}

	// Otherwise, create a session token using username/password
	ctx := context.Background()

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

	var authResp models.AuthResponse
	if err := resp.JSON(&authResp); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	s.authToken = authResp.SessionKey
	return nil
}

// CreateIndex creates a new Splunk index
func (s *SplunkService) CreateIndex(indexName string) error {
	formData := map[string]string{
		"name":        indexName,
		"datatype":    "event",
		"output_mode": "json",
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Splunk %s", s.authToken),
	}

	ctx := context.Background()
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

// VerifyIndexExists checks if the specified index exists in Splunk
func (s *SplunkService) VerifyIndexExists(indexName string) error {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Splunk %s", s.authToken),
	}

	ctx := context.Background()
	path := fmt.Sprintf("/services/data/indexes/%s?output_mode=json", indexName)
	resp, err := s.httpClient.Get(ctx, path, headers)
	if err != nil {
		return fmt.Errorf("failed to verify index: %w", err)
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("index '%s' does not exist", indexName)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to verify index: status %d - %s", resp.StatusCode, resp.String())
	}

	return nil
}

// VerifyAccountExists checks if the specified Salesforce account exists in Splunk
func (s *SplunkService) VerifyAccountExists(accountName string) error {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Splunk %s", s.authToken),
	}

	ctx := context.Background()
	path := fmt.Sprintf("/servicesNS/-/Splunk_TA_salesforce/configs/conf-sfdc_connection/%s?output_mode=json", accountName)
	resp, err := s.httpClient.Get(ctx, path, headers)
	if err != nil {
		return fmt.Errorf("failed to verify account: %w", err)
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("salesforce account '%s' does not exist", accountName)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to verify account: status %d - %s", resp.StatusCode, resp.String())
	}

	return nil
}

// CreateSalesforceAccount creates a Salesforce account in Splunk
func (s *SplunkService) CreateSalesforceAccount() error {
	formData := map[string]string{
		"name":             s.config.Salesforce.AccountName,
		"endpoint":         s.config.Salesforce.Endpoint,
		"sfdc_api_version": s.config.Salesforce.APIVersion,
		"auth_type":        s.config.Salesforce.AuthType,
		"output_mode":      "json",
	}

	// Add auth-specific parameters
	if s.config.Salesforce.AuthType == "oauth_client_credentials" {
		formData["client_id_oauth_credentials"] = s.config.Salesforce.ClientID
		formData["client_secret_oauth_credentials"] = s.config.Salesforce.ClientSecret
	} else {
		formData["username"] = s.config.Salesforce.Username
		formData["password"] = s.config.Salesforce.Password
		if s.config.Salesforce.SecurityToken != "" {
			formData["token"] = s.config.Salesforce.SecurityToken
		}
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Splunk %s", s.authToken),
	}

	ctx := context.Background()
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
func (s *SplunkService) CreateDataInput(input *utils.DataInput) error {
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

	ctx := context.Background()
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
func (s *SplunkService) ListDataInputs() ([]string, error) {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Splunk %s", s.authToken),
	}

	ctx := context.Background()
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

// DeleteDataInput deletes a Salesforce object data input
func (s *SplunkService) DeleteDataInput(name string) error {
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Splunk %s", s.authToken),
	}

	ctx := context.Background()
	path := fmt.Sprintf("/servicesNS/-/Splunk_TA_salesforce/Splunk_TA_salesforce_sfdc_object/%s", name)
	resp, err := s.httpClient.Delete(ctx, path, headers)
	if err != nil {
		return fmt.Errorf("failed to delete data input: %w", err)
	}

	if !resp.IsSuccess() {
		return fmt.Errorf("failed to delete data input: status %d - %s", resp.StatusCode, resp.String())
	}

	return nil
}
