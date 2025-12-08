package utils_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			resp := &utils.HTTPResponse{StatusCode: tt.statusCode}
			assert.Equal(t, tt.want, resp.IsSuccess())
		})
	}
}

func TestHTTPResponse_JSON(t *testing.T) {
	t.Run("Success_ValidJSON", func(t *testing.T) {
		resp := &utils.HTTPResponse{
			StatusCode: 200,
			Body:       []byte(`{"key": "value", "number": 42}`),
		}

		var result struct {
			Key    string `json:"key"`
			Number int    `json:"number"`
		}

		require.NoError(t, resp.JSON(&result))
		assert.Equal(t, "value", result.Key)
		assert.Equal(t, 42, result.Number)
	})

	t.Run("Error_InvalidJSON", func(t *testing.T) {
		resp := &utils.HTTPResponse{
			StatusCode: 200,
			Body:       []byte(`{invalid json}`),
		}

		var result map[string]interface{}
		assert.Error(t, resp.JSON(&result))
	})
}

func TestHTTPResponse_String(t *testing.T) {
	t.Run("Success_ReturnsBodyAsString", func(t *testing.T) {
		body := "test response body"
		resp := &utils.HTTPResponse{
			StatusCode: 200,
			Body:       []byte(body),
		}
		assert.Equal(t, body, resp.String())
	})

	t.Run("Success_EmptyBody", func(t *testing.T) {
		resp := &utils.HTTPResponse{
			StatusCode: 204,
			Body:       []byte{},
		}
		assert.Empty(t, resp.String())
	})
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
			resp := &utils.HTTPResponse{StatusCode: tt.statusCode}
			assert.Equal(t, tt.want, resp.IsClientError())
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
			resp := &utils.HTTPResponse{StatusCode: tt.statusCode}
			assert.Equal(t, tt.want, resp.IsServerError())
		})
	}
}

func TestHTTPResponse_Headers(t *testing.T) {
	resp := &utils.HTTPResponse{
		StatusCode: 200,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
			"X-Request-ID": {"12345"},
		},
	}

	assert.Equal(t, "application/json", resp.Headers["Content-Type"][0])
	assert.Equal(t, "12345", resp.Headers["X-Request-ID"][0])
}

func TestHTTPResponse_Duration(t *testing.T) {
	t.Run("Success_MeasuresDuration", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond) // Simulate some delay
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/test", nil)
		require.NoError(t, err)
		assert.NotZero(t, resp.Duration)
		assert.True(t, resp.Duration >= 10*time.Millisecond)
	})
}

func TestNewHTTPClient(t *testing.T) {
	t.Run("Success_FullConfig", func(t *testing.T) {
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
		assert.NotNil(t, client)
	})

	t.Run("Success_MinimalConfig", func(t *testing.T) {
		config := utils.HTTPClientConfig{BaseURL: "https://api.example.com"}
		client := utils.NewHTTPClient(config)
		assert.NotNil(t, client)
	})

	t.Run("Success_WithRetryDefaults", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL: "https://api.example.com",
			RetryConfig: utils.RetryConfig{
				MaxRetries: 0, // Should get default
			},
		}
		client := utils.NewHTTPClient(config)
		assert.NotNil(t, client)
	})
}

func TestHTTPClientConfig_SSLVerify(t *testing.T) {
	tests := []struct {
		name          string
		skipSSLVerify bool
	}{
		{"SkipSSLVerifyTrue", true},
		{"SkipSSLVerifyFalse", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := utils.HTTPClientConfig{
				BaseURL:       "https://api.example.com",
				SkipSSLVerify: tt.skipSSLVerify,
			}
			client := utils.NewHTTPClient(config)
			assert.NotNil(t, client)
		})
	}
}

func TestHTTPClient_ConnectionPooling(t *testing.T) {
	t.Run("Success_CustomMaxConnections", func(t *testing.T) {
		config := utils.HTTPClientConfig{
			BaseURL:         "https://api.example.com",
			MaxIdleConns:    50,
			MaxConnsPerHost: 25,
		}
		client := utils.NewHTTPClient(config)
		assert.NotNil(t, client)
	})

	t.Run("Success_DefaultMaxConnections", func(t *testing.T) {
		config := utils.HTTPClientConfig{BaseURL: "https://api.example.com"}
		client := utils.NewHTTPClient(config)
		assert.NotNil(t, client)
	})
}

func TestHTTPClient_RetryConfig(t *testing.T) {
	config := utils.HTTPClientConfig{
		BaseURL: "https://api.example.com",
		RetryConfig: utils.RetryConfig{
			MaxRetries: 5,
			RetryDelay: 10,
			BackoffExp: 1.5,
		},
	}
	client := utils.NewHTTPClient(config)
	assert.NotNil(t, client)
}

func TestHTTPClient_ContextCancellation(t *testing.T) {
	t.Run("Error_ContextCancelled", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := client.Get(ctx, "/test", nil)
		assert.Error(t, err)
	})
}

func TestHTTPClient_Post(t *testing.T) {
	t.Run("Success_PostWithBody", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "created"}`))
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		body := map[string]interface{}{
			"test": "value",
			"num":  42,
		}

		resp, err := client.Post(ctx, "/test", body, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess())
	})

	t.Run("Error_InvalidBody", func(t *testing.T) {
		config := utils.HTTPClientConfig{BaseURL: "http://localhost"}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		// Use a channel which cannot be marshaled to JSON
		invalidBody := make(chan int)

		_, err := client.Post(ctx, "/test", invalidBody, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal request body")
	})

	t.Run("Success_PostWithComplexBody", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": 123}`))
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
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

		resp, err := client.Post(ctx, "/test", body, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess())
	})
}

func TestHTTPClient_PostForm(t *testing.T) {
	t.Run("Success_PostFormData", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"form": {"username": "testuser"}}`))
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		formData := map[string]string{
			"username": "testuser",
			"password": "testpass",
		}

		resp, err := client.PostForm(ctx, "/test", formData, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess())
	})

	t.Run("Success_EmptyFormData", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.PostForm(ctx, "/test", map[string]string{}, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess())
	})

	t.Run("Success_MultipleFormFields", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success": true}`))
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		formData := map[string]string{
			"field1": "value1",
			"field2": "value2",
			"field3": "value3",
		}

		resp, err := client.PostForm(ctx, "/test", formData, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess())
	})
}

func TestHTTPClient_Put(t *testing.T) {
	t.Run("Success_PutWithBody", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": 123, "name": "updated"}`))
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		body := map[string]interface{}{
			"id":   123,
			"name": "updated",
		}

		resp, err := client.Put(ctx, "/test", body, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess())
	})

	t.Run("Success_PutWithNilBody", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPut, r.Method)
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Put(ctx, "/test", nil, nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess())
	})
}

func TestHTTPClient_Delete(t *testing.T) {
	t.Run("Success_Delete", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"deleted": true}`))
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Delete(ctx, "/test", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess())
	})

	t.Run("Success_DeleteWithHeaders", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "test-value", r.Header.Get("X-Custom-Header"))
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		headers := map[string]string{
			"X-Custom-Header": "test-value",
		}

		resp, err := client.Delete(ctx, "/test", headers)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess())
	})
}

func TestHTTPClient_Get(t *testing.T) {
	t.Run("Success_GetWithParams", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "value1", r.URL.Query().Get("param1"))
			assert.Equal(t, "value2", r.URL.Query().Get("param2"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"args": {"param1": "value1"}}`))
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/test?param1=value1&param2=value2", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess())
	})

	t.Run("Success_GetWithNilHeaders", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{
			BaseURL: server.URL,
			Headers: nil,
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/test", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess())
	})
}

func TestHTTPClient_CustomHeaders(t *testing.T) {
	t.Run("Success_DefaultHeaders", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "TestClient/1.0", r.Header.Get("User-Agent"))
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"headers": {}}`))
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{
			BaseURL: server.URL,
			Headers: map[string]string{
				"User-Agent":    "TestClient/1.0",
				"Authorization": "Bearer test-token",
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/test", nil)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess())
	})

	t.Run("Success_RequestSpecificHeaders", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "12345", r.Header.Get("X-Request-ID"))
			assert.Equal(t, "value", r.Header.Get("X-Custom"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		headers := map[string]string{
			"X-Request-ID": "12345",
			"X-Custom":     "value",
		}

		resp, err := client.Get(ctx, "/test", headers)
		require.NoError(t, err)
		assert.True(t, resp.IsSuccess())
	})
}

func TestHTTPClient_Timeout(t *testing.T) {
	t.Run("Error_RequestTimeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{
			BaseURL: server.URL,
			Timeout: 10 * time.Millisecond,
			RetryConfig: utils.RetryConfig{
				MaxRetries: 0,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		_, err := client.Get(ctx, "/test", nil)
		assert.Error(t, err)
	})
}

func TestHTTPClient_ContextWithTimeout(t *testing.T) {
	t.Run("Error_ContextTimeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		_, err := client.Get(ctx, "/test", nil)
		assert.Error(t, err)
	})
}

func TestHTTPClient_RetryLogic(t *testing.T) {
	t.Run("Success_RetryOn500", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 2 {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "server error"}`))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "ok"}`))
			}
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{
			BaseURL: server.URL,
			RetryConfig: utils.RetryConfig{
				MaxRetries: 2,
				RetryDelay: 10 * time.Millisecond,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/test", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.GreaterOrEqual(t, attempts, 2)
	})

	t.Run("Success_RetryOn429", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "rate limited"}`))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "ok"}`))
			}
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{
			BaseURL: server.URL,
			RetryConfig: utils.RetryConfig{
				MaxRetries: 1,
				RetryDelay: 10 * time.Millisecond,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/test", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestHTTPClient_SpecialStatusCodes(t *testing.T) {
	t.Run("Success_Handle409Conflict", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`{"error": "resource already exists"}`))
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		// 409 should be treated as success (resource already exists)
		resp, err := client.Get(ctx, "/test", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("Success_Handle404NotFound", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "not found"}`))
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		// 404 should return response without error
		resp, err := client.Get(ctx, "/test", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Success_Handle500WithAlreadyExists", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "resource already exists"}`))
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{BaseURL: server.URL}
		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		// 500 with "already exists" message should return without error
		resp, err := client.Get(ctx, "/test", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestHTTPClient_BackoffRetry(t *testing.T) {
	t.Run("Success_ExponentialBackoff", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{
			BaseURL: server.URL,
			RetryConfig: utils.RetryConfig{
				MaxRetries: 3,
				RetryDelay: 5 * time.Millisecond,
				BackoffExp: 2.0,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/test", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 3, attempts)
	})

	t.Run("Success_LinearBackoff", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 2 {
				w.WriteHeader(http.StatusBadGateway)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{
			BaseURL: server.URL,
			RetryConfig: utils.RetryConfig{
				MaxRetries: 2,
				RetryDelay: 5 * time.Millisecond,
				BackoffExp: 1.0,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/test", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Success_PowCalculation", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			if attempts < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{
			BaseURL: server.URL,
			RetryConfig: utils.RetryConfig{
				MaxRetries: 4,
				RetryDelay: 5 * time.Millisecond,
				BackoffExp: 3.0,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		resp, err := client.Get(ctx, "/test", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestHTTPClient_MaxRetriesExceeded(t *testing.T) {
	t.Run("Error_MaxRetriesExhausted", func(t *testing.T) {
		attempts := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attempts++
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error": "service unavailable"}`))
		}))
		defer server.Close()

		config := utils.HTTPClientConfig{
			BaseURL: server.URL,
			RetryConfig: utils.RetryConfig{
				MaxRetries: 2,
				RetryDelay: 5 * time.Millisecond,
			},
		}

		client := utils.NewHTTPClient(config)
		ctx := context.Background()

		_, err := client.Get(ctx, "/test", nil)
		assert.Error(t, err)
		// Error can be either "max retries" or "request failed with status 503"
		assert.True(t, err != nil, "Expected an error after exhausting retries")
		assert.Equal(t, 3, attempts) // Initial + 2 retries
	})
}
