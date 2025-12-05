// Package mocks provides mock implementations for testing
package mocks

import (
	"context"

	"salesforce-splunk-migration/utils"
)

// MockSplunkService is a mock implementation of SplunkServiceInterface
type MockSplunkService struct {
	AuthenticateFunc            func(ctx context.Context) error
	CheckSalesforceAddonFunc    func(ctx context.Context) error
	CreateIndexFunc             func(ctx context.Context, indexName string) error
	CreateSalesforceAccountFunc func(ctx context.Context) error
	CreateDataInputFunc         func(ctx context.Context, input *utils.DataInput) error
	ListDataInputsFunc          func(ctx context.Context) ([]string, error)

	// Call tracking
	AuthenticateCalls            int
	CheckSalesforceAddonCalls    int
	CreateIndexCalls             int
	CreateSalesforceAccountCalls int
	CreateDataInputCalls         int
	ListDataInputsCalls          int
}

// Authenticate mocks authentication
func (m *MockSplunkService) Authenticate(ctx context.Context) error {
	m.AuthenticateCalls++
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(ctx)
	}
	return nil
}

// CheckSalesforceAddon mocks addon check
func (m *MockSplunkService) CheckSalesforceAddon(ctx context.Context) error {
	m.CheckSalesforceAddonCalls++
	if m.CheckSalesforceAddonFunc != nil {
		return m.CheckSalesforceAddonFunc(ctx)
	}
	return nil
}

// CreateIndex mocks index creation
func (m *MockSplunkService) CreateIndex(ctx context.Context, indexName string) error {
	m.CreateIndexCalls++
	if m.CreateIndexFunc != nil {
		return m.CreateIndexFunc(ctx, indexName)
	}
	return nil
}

// CreateSalesforceAccount mocks account creation
func (m *MockSplunkService) CreateSalesforceAccount(ctx context.Context) error {
	m.CreateSalesforceAccountCalls++
	if m.CreateSalesforceAccountFunc != nil {
		return m.CreateSalesforceAccountFunc(ctx)
	}
	return nil
}

// CreateDataInput mocks data input creation
func (m *MockSplunkService) CreateDataInput(ctx context.Context, input *utils.DataInput) error {
	m.CreateDataInputCalls++
	if m.CreateDataInputFunc != nil {
		return m.CreateDataInputFunc(ctx, input)
	}
	return nil
}

// ListDataInputs mocks listing data inputs
func (m *MockSplunkService) ListDataInputs(ctx context.Context) ([]string, error) {
	m.ListDataInputsCalls++
	if m.ListDataInputsFunc != nil {
		return m.ListDataInputsFunc(ctx)
	}
	return []string{}, nil
}

// Reset resets all call counters
func (m *MockSplunkService) Reset() {
	m.AuthenticateCalls = 0
	m.CheckSalesforceAddonCalls = 0
	m.CreateIndexCalls = 0
	m.CreateSalesforceAccountCalls = 0
	m.CreateDataInputCalls = 0
	m.ListDataInputsCalls = 0
}
