// Package mocks provides mock implementations for testing
package mocks

import (
	"context"

	"salesforce-splunk-migration/utils"
)

// MockSplunkService is a mock implementation of SplunkServiceInterface
type MockSplunkService struct {
	AuthenticateFunc                 func(ctx context.Context) error
	GetAuthTokenFunc                 func() string
	CheckSalesforceAddonFunc         func(ctx context.Context) error
	CreateIndexFunc                  func(ctx context.Context, indexName string) error
	CheckIndexExistsFunc             func(ctx context.Context, indexName string) (bool, error)
	UpdateIndexFunc                  func(ctx context.Context, indexName string) error
	CreateSalesforceAccountFunc      func(ctx context.Context) error
	CheckSalesforceAccountExistsFunc func(ctx context.Context) (bool, error)
	UpdateSalesforceAccountFunc      func(ctx context.Context) error
	CreateDataInputFunc              func(ctx context.Context, input *utils.DataInput) error
	UpdateDataInputFunc              func(ctx context.Context, input *utils.DataInput) error
	CheckDataInputExistsFunc         func(ctx context.Context, inputName string) (bool, error)
	ListDataInputsFunc               func(ctx context.Context) ([]string, error)

	// Call tracking
	AuthenticateCalls                 int
	GetAuthTokenCalls                 int
	CheckSalesforceAddonCalls         int
	CreateIndexCalls                  int
	CheckIndexExistsCalls             int
	UpdateIndexCalls                  int
	CreateSalesforceAccountCalls      int
	CheckSalesforceAccountExistsCalls int
	UpdateSalesforceAccountCalls      int
	CreateDataInputCalls              int
	UpdateDataInputCalls              int
	CheckDataInputExistsCalls         int
	ListDataInputsCalls               int
}

// Authenticate mocks authentication
func (m *MockSplunkService) Authenticate(ctx context.Context) error {
	m.AuthenticateCalls++
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(ctx)
	}
	return nil
}

// GetAuthToken mocks getting the auth token
func (m *MockSplunkService) GetAuthToken() string {
	m.GetAuthTokenCalls++
	if m.GetAuthTokenFunc != nil {
		return m.GetAuthTokenFunc()
	}
	return "mock-token"
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

// CheckIndexExists mocks index existence check
func (m *MockSplunkService) CheckIndexExists(ctx context.Context, indexName string) (bool, error) {
	m.CheckIndexExistsCalls++
	if m.CheckIndexExistsFunc != nil {
		return m.CheckIndexExistsFunc(ctx, indexName)
	}
	return false, nil
}

// UpdateIndex mocks index update
func (m *MockSplunkService) UpdateIndex(ctx context.Context, indexName string) error {
	m.UpdateIndexCalls++
	if m.UpdateIndexFunc != nil {
		return m.UpdateIndexFunc(ctx, indexName)
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

// CheckSalesforceAccountExists mocks account existence check
func (m *MockSplunkService) CheckSalesforceAccountExists(ctx context.Context) (bool, error) {
	m.CheckSalesforceAccountExistsCalls++
	if m.CheckSalesforceAccountExistsFunc != nil {
		return m.CheckSalesforceAccountExistsFunc(ctx)
	}
	return false, nil
}

// UpdateSalesforceAccount mocks account update
func (m *MockSplunkService) UpdateSalesforceAccount(ctx context.Context) error {
	m.UpdateSalesforceAccountCalls++
	if m.UpdateSalesforceAccountFunc != nil {
		return m.UpdateSalesforceAccountFunc(ctx)
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

// UpdateDataInput mocks data input update
func (m *MockSplunkService) UpdateDataInput(ctx context.Context, input *utils.DataInput) error {
	m.UpdateDataInputCalls++
	if m.UpdateDataInputFunc != nil {
		return m.UpdateDataInputFunc(ctx, input)
	}
	return nil
}

// CheckDataInputExists mocks checking if data input exists
func (m *MockSplunkService) CheckDataInputExists(ctx context.Context, inputName string) (bool, error) {
	m.CheckDataInputExistsCalls++
	if m.CheckDataInputExistsFunc != nil {
		return m.CheckDataInputExistsFunc(ctx, inputName)
	}
	return false, nil
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
	m.UpdateSalesforceAccountCalls = 0
	m.CreateDataInputCalls = 0
	m.UpdateDataInputCalls = 0
	m.CheckDataInputExistsCalls = 0
	m.ListDataInputsCalls = 0
}
