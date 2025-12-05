// Package models contains request/response data structures
package models

import (
	"fmt"
	"strings"
	"time"
)

// AuthRequest represents a Splunk authentication request
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Validate validates the AuthRequest
func (r *AuthRequest) Validate() error {
	if strings.TrimSpace(r.Username) == "" {
		return fmt.Errorf("username is required")
	}
	if strings.TrimSpace(r.Password) == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

// IndexRequest represents a Splunk index creation request
type IndexRequest struct {
	Name     string `json:"name"`
	DataType string `json:"datatype"`
}

// Validate validates the IndexRequest
func (r *IndexRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("index name is required")
	}
	if len(r.Name) > 255 {
		return fmt.Errorf("index name must be 255 characters or less")
	}
	if r.DataType != "" && r.DataType != "event" && r.DataType != "metric" {
		return fmt.Errorf("datatype must be 'event' or 'metric'")
	}
	return nil
}

// SalesforceAccountRequest represents a Salesforce account creation request
type SalesforceAccountRequest struct {
	Name                         string `json:"name"`
	Endpoint                     string `json:"endpoint"`
	SFDCAPIVersion               string `json:"sfdc_api_version"`
	AuthType                     string `json:"auth_type"`
	Username                     string `json:"username,omitempty"`
	Password                     string `json:"password,omitempty"`
	Token                        string `json:"token,omitempty"`
	ClientIDOAuthCredentials     string `json:"client_id_oauth_credentials,omitempty"`
	ClientSecretOAuthCredentials string `json:"client_secret_oauth_credentials,omitempty"`
}

// Validate validates the SalesforceAccountRequest
func (r *SalesforceAccountRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("account name is required")
	}
	if strings.TrimSpace(r.Endpoint) == "" {
		return fmt.Errorf("endpoint is required")
	}
	if !strings.HasPrefix(r.Endpoint, "https://") {
		return fmt.Errorf("endpoint must start with https://")
	}
	if strings.TrimSpace(r.AuthType) == "" {
		return fmt.Errorf("auth_type is required")
	}
	validAuthTypes := map[string]bool{"basic": true, "oauth2": true}
	if !validAuthTypes[r.AuthType] {
		return fmt.Errorf("auth_type must be 'basic' or 'oauth2'")
	}

	// Validate credentials based on auth type
	if r.AuthType == "oauth2" {
		if strings.TrimSpace(r.ClientIDOAuthCredentials) == "" {
			return fmt.Errorf("client_id is required for oauth2 auth")
		}
		if strings.TrimSpace(r.ClientSecretOAuthCredentials) == "" {
			return fmt.Errorf("client_secret is required for oauth2 auth")
		}
	}
	return nil
}

// DataInputRequest represents a Salesforce object data input request
type DataInputRequest struct {
	Name         string `json:"name"`
	Account      string `json:"account"`
	Object       string `json:"object"`
	ObjectFields string `json:"object_fields"`
	OrderBy      string `json:"order_by"`
	StartDate    string `json:"start_date"`
	Interval     int    `json:"interval"`
	Delay        int    `json:"delay"`
	Index        string `json:"index"`
}

// Validate validates the DataInputRequest
func (r *DataInputRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("data input name is required")
	}
	if strings.TrimSpace(r.Account) == "" {
		return fmt.Errorf("account is required")
	}
	if strings.TrimSpace(r.Object) == "" {
		return fmt.Errorf("object is required")
	}
	if r.Interval < 0 {
		return fmt.Errorf("interval must be non-negative")
	}
	if r.Delay < 0 {
		return fmt.Errorf("delay must be non-negative")
	}
	// Validate start date format if provided
	if r.StartDate != "" {
		_, err := time.Parse("2006-01-02", r.StartDate)
		if err != nil {
			return fmt.Errorf("start_date must be in YYYY-MM-DD format: %w", err)
		}
	}
	return nil
}
