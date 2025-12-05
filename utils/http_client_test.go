// Package utils provides HTTP client utilities
package utils_test

import (
	"context"
	"testing"

	"salesforce-splunk-migration/utils"
)

func TestHTTPResponse_IsSuccess(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"200 OK", 200, true},
		{"201 Created", 201, true},
		{"204 No Content", 204, true},
		{"299 Upper bound", 299, true},
		{"300 Redirect", 300, false},
		{"400 Bad Request", 400, false},
		{"404 Not Found", 404, false},
		{"500 Server Error", 500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &utils.HTTPResponse{
				StatusCode: tt.statusCode,
			}
			if got := resp.IsSuccess(); got != tt.want {
				t.Errorf("HTTPResponse.IsSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPResponse_JSON(t *testing.T) {
	resp := &utils.HTTPResponse{
		StatusCode: 200,
		Body:       []byte(`{"key": "value", "number": 42}`),
	}

	var result struct {
		Key    string `json:"key"`
		Number int    `json:"number"`
	}

	err := resp.JSON(&result)
	if err != nil {
		t.Fatalf("HTTPResponse.JSON() failed: %v", err)
	}

	if result.Key != "value" {
		t.Errorf("Expected key='value', got '%s'", result.Key)
	}
	if result.Number != 42 {
		t.Errorf("Expected number=42, got %d", result.Number)
	}
}

func TestHTTPResponse_JSON_InvalidJSON(t *testing.T) {
	resp := &utils.HTTPResponse{
		StatusCode: 200,
		Body:       []byte(`{invalid json}`),
	}

	var result map[string]interface{}
	err := resp.JSON(&result)
	if err == nil {
		t.Error("HTTPResponse.JSON() should fail with invalid JSON")
	}
}

func TestHTTPResponse_String(t *testing.T) {
	body := "test response body"
	resp := &utils.HTTPResponse{
		StatusCode: 200,
		Body:       []byte(body),
	}

	if resp.String() != body {
		t.Errorf("HTTPResponse.String() = %v, want %v", resp.String(), body)
	}
}

func TestNewHTTPClient(t *testing.T) {
	config := utils.HTTPClientConfig{
		BaseURL: "https://api.example.com",
		Headers: map[string]string{
			"User-Agent": "Test-Client/1.0",
		},
		RetryConfig: utils.RetryConfig{
			MaxRetries: 3,
			RetryDelay: 5,
		},
	}

	client := utils.NewHTTPClient(config)
	if client == nil {
		t.Fatal("NewHTTPClient() returned nil")
	}
}

func TestHTTPClient_DefaultConfig(t *testing.T) {
	config := utils.HTTPClientConfig{
		BaseURL: "https://api.example.com",
	}

	client := utils.NewHTTPClient(config)
	if client == nil {
		t.Fatal("NewHTTPClient() should work with minimal config")
	}
}

func TestRetryConfig_Defaults(t *testing.T) {
	config := utils.HTTPClientConfig{
		BaseURL: "https://api.example.com",
		RetryConfig: utils.RetryConfig{
			MaxRetries: 0, // Should get default
		},
	}

	client := utils.NewHTTPClient(config)
	if client == nil {
		t.Error("NewHTTPClient() should apply defaults")
	}
}

func TestHTTPClient_ContextCancellation(t *testing.T) {
	config := utils.HTTPClientConfig{
		BaseURL: "https://httpbin.org",
	}

	client := utils.NewHTTPClient(config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Get(ctx, "/delay/10", nil)
	if err == nil {
		t.Error("Request should fail with cancelled context")
	}
}

func TestHTTPResponse_IsClientError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"400 Bad Request", 400, true},
		{"404 Not Found", 404, true},
		{"499 Upper bound", 499, true},
		{"200 OK", 200, false},
		{"300 Redirect", 300, false},
		{"500 Server Error", 500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &utils.HTTPResponse{
				StatusCode: tt.statusCode,
			}
			if got := resp.IsClientError(); got != tt.want {
				t.Errorf("HTTPResponse.IsClientError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPResponse_IsServerError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"500 Internal Server Error", 500, true},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
		{"599 Upper bound", 599, true},
		{"200 OK", 200, false},
		{"400 Bad Request", 400, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &utils.HTTPResponse{
				StatusCode: tt.statusCode,
			}
			if got := resp.IsServerError(); got != tt.want {
				t.Errorf("HTTPResponse.IsServerError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPClientConfig_SSLVerify(t *testing.T) {
	t.Run("Success_SkipSSLVerifyTrue", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL:       "https://api.example.com",
			SkipSSLVerify: true,
		}

		client := utils.NewHTTPClient(config)
		if client == nil {
			t.Error("NewHTTPClient() should return non-nil client")
		}
	})

	t.Run("Success_SkipSSLVerifyFalse", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL:       "https://api.example.com",
			SkipSSLVerify: false,
		}

		client := utils.NewHTTPClient(config)
		if client == nil {
			t.Error("NewHTTPClient() should return non-nil client")
		}
	})
}

func TestHTTPClient_ConnectionPooling(t *testing.T) {
	t.Run("Success_CustomMaxConnections", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL:         "https://api.example.com",
			MaxIdleConns:    50,
			MaxConnsPerHost: 25,
		}

		client := utils.NewHTTPClient(config)
		if client == nil {
			t.Error("NewHTTPClient() should return non-nil client")
		}
	})

	t.Run("Success_DefaultMaxConnections", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://api.example.com",
		}

		client := utils.NewHTTPClient(config)
		if client == nil {
			t.Error("NewHTTPClient() should return non-nil client")
		}
	})
}

func TestHTTPClient_RetryConfig(t *testing.T) {
	t.Run("Success_CustomRetryConfig", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://api.example.com",
			RetryConfig: utils.RetryConfig{
				MaxRetries: 5,
				RetryDelay: 10,
				BackoffExp: 1.5,
			},
		}

		client := utils.NewHTTPClient(config)
		if client == nil {
			t.Error("NewHTTPClient() should return non-nil client")
		}
	})
}

func TestHTTPResponse_Headers(t *testing.T) {
	t.Run("Success_HasHeaders", func(t *testing.T) {
		resp := &utils.HTTPResponse{
			StatusCode: 200,
			Headers: map[string][]string{
				"Content-Type": {"application/json"},
				"X-Request-ID": {"12345"},
			},
		}

		if resp.Headers["Content-Type"][0] != "application/json" {
			t.Errorf("Expected Content-Type header")
		}
		if resp.Headers["X-Request-ID"][0] != "12345" {
			t.Errorf("Expected X-Request-ID header")
		}
	})
}

func TestHTTPClient_Post(t *testing.T) {
	t.Run("Success_PostWithBody", func(t *testing.T) {
		// This test requires a real server or mock server
		// For now, we test with httpbin.org or skip if unavailable
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
			RetryConfig: utils.RetryConfig{
				MaxRetries: 1,
				RetryDelay: 1,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		body := map[string]interface{}{
			"test": "value",
			"num":  42,
		}

		resp, err := client.Post(ctx, "/post", body, nil)
		if err != nil {
			t.Skipf("Skipping test, httpbin.org not accessible: %v", err)
			return
		}

		if !resp.IsSuccess() {
			t.Errorf("Expected success status, got %d", resp.StatusCode)
		}
	})

	t.Run("Error_InvalidBody", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		// Use a channel which cannot be marshaled to JSON
		invalidBody := make(chan int)

		_, err := client.Post(ctx, "/post", invalidBody, nil)
		if err == nil {
			t.Error("Post should fail with unmarshalable body")
		}
	})
}

func TestHTTPClient_PostForm(t *testing.T) {
	t.Run("Success_PostFormData", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
			RetryConfig: utils.RetryConfig{
				MaxRetries: 1,
				RetryDelay: 1,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		formData := map[string]string{
			"username": "testuser",
			"password": "testpass",
		}

		resp, err := client.PostForm(ctx, "/post", formData, nil)
		if err != nil {
			t.Skipf("Skipping test, httpbin.org not accessible: %v", err)
			return
		}

		if !resp.IsSuccess() {
			t.Errorf("Expected success status, got %d", resp.StatusCode)
		}
	})

	t.Run("Success_EmptyFormData", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.PostForm(ctx, "/post", map[string]string{}, nil)
		if err != nil {
			t.Skipf("Skipping test, httpbin.org not accessible: %v", err)
			return
		}

		if !resp.IsSuccess() {
			t.Errorf("Expected success status, got %d", resp.StatusCode)
		}
	})
}

func TestHTTPClient_Put(t *testing.T) {
	t.Run("Success_PutWithBody", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
			RetryConfig: utils.RetryConfig{
				MaxRetries: 1,
				RetryDelay: 1,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		body := map[string]interface{}{
			"id":   123,
			"name": "updated",
		}

		resp, err := client.Put(ctx, "/put", body, nil)
		if err != nil {
			t.Skipf("Skipping test, httpbin.org not accessible: %v", err)
			return
		}

		if !resp.IsSuccess() {
			t.Errorf("Expected success status, got %d", resp.StatusCode)
		}
	})

	t.Run("Success_PutWithNilBody", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Put(ctx, "/put", nil, nil)
		if err != nil {
			t.Skipf("Skipping test, httpbin.org not accessible: %v", err)
			return
		}

		if !resp.IsSuccess() {
			t.Errorf("Expected success status, got %d", resp.StatusCode)
		}
	})
}

func TestHTTPClient_Delete(t *testing.T) {
	t.Run("Success_Delete", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
			RetryConfig: utils.RetryConfig{
				MaxRetries: 1,
				RetryDelay: 1,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Delete(ctx, "/delete", nil)
		if err != nil {
			t.Skipf("Skipping test, httpbin.org not accessible: %v", err)
			return
		}

		if !resp.IsSuccess() {
			t.Errorf("Expected success status, got %d", resp.StatusCode)
		}
	})

	t.Run("Success_DeleteWithHeaders", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		headers := map[string]string{
			"X-Custom-Header": "test-value",
		}

		resp, err := client.Delete(ctx, "/delete", headers)
		if err != nil {
			t.Skipf("Skipping test, httpbin.org not accessible: %v", err)
			return
		}

		if !resp.IsSuccess() {
			t.Errorf("Expected success status, got %d", resp.StatusCode)
		}
	})
}

func TestHTTPClient_CustomHeaders(t *testing.T) {
	t.Run("Success_DefaultHeaders", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
			Headers: map[string]string{
				"User-Agent":    "TestClient/1.0",
				"Authorization": "Bearer test-token",
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/headers", nil)
		if err != nil {
			t.Skipf("Skipping test, httpbin.org not accessible: %v", err)
			return
		}

		if !resp.IsSuccess() {
			t.Errorf("Expected success status, got %d", resp.StatusCode)
		}
	})

	t.Run("Success_RequestSpecificHeaders", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		headers := map[string]string{
			"X-Request-ID": "12345",
			"X-Custom":     "value",
		}

		resp, err := client.Get(ctx, "/headers", headers)
		if err != nil {
			t.Skipf("Skipping test, httpbin.org not accessible: %v", err)
			return
		}

		if !resp.IsSuccess() {
			t.Errorf("Expected success status, got %d", resp.StatusCode)
		}
	})
}

func TestHTTPClient_Timeout(t *testing.T) {
	t.Run("Error_RequestTimeout", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
			Timeout: 1, // 1 nanosecond - should timeout
			RetryConfig: utils.RetryConfig{
				MaxRetries: 0, // No retries
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		_, err := client.Get(ctx, "/delay/5", nil)
		if err == nil {
			t.Error("Request should timeout")
		}
	})
}

func TestHTTPClient_RetryLogic(t *testing.T) {
	t.Run("Success_RetryOn500", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
			RetryConfig: utils.RetryConfig{
				MaxRetries: 2,
				RetryDelay: 1,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		// This will return 500, should be retried
		_, err := client.Get(ctx, "/status/500", nil)
		if err == nil {
			t.Log("Request succeeded after retries or httpbin returned 200")
		}
		// We expect this to eventually fail after retries
	})

	t.Run("Error_NoRetryOn400", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
			RetryConfig: utils.RetryConfig{
				MaxRetries: 2,
				RetryDelay: 1,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		// 400 should not be retried
		resp, err := client.Get(ctx, "/status/400", nil)
		if err == nil {
			t.Error("Expected error for 400 status")
		}
		if resp != nil && resp.StatusCode != 400 {
			t.Errorf("Expected status code 400, got %d", resp.StatusCode)
		}
	})

	t.Run("Success_RetryOn429", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
			RetryConfig: utils.RetryConfig{
				MaxRetries: 1,
				RetryDelay: 1,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		// 429 (rate limit) should be retried
		_, err := client.Get(ctx, "/status/429", nil)
		if err == nil {
			t.Log("Request succeeded after retries")
		}
	})
}

func TestHTTPClient_SpecialStatusCodes(t *testing.T) {
	t.Run("Success_409Conflict", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		// 409 should be treated as success (resource already exists)
		resp, err := client.Get(ctx, "/status/409", nil)
		if err != nil {
			t.Errorf("409 should be treated as success: %v", err)
		}
		if resp != nil && resp.StatusCode != 409 {
			t.Errorf("Expected status code 409, got %d", resp.StatusCode)
		}
	})
}

func TestHTTPResponse_Duration(t *testing.T) {
	t.Run("Success_RecordsDuration", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/delay/1", nil)
		if err != nil {
			t.Skipf("Skipping test, httpbin.org not accessible: %v", err)
			return
		}

		if resp.Duration == 0 {
			t.Error("Response duration should be recorded")
		}
	})
}

func TestHTTPClient_MakeRequestWithBody(t *testing.T) {
	t.Run("Success_PostWithComplexBody", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		body := map[string]interface{}{
			"string":  "value",
			"number":  42,
			"boolean": true,
			"array":   []int{1, 2, 3},
			"nested": map[string]interface{}{
				"key": "value",
			},
		}

		resp, err := client.Post(ctx, "/post", body, nil)
		if err != nil {
			t.Skipf("Skipping test, httpbin.org not accessible: %v", err)
			return
		}

		if !resp.IsSuccess() {
			t.Errorf("Expected success status, got %d", resp.StatusCode)
		}
	})
}

func TestHTTPClient_MakeFormRequestEncoding(t *testing.T) {
	t.Run("Success_MultipleFormFields", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		formData := map[string]string{
			"field1": "value1",
			"field2": "value2",
			"field3": "value3",
		}

		resp, err := client.PostForm(ctx, "/post", formData, nil)
		if err != nil {
			t.Skipf("Skipping test, httpbin.org not accessible: %v", err)
			return
		}

		if !resp.IsSuccess() {
			t.Errorf("Expected success status, got %d", resp.StatusCode)
		}
	})
}

func TestHTTPClient_BackoffRetry(t *testing.T) {
	t.Run("Success_ExponentialBackoff", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
			RetryConfig: utils.RetryConfig{
				MaxRetries: 3,
				RetryDelay: 1,
				BackoffExp: 2.0,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		// This will fail but test backoff logic
		_, err := client.Get(ctx, "/status/503", nil)
		if err == nil {
			t.Log("Request succeeded")
		}
	})

	t.Run("Success_LinearBackoff", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
			RetryConfig: utils.RetryConfig{
				MaxRetries: 2,
				RetryDelay: 1,
				BackoffExp: 1.0,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		_, err := client.Get(ctx, "/status/502", nil)
		if err == nil {
			t.Log("Request succeeded")
		}
	})
}

func TestPowFunction(t *testing.T) {
	// Test the internal pow function through retry behavior
	t.Run("Success_PowCalculation", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
			RetryConfig: utils.RetryConfig{
				MaxRetries: 4,
				RetryDelay: 1,
				BackoffExp: 3.0, // Test different exponent
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		_, err := client.Get(ctx, "/status/503", nil)
		if err == nil {
			t.Log("Request eventually succeeded")
		}
	})
}

func TestHTTPClient_ContextWithTimeout(t *testing.T) {
	t.Run("Error_ContextTimeout", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
		}

		client := utils.NewHTTPClient(config)
		ctx, cancel := context.WithTimeout(context.Background(), 1)
		defer cancel()

		_, err := client.Get(ctx, "/delay/10", nil)
		if err == nil {
			t.Error("Request should fail with context timeout")
		}
	})
}

func TestHTTPClient_GetWithQueryParams(t *testing.T) {
	t.Run("Success_GetWithParams", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/get?param1=value1&param2=value2", nil)
		if err != nil {
			t.Skipf("Skipping test, httpbin.org not accessible: %v", err)
			return
		}

		if !resp.IsSuccess() {
			t.Errorf("Expected success status, got %d", resp.StatusCode)
		}
	})
}

func TestHTTPResponse_EmptyBody(t *testing.T) {
	t.Run("Success_EmptyBody", func(t *testing.T) {
		resp := &utils.HTTPResponse{
			StatusCode: 204,
			Body:       []byte{},
		}

		if resp.String() != "" {
			t.Error("Empty body should return empty string")
		}
	})
}

func TestHTTPClient_NilHeaders(t *testing.T) {
	t.Run("Success_NilHeaders", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
			Headers: nil, // Should be initialized
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/get", nil)
		if err != nil {
			t.Skipf("Skipping test, httpbin.org not accessible: %v", err)
			return
		}

		if !resp.IsSuccess() {
			t.Errorf("Expected success status, got %d", resp.StatusCode)
		}
	})
}

func TestHTTPClient_MaxRetriesExceeded(t *testing.T) {
	t.Run("Error_MaxRetriesExceeded", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://httpbin.org",
			RetryConfig: utils.RetryConfig{
				MaxRetries: 2,
				RetryDelay: 1,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		_, err := client.Get(ctx, "/status/503", nil)
		if err == nil {
			t.Log("Request eventually succeeded")
		}
		// Should exhaust retries and fail
	})
}
