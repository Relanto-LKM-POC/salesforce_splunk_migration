// Package mocks provides mock implementations for testing
package mocks_test

import (
	"context"
	"errors"
	"testing"

	"salesforce-splunk-migration/mocks"
	"salesforce-splunk-migration/utils"
)

func TestMockHTTPClient_Get(t *testing.T) {
	t.Run("Success_DefaultBehavior", func(t *testing.T) {
		mock := &mocks.MockHTTPClient{}
		ctx := context.Background()

		resp, err := mock.Get(ctx, "/test", nil)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}

		if string(resp.Body) != "{}" {
			t.Errorf("Expected body '{}', got '%s'", string(resp.Body))
		}

		if mock.GetCalls != 1 {
			t.Errorf("Expected 1 Get call, got %d", mock.GetCalls)
		}
	})

	t.Run("Success_CustomFunction", func(t *testing.T) {
		mock := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 201,
					Body:       []byte(`{"custom": "response"}`),
				}, nil
			},
		}
		ctx := context.Background()

		resp, err := mock.Get(ctx, "/custom", nil)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		if resp.StatusCode != 201 {
			t.Errorf("Expected status code 201, got %d", resp.StatusCode)
		}

		if string(resp.Body) != `{"custom": "response"}` {
			t.Errorf("Expected custom response, got '%s'", string(resp.Body))
		}

		if mock.GetCalls != 1 {
			t.Errorf("Expected 1 Get call, got %d", mock.GetCalls)
		}
	})

	t.Run("Error_CustomError", func(t *testing.T) {
		expectedErr := errors.New("custom error")
		mock := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, expectedErr
			},
		}
		ctx := context.Background()

		_, err := mock.Get(ctx, "/error", nil)
		if err != expectedErr {
			t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
		}

		if mock.GetCalls != 1 {
			t.Errorf("Expected 1 Get call, got %d", mock.GetCalls)
		}
	})

	t.Run("Success_WithHeaders", func(t *testing.T) {
		var capturedHeaders map[string]string
		mock := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				capturedHeaders = headers
				return &utils.HTTPResponse{StatusCode: 200, Body: []byte("{}")}, nil
			},
		}
		ctx := context.Background()

		headers := map[string]string{
			"Authorization": "Bearer token",
			"X-Custom":      "value",
		}

		_, err := mock.Get(ctx, "/test", headers)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		if capturedHeaders["Authorization"] != "Bearer token" {
			t.Error("Headers not passed correctly")
		}
		if capturedHeaders["X-Custom"] != "value" {
			t.Error("Custom header not passed correctly")
		}
	})

	t.Run("Success_MultipleCalls", func(t *testing.T) {
		mock := &mocks.MockHTTPClient{}
		ctx := context.Background()

		for i := 0; i < 5; i++ {
			_, err := mock.Get(ctx, "/test", nil)
			if err != nil {
				t.Fatalf("Get() call %d failed: %v", i+1, err)
			}
		}

		if mock.GetCalls != 5 {
			t.Errorf("Expected 5 Get calls, got %d", mock.GetCalls)
		}
	})
}

func TestMockHTTPClient_Post(t *testing.T) {
	t.Run("Success_DefaultBehavior", func(t *testing.T) {
		mock := &mocks.MockHTTPClient{}
		ctx := context.Background()

		body := map[string]interface{}{
			"key": "value",
		}

		resp, err := mock.Post(ctx, "/test", body, nil)
		if err != nil {
			t.Fatalf("Post() failed: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}

		if mock.PostCalls != 1 {
			t.Errorf("Expected 1 Post call, got %d", mock.PostCalls)
		}
	})

	t.Run("Success_CustomFunction", func(t *testing.T) {
		var capturedBody interface{}
		mock := &mocks.MockHTTPClient{
			PostFunc: func(ctx context.Context, path string, body interface{}, headers map[string]string) (*utils.HTTPResponse, error) {
				capturedBody = body
				return &utils.HTTPResponse{
					StatusCode: 201,
					Body:       []byte(`{"created": true}`),
				}, nil
			},
		}
		ctx := context.Background()

		body := map[string]interface{}{
			"name": "test",
			"id":   123,
		}

		resp, err := mock.Post(ctx, "/create", body, nil)
		if err != nil {
			t.Fatalf("Post() failed: %v", err)
		}

		if resp.StatusCode != 201 {
			t.Errorf("Expected status code 201, got %d", resp.StatusCode)
		}

		if capturedBody == nil {
			t.Error("Body not captured")
		}

		if mock.PostCalls != 1 {
			t.Errorf("Expected 1 Post call, got %d", mock.PostCalls)
		}
	})

	t.Run("Error_CustomError", func(t *testing.T) {
		expectedErr := errors.New("post error")
		mock := &mocks.MockHTTPClient{
			PostFunc: func(ctx context.Context, path string, body interface{}, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, expectedErr
			},
		}
		ctx := context.Background()

		_, err := mock.Post(ctx, "/test", nil, nil)
		if err != expectedErr {
			t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
		}
	})

	t.Run("Success_NilBody", func(t *testing.T) {
		mock := &mocks.MockHTTPClient{}
		ctx := context.Background()

		resp, err := mock.Post(ctx, "/test", nil, nil)
		if err != nil {
			t.Fatalf("Post() with nil body failed: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}
	})
}

func TestMockHTTPClient_PostForm(t *testing.T) {
	t.Run("Success_DefaultBehavior", func(t *testing.T) {
		mock := &mocks.MockHTTPClient{}
		ctx := context.Background()

		formData := map[string]string{
			"username": "testuser",
			"password": "testpass",
		}

		resp, err := mock.PostForm(ctx, "/login", formData, nil)
		if err != nil {
			t.Fatalf("PostForm() failed: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}

		if mock.PostFormCalls != 1 {
			t.Errorf("Expected 1 PostForm call, got %d", mock.PostFormCalls)
		}
	})

	t.Run("Success_CustomFunction", func(t *testing.T) {
		var capturedFormData map[string]string
		mock := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				capturedFormData = formData
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       []byte(`{"authenticated": true}`),
				}, nil
			},
		}
		ctx := context.Background()

		formData := map[string]string{
			"field1": "value1",
			"field2": "value2",
		}

		resp, err := mock.PostForm(ctx, "/form", formData, nil)
		if err != nil {
			t.Fatalf("PostForm() failed: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}

		if capturedFormData["field1"] != "value1" {
			t.Error("Form data not captured correctly")
		}

		if mock.PostFormCalls != 1 {
			t.Errorf("Expected 1 PostForm call, got %d", mock.PostFormCalls)
		}
	})

	t.Run("Error_CustomError", func(t *testing.T) {
		expectedErr := errors.New("form error")
		mock := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, expectedErr
			},
		}
		ctx := context.Background()

		_, err := mock.PostForm(ctx, "/test", nil, nil)
		if err != expectedErr {
			t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
		}
	})

	t.Run("Success_EmptyFormData", func(t *testing.T) {
		mock := &mocks.MockHTTPClient{}
		ctx := context.Background()

		resp, err := mock.PostForm(ctx, "/test", map[string]string{}, nil)
		if err != nil {
			t.Fatalf("PostForm() with empty form data failed: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}
	})
}

func TestMockHTTPClient_Put(t *testing.T) {
	t.Run("Success_DefaultBehavior", func(t *testing.T) {
		mock := &mocks.MockHTTPClient{}
		ctx := context.Background()

		body := map[string]interface{}{
			"id":   123,
			"name": "updated",
		}

		resp, err := mock.Put(ctx, "/resource/123", body, nil)
		if err != nil {
			t.Fatalf("Put() failed: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}

		if mock.PutCalls != 1 {
			t.Errorf("Expected 1 Put call, got %d", mock.PutCalls)
		}
	})

	t.Run("Success_CustomFunction", func(t *testing.T) {
		var capturedPath string
		var capturedBody interface{}
		mock := &mocks.MockHTTPClient{
			PutFunc: func(ctx context.Context, path string, body interface{}, headers map[string]string) (*utils.HTTPResponse, error) {
				capturedPath = path
				capturedBody = body
				return &utils.HTTPResponse{
					StatusCode: 204,
					Body:       []byte(""),
				}, nil
			},
		}
		ctx := context.Background()

		body := map[string]interface{}{
			"status": "active",
		}

		resp, err := mock.Put(ctx, "/update/456", body, nil)
		if err != nil {
			t.Fatalf("Put() failed: %v", err)
		}

		if resp.StatusCode != 204 {
			t.Errorf("Expected status code 204, got %d", resp.StatusCode)
		}

		if capturedPath != "/update/456" {
			t.Errorf("Expected path '/update/456', got '%s'", capturedPath)
		}

		if capturedBody == nil {
			t.Error("Body not captured")
		}

		if mock.PutCalls != 1 {
			t.Errorf("Expected 1 Put call, got %d", mock.PutCalls)
		}
	})

	t.Run("Error_CustomError", func(t *testing.T) {
		expectedErr := errors.New("put error")
		mock := &mocks.MockHTTPClient{
			PutFunc: func(ctx context.Context, path string, body interface{}, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, expectedErr
			},
		}
		ctx := context.Background()

		_, err := mock.Put(ctx, "/test", nil, nil)
		if err != expectedErr {
			t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
		}
	})
}

func TestMockHTTPClient_Delete(t *testing.T) {
	t.Run("Success_DefaultBehavior", func(t *testing.T) {
		mock := &mocks.MockHTTPClient{}
		ctx := context.Background()

		resp, err := mock.Delete(ctx, "/resource/123", nil)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}

		if mock.DeleteCalls != 1 {
			t.Errorf("Expected 1 Delete call, got %d", mock.DeleteCalls)
		}
	})

	t.Run("Success_CustomFunction", func(t *testing.T) {
		var capturedPath string
		var capturedHeaders map[string]string
		mock := &mocks.MockHTTPClient{
			DeleteFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				capturedPath = path
				capturedHeaders = headers
				return &utils.HTTPResponse{
					StatusCode: 204,
					Body:       []byte(""),
				}, nil
			},
		}
		ctx := context.Background()

		headers := map[string]string{
			"Authorization": "Bearer token",
		}

		resp, err := mock.Delete(ctx, "/delete/789", headers)
		if err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}

		if resp.StatusCode != 204 {
			t.Errorf("Expected status code 204, got %d", resp.StatusCode)
		}

		if capturedPath != "/delete/789" {
			t.Errorf("Expected path '/delete/789', got '%s'", capturedPath)
		}

		if capturedHeaders["Authorization"] != "Bearer token" {
			t.Error("Headers not captured correctly")
		}

		if mock.DeleteCalls != 1 {
			t.Errorf("Expected 1 Delete call, got %d", mock.DeleteCalls)
		}
	})

	t.Run("Error_CustomError", func(t *testing.T) {
		expectedErr := errors.New("delete error")
		mock := &mocks.MockHTTPClient{
			DeleteFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, expectedErr
			},
		}
		ctx := context.Background()

		_, err := mock.Delete(ctx, "/test", nil)
		if err != expectedErr {
			t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
		}
	})

	t.Run("Success_NilHeaders", func(t *testing.T) {
		mock := &mocks.MockHTTPClient{}
		ctx := context.Background()

		resp, err := mock.Delete(ctx, "/test", nil)
		if err != nil {
			t.Fatalf("Delete() with nil headers failed: %v", err)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}
	})
}

func TestMockHTTPClient_Reset(t *testing.T) {
	t.Run("Success_ResetsAllCounters", func(t *testing.T) {
		mock := &mocks.MockHTTPClient{}
		ctx := context.Background()

		// Make various calls
		mock.Get(ctx, "/test", nil)
		mock.Get(ctx, "/test", nil)
		mock.Post(ctx, "/test", nil, nil)
		mock.PostForm(ctx, "/test", nil, nil)
		mock.PostForm(ctx, "/test", nil, nil)
		mock.PostForm(ctx, "/test", nil, nil)
		mock.Put(ctx, "/test", nil, nil)
		mock.Delete(ctx, "/test", nil)
		mock.Delete(ctx, "/test", nil)

		// Verify counters are incremented
		if mock.GetCalls != 2 {
			t.Errorf("Expected 2 Get calls, got %d", mock.GetCalls)
		}
		if mock.PostCalls != 1 {
			t.Errorf("Expected 1 Post call, got %d", mock.PostCalls)
		}
		if mock.PostFormCalls != 3 {
			t.Errorf("Expected 3 PostForm calls, got %d", mock.PostFormCalls)
		}
		if mock.PutCalls != 1 {
			t.Errorf("Expected 1 Put call, got %d", mock.PutCalls)
		}
		if mock.DeleteCalls != 2 {
			t.Errorf("Expected 2 Delete calls, got %d", mock.DeleteCalls)
		}

		// Reset
		mock.Reset()

		// Verify all counters are zero
		if mock.GetCalls != 0 {
			t.Errorf("Expected 0 Get calls after reset, got %d", mock.GetCalls)
		}
		if mock.PostCalls != 0 {
			t.Errorf("Expected 0 Post calls after reset, got %d", mock.PostCalls)
		}
		if mock.PostFormCalls != 0 {
			t.Errorf("Expected 0 PostForm calls after reset, got %d", mock.PostFormCalls)
		}
		if mock.PutCalls != 0 {
			t.Errorf("Expected 0 Put calls after reset, got %d", mock.PutCalls)
		}
		if mock.DeleteCalls != 0 {
			t.Errorf("Expected 0 Delete calls after reset, got %d", mock.DeleteCalls)
		}
	})

	t.Run("Success_ResetDoesNotAffectFunctions", func(t *testing.T) {
		customGetFunc := func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
			return &utils.HTTPResponse{StatusCode: 201, Body: []byte("custom")}, nil
		}

		mock := &mocks.MockHTTPClient{
			GetFunc: customGetFunc,
		}
		ctx := context.Background()

		// Make a call
		resp1, _ := mock.Get(ctx, "/test", nil)
		if resp1.StatusCode != 201 {
			t.Error("Custom function not working before reset")
		}

		// Reset
		mock.Reset()

		// Verify custom function still works
		resp2, _ := mock.Get(ctx, "/test", nil)
		if resp2.StatusCode != 201 {
			t.Error("Custom function not working after reset")
		}
		if mock.GetCalls != 1 {
			t.Errorf("Expected 1 Get call after reset, got %d", mock.GetCalls)
		}
	})
}

func TestMockHTTPClient_CallTracking(t *testing.T) {
	t.Run("Success_IndependentCounters", func(t *testing.T) {
		mock := &mocks.MockHTTPClient{}
		ctx := context.Background()

		// Test that each method has independent counter
		mock.Get(ctx, "/test", nil)
		if mock.GetCalls != 1 || mock.PostCalls != 0 {
			t.Error("Counters not independent after Get")
		}

		mock.Post(ctx, "/test", nil, nil)
		if mock.GetCalls != 1 || mock.PostCalls != 1 {
			t.Error("Counters not independent after Post")
		}

		mock.PostForm(ctx, "/test", nil, nil)
		if mock.PostFormCalls != 1 {
			t.Error("PostForm counter not working")
		}

		mock.Put(ctx, "/test", nil, nil)
		if mock.PutCalls != 1 {
			t.Error("Put counter not working")
		}

		mock.Delete(ctx, "/test", nil)
		if mock.DeleteCalls != 1 {
			t.Error("Delete counter not working")
		}
	})

	t.Run("Success_CountersIncrementCorrectly", func(t *testing.T) {
		mock := &mocks.MockHTTPClient{}
		ctx := context.Background()

		// Test multiple calls increment correctly
		for i := 1; i <= 10; i++ {
			mock.Get(ctx, "/test", nil)
			if mock.GetCalls != i {
				t.Errorf("Expected %d Get calls, got %d", i, mock.GetCalls)
			}
		}
	})
}

func TestMockHTTPClient_ContextPropagation(t *testing.T) {
	t.Run("Success_ContextPassedToCustomFunc", func(t *testing.T) {
		type contextKey string
		key := contextKey("test-key")
		expectedValue := "test-value"

		var capturedCtx context.Context
		mock := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				capturedCtx = ctx
				return &utils.HTTPResponse{StatusCode: 200, Body: []byte("{}")}, nil
			},
		}

		ctx := context.WithValue(context.Background(), key, expectedValue)
		mock.Get(ctx, "/test", nil)

		if capturedCtx == nil {
			t.Fatal("Context not passed to custom function")
		}

		value := capturedCtx.Value(key)
		if value != expectedValue {
			t.Errorf("Expected context value '%s', got '%v'", expectedValue, value)
		}
	})

	t.Run("Success_CancelledContext", func(t *testing.T) {
		mock := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					return &utils.HTTPResponse{StatusCode: 200, Body: []byte("{}")}, nil
				}
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := mock.Get(ctx, "/test", nil)
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
	})
}

func TestMockHTTPClient_ConcurrentCalls(t *testing.T) {
	t.Run("Success_ConcurrentAccess", func(t *testing.T) {
		t.Skip("Skipping concurrent test - MockHTTPClient counters are not thread-safe by design")

		mock := &mocks.MockHTTPClient{}
		ctx := context.Background()

		done := make(chan bool)
		numGoroutines := 10
		callsPerGoroutine := 10

		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < callsPerGoroutine; j++ {
					mock.Get(ctx, "/test", nil)
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		expectedCalls := numGoroutines * callsPerGoroutine
		if mock.GetCalls != expectedCalls {
			t.Errorf("Expected %d Get calls, got %d", expectedCalls, mock.GetCalls)
		}
	})
}

func TestMockHTTPClient_ResponseVariations(t *testing.T) {
	t.Run("Success_DifferentStatusCodes", func(t *testing.T) {
		statusCodes := []int{200, 201, 204, 400, 404, 500, 503}

		for _, code := range statusCodes {
			mock := &mocks.MockHTTPClient{
				GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
					return &utils.HTTPResponse{StatusCode: code, Body: []byte("{}")}, nil
				},
			}
			ctx := context.Background()

			resp, err := mock.Get(ctx, "/test", nil)
			if err != nil {
				t.Fatalf("Get() failed for status code %d: %v", code, err)
			}

			if resp.StatusCode != code {
				t.Errorf("Expected status code %d, got %d", code, resp.StatusCode)
			}
		}
	})

	t.Run("Success_DifferentBodyTypes", func(t *testing.T) {
		bodies := []string{
			"{}",
			`{"key": "value"}`,
			`[1, 2, 3]`,
			"plain text",
			"",
		}

		for _, body := range bodies {
			mock := &mocks.MockHTTPClient{
				GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
					return &utils.HTTPResponse{StatusCode: 200, Body: []byte(body)}, nil
				},
			}
			ctx := context.Background()

			resp, err := mock.Get(ctx, "/test", nil)
			if err != nil {
				t.Fatalf("Get() failed: %v", err)
			}

			if string(resp.Body) != body {
				t.Errorf("Expected body '%s', got '%s'", body, string(resp.Body))
			}
		}
	})

	t.Run("Success_WithHeaders", func(t *testing.T) {
		expectedHeaders := map[string][]string{
			"Content-Type": {"application/json"},
			"X-Custom":     {"value"},
		}

		mock := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       []byte("{}"),
					Headers:    expectedHeaders,
				}, nil
			},
		}
		ctx := context.Background()

		resp, err := mock.Get(ctx, "/test", nil)
		if err != nil {
			t.Fatalf("Get() failed: %v", err)
		}

		if resp.Headers["Content-Type"][0] != "application/json" {
			t.Error("Headers not set correctly")
		}
	})
}
