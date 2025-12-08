// Package services implements business logic for Splunk API interactions
package services_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"salesforce-splunk-migration/mocks"
	"salesforce-splunk-migration/models"
	"salesforce-splunk-migration/services"
	"salesforce-splunk-migration/utils"
)

func TestNewSplunkService(t *testing.T) {
	t.Run("Success_WithAllParameters", func(t *testing.T) {
		config := &utils.Config{
			Splunk: utils.SplunkConfig{
				URL:            "https://splunk.example.com:8089",
				Username:       "admin",
				Password:       "password",
				RequestTimeout: 30,
				MaxRetries:     3,
				RetryDelay:     5,
			},
		}

		service, err := services.NewSplunkService(config)
		require.NoError(t, err)
		assert.NotNil(t, service)
	})

	t.Run("Success_WithDefaults", func(t *testing.T) {
		config := &utils.Config{
			Splunk: utils.SplunkConfig{
				URL:      "https://splunk.example.com:8089",
				Username: "admin",
				Password: "password",
			},
		}

		service, err := services.NewSplunkService(config)
		require.NoError(t, err)
		assert.NotNil(t, service)
	})

	t.Run("Success_EmptyURL", func(t *testing.T) {
		config := &utils.Config{
			Splunk: utils.SplunkConfig{
				URL:      "",
				Username: "admin",
				Password: "password",
			},
		}

		service, err := services.NewSplunkService(config)
		require.NoError(t, err)
		assert.NotNil(t, service)
	})

	t.Run("Success_CustomTimeout", func(t *testing.T) {
		config := &utils.Config{
			Splunk: utils.SplunkConfig{
				URL:            "https://splunk.example.com:8089",
				Username:       "admin",
				Password:       "password",
				RequestTimeout: 60,
			},
		}

		service, err := services.NewSplunkService(config)
		require.NoError(t, err)
		assert.NotNil(t, service)
	})
}

func TestSplunkService_Authenticate(t *testing.T) {
	t.Run("Success_ValidCredentials", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				authResp := models.AuthResponse{
					SessionKey: "test-session-key-12345",
				}
				body, _ := json.Marshal(authResp)
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       body,
				}, nil
			},
		}

		config := &utils.Config{
			Splunk: utils.SplunkConfig{
				Username: "admin",
				Password: "password",
			},
		}
		service, err := services.NewSplunkServiceWithClient(config, mockClient)
		require.NoError(t, err)

		err = service.Authenticate(context.Background())
		require.NoError(t, err)
	})

	t.Run("Error_UnauthorizedCredentials", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 401,
					Body:       []byte("Unauthorized"),
				}, nil
			},
		}

		config := &utils.Config{
			Splunk: utils.SplunkConfig{
				Username: "admin",
				Password: "wrong",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.Authenticate(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
	})

	t.Run("Error_NetworkError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, fmt.Errorf("network error")
			},
		}

		config := &utils.Config{
			Splunk: utils.SplunkConfig{
				Username: "admin",
				Password: "password",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.Authenticate(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "network error")
	})

	t.Run("Error_InvalidJSON", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       []byte("invalid json"),
				}, nil
			},
		}

		config := &utils.Config{
			Splunk: utils.SplunkConfig{
				Username: "admin",
				Password: "password",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.Authenticate(context.Background())
		require.Error(t, err)
	})
}

func TestSplunkService_CreateIndex(t *testing.T) {
	t.Run("Success_ValidIndexName", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 201,
					Body:       []byte(`{"entry": []}`),
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CreateIndex(context.Background(), "test_index")
		require.NoError(t, err)
	})

	t.Run("Error_EmptyIndexName", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{}
		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CreateIndex(context.Background(), "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "index name cannot be empty")
	})

	t.Run("Error_HTTPError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 400,
					Body:       []byte("Bad Request"),
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CreateIndex(context.Background(), "test_index")
		require.Error(t, err)
	})

	t.Run("Error_NetworkError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, fmt.Errorf("network error")
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CreateIndex(context.Background(), "test_index")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "network error")
	})
}

func TestSplunkService_CreateSalesforceAccount(t *testing.T) {
	t.Run("Success_ValidConfiguration", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 201,
					Body:       []byte(`{"entry": []}`),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName:  "test_account",
				Endpoint:     "https://login.salesforce.com",
				ClientID:     "client_id",
				ClientSecret: "client_secret",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CreateSalesforceAccount(context.Background())
		require.NoError(t, err)
	})

	t.Run("Error_HTTPError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 400,
					Body:       []byte("Bad Request"),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CreateSalesforceAccount(context.Background())
		require.Error(t, err)
	})

	t.Run("Error_NetworkError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, fmt.Errorf("network error")
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CreateSalesforceAccount(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "network error")
	})
}

func TestSplunkService_CreateDataInput(t *testing.T) {
	t.Run("Success_ValidInput", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 201,
					Body:       []byte(`{"entry": []}`),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
			Splunk: utils.SplunkConfig{
				DefaultIndex: "main",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		input := &utils.DataInput{
			Name:   "Account_Input",
			Object: "Account",
		}

		err := service.CreateDataInput(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("Error_NilInput", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{}
		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CreateDataInput(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "data input cannot be nil")
	})

	t.Run("Error_HTTPError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 400,
					Body:       []byte("Bad Request"),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		input := &utils.DataInput{
			Name:   "Account_Input",
			Object: "Account",
		}

		err := service.CreateDataInput(context.Background(), input)
		require.Error(t, err)
	})

	t.Run("Error_NetworkError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, fmt.Errorf("network error")
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		input := &utils.DataInput{
			Name:   "Account_Input",
			Object: "Account",
		}

		err := service.CreateDataInput(context.Background(), input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "network error")
	})
}

func TestSplunkService_ListDataInputs(t *testing.T) {
	t.Run("Success_MultipleInputs", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				response := map[string]interface{}{
					"entry": []interface{}{
						map[string]interface{}{"name": "Account_Input"},
						map[string]interface{}{"name": "Contact_Input"},
					},
				}
				body, _ := json.Marshal(response)
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       body,
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		inputs, err := service.ListDataInputs(context.Background())
		require.NoError(t, err)
		assert.Len(t, inputs, 2)
		assert.Contains(t, inputs, "Account_Input")
		assert.Contains(t, inputs, "Contact_Input")
	})

	t.Run("Success_EmptyList", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				response := map[string]interface{}{
					"entry": []interface{}{},
				}
				body, _ := json.Marshal(response)
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       body,
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		inputs, err := service.ListDataInputs(context.Background())
		require.NoError(t, err)
		assert.Empty(t, inputs)
	})

	t.Run("Error_InvalidJSON", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       []byte("invalid json"),
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		inputs, err := service.ListDataInputs(context.Background())
		require.Error(t, err)
		assert.Nil(t, inputs)
	})

	t.Run("Error_NetworkError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, fmt.Errorf("network error")
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		inputs, err := service.ListDataInputs(context.Background())
		require.Error(t, err)
		assert.Nil(t, inputs)
		assert.Contains(t, err.Error(), "network error")
	})
}

func TestSplunkService_CheckSalesforceAddon(t *testing.T) {
	t.Run("Success_AddonInstalled", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				response := map[string]interface{}{
					"entry": []interface{}{
						map[string]interface{}{
							"name": "Splunk_TA_salesforce",
							"content": map[string]interface{}{
								"disabled": false,
							},
						},
					},
				}
				body, _ := json.Marshal(response)
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       body,
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CheckSalesforceAddon(context.Background())
		require.NoError(t, err)
	})

	t.Run("Error_AddonNotFound", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				response := map[string]interface{}{
					"entry": []interface{}{},
				}
				body, _ := json.Marshal(response)
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       body,
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CheckSalesforceAddon(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Splunk Add-on for Salesforce")
	})

	t.Run("Error_NetworkError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, fmt.Errorf("network error")
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CheckSalesforceAddon(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "network error")
	})

	t.Run("Error_HTTPStatusError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 500,
					Body:       []byte("Internal Server Error"),
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CheckSalesforceAddon(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list installed apps")
	})

	t.Run("Error_JSONParseError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       []byte("invalid json"),
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CheckSalesforceAddon(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse apps list")
	})

	t.Run("Error_AddonDisabled", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				response := map[string]interface{}{
					"entry": []interface{}{
						map[string]interface{}{
							"name": "Splunk_TA_salesforce",
							"content": map[string]interface{}{
								"disabled": true,
								"version":  "4.0.0",
							},
						},
					},
				}
				body, _ := json.Marshal(response)
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       body,
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CheckSalesforceAddon(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "installed but disabled")
		assert.Contains(t, err.Error(), "4.0.0")
	})
}

func TestSplunkService_CheckResponseMessages(t *testing.T) {
	t.Run("Success_NoMessages", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				response := map[string]interface{}{
					"entry": []interface{}{},
				}
				body, _ := json.Marshal(response)
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       body,
				}, nil
			},
		}

		config := &utils.Config{
			Splunk: utils.SplunkConfig{
				IndexName: "test_index",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CreateIndex(context.Background(), "test_index")
		require.NoError(t, err)
	})

	t.Run("Success_WithErrorMessageAlreadyExists", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				response := map[string]interface{}{
					"messages": []interface{}{
						map[string]interface{}{
							"type": "ERROR",
							"text": "Index already exists",
						},
					},
				}
				body, _ := json.Marshal(response)
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       body,
				}, nil
			},
		}

		config := &utils.Config{
			Splunk: utils.SplunkConfig{
				IndexName: "test_index",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CreateIndex(context.Background(), "test_index")
		require.NoError(t, err) // Should not error for "already exists"
	})

	t.Run("Success_InvalidJSONButSuccessStatus", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       []byte("invalid json"),
				}, nil
			},
		}

		config := &utils.Config{
			Splunk: utils.SplunkConfig{
				IndexName: "test_index",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CreateIndex(context.Background(), "test_index")
		require.NoError(t, err) // Should not error if status is success even with invalid JSON
	})

	t.Run("Error_InvalidJSONAndFailureStatus", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 500,
					Body:       []byte("invalid json"),
				}, nil
			},
		}

		config := &utils.Config{
			Splunk: utils.SplunkConfig{
				IndexName: "test_index",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.CreateIndex(context.Background(), "test_index")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse response")
	})
}

func TestSplunkService_CheckIndexExists(t *testing.T) {
	t.Run("Success_IndexExists", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				response := map[string]interface{}{
					"entry": []interface{}{
						map[string]interface{}{"name": "test_index"},
					},
				}
				body, _ := json.Marshal(response)
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       body,
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckIndexExists(context.Background(), "test_index")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Success_IndexNotExists", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 404,
					Body:       []byte("Not Found"),
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckIndexExists(context.Background(), "test_index")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Error_EmptyIndexName", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{}
		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckIndexExists(context.Background(), "")
		require.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "index name cannot be empty")
	})

	t.Run("Error_NetworkError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, fmt.Errorf("network error")
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckIndexExists(context.Background(), "test_index")
		require.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "network error")
	})

	t.Run("Success_InvalidJSON_ReturnsTrue", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       []byte("invalid json"),
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckIndexExists(context.Background(), "test_index")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Success_UnexpectedStatusCode_ReturnsFalse", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 500,
					Body:       []byte("Internal Server Error"),
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckIndexExists(context.Background(), "test_index")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestSplunkService_UpdateIndex(t *testing.T) {
	t.Run("Success_ValidUpdate", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       []byte(`{"entry": []}`),
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.UpdateIndex(context.Background(), "test_index")
		require.NoError(t, err)
	})

	t.Run("Error_EmptyIndexName", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{}
		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.UpdateIndex(context.Background(), "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "index name cannot be empty")
	})

	t.Run("Error_HTTPError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 400,
					Body:       []byte("Bad Request"),
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.UpdateIndex(context.Background(), "test_index")
		require.Error(t, err)
	})

	t.Run("Error_NetworkError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, fmt.Errorf("network error")
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.UpdateIndex(context.Background(), "test_index")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "network error")
	})

	t.Run("Error_InternalServerError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 500,
					Body:       []byte("Internal Server Error"),
				}, nil
			},
		}

		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.UpdateIndex(context.Background(), "test_index")
		require.Error(t, err)
	})
}

func TestSplunkService_CheckSalesforceAccountExists(t *testing.T) {
	t.Run("Success_AccountExists", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				response := map[string]interface{}{
					"entry": []interface{}{
						map[string]interface{}{"name": "test_account"},
					},
				}
				body, _ := json.Marshal(response)
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       body,
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckSalesforceAccountExists(context.Background())
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Success_AccountNotExists", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 404,
					Body:       []byte("Not Found"),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckSalesforceAccountExists(context.Background())
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Error_NetworkError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, fmt.Errorf("network error")
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckSalesforceAccountExists(context.Background())
		require.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "network error")
	})

	t.Run("Success_InvalidJSON_ReturnsTrue", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       []byte("invalid json"),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckSalesforceAccountExists(context.Background())
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Success_500_WithoutNotFoundMessage_ReturnsFalse", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 500,
					Body:       []byte("Internal Server Error"),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckSalesforceAccountExists(context.Background())
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Success_500_WithNotFoundMessage_ReturnsFalse", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 500,
					Body:       []byte("Not Found - [404] Could not find object"),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckSalesforceAccountExists(context.Background())
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestSplunkService_UpdateSalesforceAccount(t *testing.T) {
	t.Run("Success_ValidUpdate", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       []byte(`{"entry": []}`),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName:  "test_account",
				Endpoint:     "https://login.salesforce.com",
				ClientID:     "client_id",
				ClientSecret: "client_secret",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.UpdateSalesforceAccount(context.Background())
		require.NoError(t, err)
	})

	t.Run("Error_HTTPError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 400,
					Body:       []byte("Bad Request"),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.UpdateSalesforceAccount(context.Background())
		require.Error(t, err)
	})

	t.Run("Error_NetworkError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, fmt.Errorf("network error")
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.UpdateSalesforceAccount(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "network error")
	})

	t.Run("Error_InternalServerError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 500,
					Body:       []byte("Internal Server Error"),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.UpdateSalesforceAccount(context.Background())
		require.Error(t, err)
	})
}

func TestSplunkService_CheckDataInputExists(t *testing.T) {
	t.Run("Success_DataInputExists", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				response := map[string]interface{}{
					"entry": []interface{}{
						map[string]interface{}{"name": "Account_Input"},
					},
				}
				body, _ := json.Marshal(response)
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       body,
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckDataInputExists(context.Background(), "Account_Input")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Success_DataInputNotExists", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 404,
					Body:       []byte("Not Found"),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckDataInputExists(context.Background(), "Account_Input")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Error_NetworkError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, fmt.Errorf("network error")
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckDataInputExists(context.Background(), "Account_Input")
		require.Error(t, err)
		assert.False(t, exists)
		assert.Contains(t, err.Error(), "network error")
	})

	t.Run("Success_InvalidJSON_ReturnsTrue", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       []byte("invalid json"),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckDataInputExists(context.Background(), "Account_Input")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Success_500_WithoutNotFoundMessage_ReturnsFalse", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 500,
					Body:       []byte("Internal Server Error"),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckDataInputExists(context.Background(), "Account_Input")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Success_500_WithNotFoundMessage_ReturnsFalse", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			GetFunc: func(ctx context.Context, path string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 500,
					Body:       []byte("Not Found - [404]"),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		exists, err := service.CheckDataInputExists(context.Background(), "Account_Input")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestSplunkService_UpdateDataInput(t *testing.T) {
	t.Run("Success_ValidUpdate", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       []byte(`{"entry": []}`),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
			Splunk: utils.SplunkConfig{
				DefaultIndex: "main",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		input := &utils.DataInput{
			Name:   "Account_Input",
			Object: "Account",
		}

		err := service.UpdateDataInput(context.Background(), input)
		require.NoError(t, err)
	})

	t.Run("Error_NilInput", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{}
		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		err := service.UpdateDataInput(context.Background(), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "data input cannot be nil")
	})

	t.Run("Error_HTTPError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 400,
					Body:       []byte("Bad Request"),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		input := &utils.DataInput{
			Name:   "Account_Input",
			Object: "Account",
		}

		err := service.UpdateDataInput(context.Background(), input)
		require.Error(t, err)
	})

	t.Run("Error_NetworkError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return nil, fmt.Errorf("network error")
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		input := &utils.DataInput{
			Name:   "Account_Input",
			Object: "Account",
		}

		err := service.UpdateDataInput(context.Background(), input)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "network error")
	})

	t.Run("Error_InternalServerError", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				return &utils.HTTPResponse{
					StatusCode: 500,
					Body:       []byte("Internal Server Error"),
				}, nil
			},
		}

		config := &utils.Config{
			Salesforce: utils.SalesforceConfig{
				AccountName: "test_account",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		input := &utils.DataInput{
			Name:   "Account_Input",
			Object: "Account",
		}

		err := service.UpdateDataInput(context.Background(), input)
		require.Error(t, err)
	})
}

func TestSplunkService_GetAuthToken(t *testing.T) {
	t.Run("Success_ReturnsToken", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{
			PostFormFunc: func(ctx context.Context, path string, formData map[string]string, headers map[string]string) (*utils.HTTPResponse, error) {
				authResponse := models.AuthResponse{
					SessionKey: "test-token-12345",
				}
				body, _ := json.Marshal(authResponse)
				return &utils.HTTPResponse{
					StatusCode: 200,
					Body:       body,
				}, nil
			},
		}

		config := &utils.Config{
			Splunk: utils.SplunkConfig{
				Username: "admin",
				Password: "password",
			},
		}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		// Authenticate first
		err := service.Authenticate(context.Background())
		require.NoError(t, err)

		// Get token
		token := service.GetAuthToken()
		assert.Equal(t, "test-token-12345", token)
	})

	t.Run("Success_EmptyTokenBeforeAuth", func(t *testing.T) {
		mockClient := &mocks.MockHTTPClient{}
		config := &utils.Config{}
		service, _ := services.NewSplunkServiceWithClient(config, mockClient)

		token := service.GetAuthToken()
		assert.Empty(t, token)
	})
}
