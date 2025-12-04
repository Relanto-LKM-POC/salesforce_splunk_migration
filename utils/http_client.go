// Package utils provides HTTP client utilities and communication abstractions
package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// HTTPClient provides a wrapper around http.Client with additional utilities
type HTTPClient struct {
	client      *http.Client
	baseURL     string
	headers     map[string]string
	timeout     time.Duration
	retryConfig RetryConfig
}

// HTTPClientConfig holds configuration for HTTP client
type HTTPClientConfig struct {
	BaseURL         string
	Timeout         time.Duration
	Headers         map[string]string
	RetryConfig     RetryConfig
	SkipSSLVerify   bool
	MaxIdleConns    int
	MaxConnsPerHost int
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries int
	RetryDelay time.Duration
	BackoffExp float64
}

// HTTPResponse represents a standardized HTTP response
type HTTPResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
	Duration   time.Duration
}

// NewHTTPClient creates a new HTTP client with connection pooling configuration
func NewHTTPClient(config HTTPClientConfig) *HTTPClient {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.Headers == nil {
		config.Headers = make(map[string]string)
	}

	if config.RetryConfig.MaxRetries == 0 {
		config.RetryConfig.MaxRetries = 3
	}

	if config.RetryConfig.RetryDelay == 0 {
		config.RetryConfig.RetryDelay = 5 * time.Second
	}

	if config.RetryConfig.BackoffExp == 0 {
		config.RetryConfig.BackoffExp = 2.0
	}

	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 100
	}

	if config.MaxConnsPerHost == 0 {
		config.MaxConnsPerHost = 100
	}

	// Create transport with connection pooling
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.SkipSSLVerify,
		},
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout:   config.Timeout,
			Transport: transport,
		},
		baseURL:     config.BaseURL,
		headers:     config.Headers,
		timeout:     config.Timeout,
		retryConfig: config.RetryConfig,
	}
}

// Get performs a GET request
func (hc *HTTPClient) Get(ctx context.Context, path string, headers map[string]string) (*HTTPResponse, error) {
	return hc.makeRequest(ctx, "GET", path, nil, headers)
}

// Post performs a POST request with JSON body
func (hc *HTTPClient) Post(ctx context.Context, path string, body interface{}, headers map[string]string) (*HTTPResponse, error) {
	return hc.makeRequest(ctx, "POST", path, body, headers)
}

// PostForm performs a POST request with form-encoded body
func (hc *HTTPClient) PostForm(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*HTTPResponse, error) {
	return hc.makeFormRequest(ctx, "POST", path, formData, headers)
}

// Put performs a PUT request with JSON body
func (hc *HTTPClient) Put(ctx context.Context, path string, body interface{}, headers map[string]string) (*HTTPResponse, error) {
	return hc.makeRequest(ctx, "PUT", path, body, headers)
}

// Delete performs a DELETE request
func (hc *HTTPClient) Delete(ctx context.Context, path string, headers map[string]string) (*HTTPResponse, error) {
	return hc.makeRequest(ctx, "DELETE", path, nil, headers)
}

// makeRequest performs the actual HTTP request with retry logic
func (hc *HTTPClient) makeRequest(ctx context.Context, method, path string, body interface{}, headers map[string]string) (*HTTPResponse, error) {
	url := hc.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	return hc.executeWithRetry(ctx, method, url, bodyReader, headers, "application/json")
}

// makeFormRequest performs a form-encoded HTTP request with retry logic
func (hc *HTTPClient) makeFormRequest(ctx context.Context, method, path string, formData map[string]string, headers map[string]string) (*HTTPResponse, error) {
	url := hc.baseURL + path

	// Encode form data
	formValues := ""
	first := true
	for key, value := range formData {
		if !first {
			formValues += "&"
		}
		formValues += fmt.Sprintf("%s=%s", key, value)
		first = false
	}

	bodyReader := bytes.NewReader([]byte(formValues))
	return hc.executeWithRetry(ctx, method, url, bodyReader, headers, "application/x-www-form-urlencoded")
}

// executeWithRetry handles the retry logic for HTTP requests
func (hc *HTTPClient) executeWithRetry(ctx context.Context, method, url string, bodyReader io.Reader, headers map[string]string, contentType string) (*HTTPResponse, error) {
	var lastErr error
	start := time.Now()

	// Store original body for retries
	var originalBody []byte
	if bodyReader != nil {
		var err error
		originalBody, err = io.ReadAll(bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
	}

	for attempt := 0; attempt <= hc.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(float64(hc.retryConfig.RetryDelay) *
				pow(hc.retryConfig.BackoffExp, float64(attempt-1)))

			fmt.Printf("Retry attempt %d/%d after %v delay...\n", attempt+1, hc.retryConfig.MaxRetries, delay)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}

			// Reset body reader for retry
			if originalBody != nil {
				bodyReader = bytes.NewReader(originalBody)
			}
		}

		// Create new body reader for each attempt
		var currentBodyReader io.Reader
		if originalBody != nil {
			currentBodyReader = bytes.NewReader(originalBody)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, currentBodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Add default headers
		for key, value := range hc.headers {
			req.Header.Set(key, value)
		}

		// Add request-specific headers
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		// Set content type
		if originalBody != nil {
			req.Header.Set("Content-Type", contentType)
			req.ContentLength = int64(len(originalBody))
		}

		resp, err := hc.client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < hc.retryConfig.MaxRetries && isRetryableError(err) {
				continue
			}
			return nil, fmt.Errorf("HTTP request failed after %d attempts: %w", attempt+1, err)
		}

		defer resp.Body.Close()

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		response := &HTTPResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       responseBody,
			Duration:   time.Since(start),
		}

		// Success response (2xx)
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return response, nil
		}

		// Handle specific error codes
		if resp.StatusCode == 409 {
			// Conflict - resource already exists
			fmt.Printf("Resource already exists (409): %s\n", string(responseBody))
			return response, nil
		}

		// Check if it's a 500 error with "already in use" message
		if resp.StatusCode == 500 {
			bodyStr := string(responseBody)
			if bytes.Contains(responseBody, []byte("already in use")) || bytes.Contains(responseBody, []byte("already exists")) {
				fmt.Printf("Resource already exists (500): %s\n", bodyStr)
				return response, nil
			}
		}

		// Retry on server errors (5xx) and rate limiting (429)
		if (resp.StatusCode >= 500 || resp.StatusCode == 429) && attempt < hc.retryConfig.MaxRetries {
			lastErr = fmt.Errorf("server error: %d - %s", resp.StatusCode, string(responseBody))
			continue
		}

		// Client error (4xx) - don't retry
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return response, fmt.Errorf("client error: %d - %s", resp.StatusCode, string(responseBody))
		}

		// For other errors, return the response
		return response, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	if lastErr != nil {
		return nil, fmt.Errorf("max retries (%d) exceeded, last error: %w", hc.retryConfig.MaxRetries, lastErr)
	}

	return nil, fmt.Errorf("max retries (%d) exceeded", hc.retryConfig.MaxRetries)
}

// JSON unmarshals the response body as JSON
func (hr *HTTPResponse) JSON(v interface{}) error {
	return json.Unmarshal(hr.Body, v)
}

// String returns the response body as a string
func (hr *HTTPResponse) String() string {
	return string(hr.Body)
}

// IsSuccess returns true if the status code indicates success (2xx)
func (hr *HTTPResponse) IsSuccess() bool {
	return hr.StatusCode >= 200 && hr.StatusCode < 300
}

// IsClientError returns true if the status code indicates client error (4xx)
func (hr *HTTPResponse) IsClientError() bool {
	return hr.StatusCode >= 400 && hr.StatusCode < 500
}

// IsServerError returns true if the status code indicates server error (5xx)
func (hr *HTTPResponse) IsServerError() bool {
	return hr.StatusCode >= 500
}

// Helper functions

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
	// Network errors are typically retryable
	if _, ok := err.(net.Error); ok {
		return true
	}
	return true
}

// pow calculates base^exp
func pow(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	if exp == 1 {
		return base
	}

	result := base
	for i := 1; i < int(exp); i++ {
		result *= base
	}
	return result
}
