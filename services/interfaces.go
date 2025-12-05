// Package services defines interfaces for dependency injection
package services

import (
	"context"

	"salesforce-splunk-migration/utils"
)

// SplunkServiceInterface defines the contract for Splunk service operations
type SplunkServiceInterface interface {
	// Authenticate authenticates with Splunk and obtains a session token
	Authenticate(ctx context.Context) error

	// CheckSalesforceAddon checks if Splunk Add-on for Salesforce is installed
	CheckSalesforceAddon(ctx context.Context) error

	// CreateIndex creates a new Splunk index
	CreateIndex(ctx context.Context, indexName string) error

	// CreateSalesforceAccount creates a Salesforce account in Splunk
	CreateSalesforceAccount(ctx context.Context) error

	// CreateDataInput creates a Salesforce object data input in Splunk
	CreateDataInput(ctx context.Context, input *utils.DataInput) error

	// ListDataInputs lists all existing Salesforce object data inputs
	ListDataInputs(ctx context.Context) ([]string, error)
}

// HTTPClientInterface defines the contract for HTTP client operations
type HTTPClientInterface interface {
	// Get performs a GET request
	Get(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error)

	// Post performs a POST request with JSON body
	Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*utils.HTTPResponse, error)

	// PostForm performs a POST request with form data
	PostForm(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error)

	// Put performs a PUT request
	Put(ctx context.Context, path string, body interface{}, headers map[string]string) (*utils.HTTPResponse, error)

	// Delete performs a DELETE request
	Delete(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error)
}
