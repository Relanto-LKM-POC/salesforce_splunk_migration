// Package models contains request/response data structures
package models_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"salesforce-splunk-migration/models"
)

func TestAuthRequest_Validate(t *testing.T) {
	t.Run("Success_ValidCredentials", func(t *testing.T) {
		request := models.AuthRequest{
			Username: "admin",
			Password: "password123",
		}
		err := request.Validate()
		assert.NoError(t, err)
	})

	t.Run("Error_EmptyUsername", func(t *testing.T) {
		request := models.AuthRequest{
			Username: "",
			Password: "password123",
		}
		err := request.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "username is required")
	})

	t.Run("Error_WhitespaceUsername", func(t *testing.T) {
		request := models.AuthRequest{
			Username: "   ",
			Password: "password123",
		}
		err := request.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "username is required")
	})

	t.Run("Error_EmptyPassword", func(t *testing.T) {
		request := models.AuthRequest{
			Username: "admin",
			Password: "",
		}
		err := request.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "password is required")
	})

	t.Run("Error_WhitespacePassword", func(t *testing.T) {
		request := models.AuthRequest{
			Username: "admin",
			Password: "   ",
		}
		err := request.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "password is required")
	})

	t.Run("Error_BothEmpty", func(t *testing.T) {
		request := models.AuthRequest{
			Username: "",
			Password: "",
		}
		err := request.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "username is required")
	})
}

func TestIndexRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request models.IndexRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid index request",
			request: models.IndexRequest{
				Name:     "salesforce_data",
				DataType: "event",
			},
			wantErr: false,
		},
		{
			name: "valid index with metric datatype",
			request: models.IndexRequest{
				Name:     "salesforce_metrics",
				DataType: "metric",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			request: models.IndexRequest{
				Name:     "",
				DataType: "event",
			},
			wantErr: true,
			errMsg:  "index name is required",
		},
		{
			name: "name too long",
			request: models.IndexRequest{
				Name:     string(make([]byte, 256)),
				DataType: "event",
			},
			wantErr: true,
			errMsg:  "index name must be 255 characters or less",
		},
		{
			name: "invalid datatype",
			request: models.IndexRequest{
				Name:     "test_index",
				DataType: "invalid",
			},
			wantErr: true,
			errMsg:  "datatype must be 'event' or 'metric'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("IndexRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("IndexRequest.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestSalesforceAccountRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request models.SalesforceAccountRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid oauth2 account",
			request: models.SalesforceAccountRequest{
				Name:                         "test_account",
				Endpoint:                     "https://login.salesforce.com",
				AuthType:                     "oauth2",
				ClientIDOAuthCredentials:     "client123",
				ClientSecretOAuthCredentials: "secret456",
			},
			wantErr: false,
		},
		{
			name: "valid basic auth account",
			request: models.SalesforceAccountRequest{
				Name:     "test_account",
				Endpoint: "https://login.salesforce.com",
				AuthType: "basic",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			request: models.SalesforceAccountRequest{
				Name:     "",
				Endpoint: "https://login.salesforce.com",
				AuthType: "oauth2",
			},
			wantErr: true,
			errMsg:  "account name is required",
		},
		{
			name: "empty endpoint",
			request: models.SalesforceAccountRequest{
				Name:     "test",
				Endpoint: "",
				AuthType: "oauth2",
			},
			wantErr: true,
			errMsg:  "endpoint is required",
		},
		{
			name: "endpoint without https",
			request: models.SalesforceAccountRequest{
				Name:     "test",
				Endpoint: "http://login.salesforce.com",
				AuthType: "oauth2",
			},
			wantErr: true,
			errMsg:  "endpoint must start with https://",
		},
		{
			name: "empty auth type",
			request: models.SalesforceAccountRequest{
				Name:     "test",
				Endpoint: "https://login.salesforce.com",
				AuthType: "",
			},
			wantErr: true,
			errMsg:  "auth_type is required",
		},
		{
			name: "invalid auth type",
			request: models.SalesforceAccountRequest{
				Name:     "test",
				Endpoint: "https://login.salesforce.com",
				AuthType: "invalid",
			},
			wantErr: true,
			errMsg:  "auth_type must be 'basic' or 'oauth2'",
		},
		{
			name: "oauth2 missing client id",
			request: models.SalesforceAccountRequest{
				Name:                         "test",
				Endpoint:                     "https://login.salesforce.com",
				AuthType:                     "oauth2",
				ClientSecretOAuthCredentials: "secret",
			},
			wantErr: true,
			errMsg:  "client_id is required for oauth2 auth",
		},
		{
			name: "oauth2 missing client secret",
			request: models.SalesforceAccountRequest{
				Name:                     "test",
				Endpoint:                 "https://login.salesforce.com",
				AuthType:                 "oauth2",
				ClientIDOAuthCredentials: "client",
			},
			wantErr: true,
			errMsg:  "client_secret is required for oauth2 auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SalesforceAccountRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("SalesforceAccountRequest.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestDataInputRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		request models.DataInputRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid data input",
			request: models.DataInputRequest{
				Name:      "Account_Input",
				Account:   "test_account",
				Object:    "Account",
				StartDate: "2024-01-01",
				Interval:  300,
				Delay:     60,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			request: models.DataInputRequest{
				Name:    "",
				Account: "test",
				Object:  "Account",
			},
			wantErr: true,
			errMsg:  "data input name is required",
		},
		{
			name: "empty account",
			request: models.DataInputRequest{
				Name:    "test",
				Account: "",
				Object:  "Account",
			},
			wantErr: true,
			errMsg:  "account is required",
		},
		{
			name: "empty object",
			request: models.DataInputRequest{
				Name:    "test",
				Account: "account",
				Object:  "",
			},
			wantErr: true,
			errMsg:  "object is required",
		},
		{
			name: "negative interval",
			request: models.DataInputRequest{
				Name:     "test",
				Account:  "account",
				Object:   "Account",
				Interval: -1,
			},
			wantErr: true,
			errMsg:  "interval must be non-negative",
		},
		{
			name: "negative delay",
			request: models.DataInputRequest{
				Name:    "test",
				Account: "account",
				Object:  "Account",
				Delay:   -5,
			},
			wantErr: true,
			errMsg:  "delay must be non-negative",
		},
		{
			name: "invalid start date format",
			request: models.DataInputRequest{
				Name:      "test",
				Account:   "account",
				Object:    "Account",
				StartDate: "01-01-2024",
			},
			wantErr: true,
			errMsg:  "start_date must be in YYYY-MM-DD format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DataInputRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !containsSubstring(err.Error(), tt.errMsg) {
				t.Errorf("DataInputRequest.Validate() error message = %v, want to contain %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func containsSubstring(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(str) > len(substr) &&
		(str[:len(substr)] == substr || str[len(str)-len(substr):] == substr ||
			findInString(str, substr)))
}

func findInString(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
