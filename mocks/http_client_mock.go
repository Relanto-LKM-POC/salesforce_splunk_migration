// Package mocks provides mock implementations for testing
package mocks

import (
	"context"

	"salesforce-splunk-migration/utils"
)

// MockHTTPClient is a mock implementation of HTTPClientInterface
type MockHTTPClient struct {
	GetFunc      func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error)
	PostFunc     func(ctx context.Context, path string, body interface{}, headers map[string]string) (*utils.HTTPResponse, error)
	PostFormFunc func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error)
	PutFunc      func(ctx context.Context, path string, body interface{}, headers map[string]string) (*utils.HTTPResponse, error)
	DeleteFunc   func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error)

	// Call tracking
	GetCalls      int
	PostCalls     int
	PostFormCalls int
	PutCalls      int
	DeleteCalls   int
}

// Get mocks GET request
func (m *MockHTTPClient) Get(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
	m.GetCalls++
	if m.GetFunc != nil {
		return m.GetFunc(ctx, path, headers)
	}
	return &utils.HTTPResponse{StatusCode: 200, Body: []byte("{}")}, nil
}

// Post mocks POST request
func (m *MockHTTPClient) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*utils.HTTPResponse, error) {
	m.PostCalls++
	if m.PostFunc != nil {
		return m.PostFunc(ctx, path, body, headers)
	}
	return &utils.HTTPResponse{StatusCode: 200, Body: []byte("{}")}, nil
}

// PostForm mocks POST form request
func (m *MockHTTPClient) PostForm(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
	m.PostFormCalls++
	if m.PostFormFunc != nil {
		return m.PostFormFunc(ctx, path, formData, headers)
	}
	return &utils.HTTPResponse{StatusCode: 200, Body: []byte("{}")}, nil
}

// Put mocks PUT request
func (m *MockHTTPClient) Put(ctx context.Context, path string, body interface{}, headers map[string]string) (*utils.HTTPResponse, error) {
	m.PutCalls++
	if m.PutFunc != nil {
		return m.PutFunc(ctx, path, body, headers)
	}
	return &utils.HTTPResponse{StatusCode: 200, Body: []byte("{}")}, nil
}

// Delete mocks DELETE request
func (m *MockHTTPClient) Delete(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
	m.DeleteCalls++
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, path, headers)
	}
	return &utils.HTTPResponse{StatusCode: 200, Body: []byte("{}")}, nil
}

// Reset resets all call counters
func (m *MockHTTPClient) Reset() {
	m.GetCalls = 0
	m.PostCalls = 0
	m.PostFormCalls = 0
	m.PutCalls = 0
	m.DeleteCalls = 0
}
