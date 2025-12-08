package mocks

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"salesforce-splunk-migration/utils"
)

func TestMockSplunkService_Authenticate(t *testing.T) {
	t.Run("Success_WithCustomFunc", func(t *testing.T) {
		mock := &MockSplunkService{
			AuthenticateFunc: func(ctx context.Context) error {
				return nil
			},
		}

		err := mock.Authenticate(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.AuthenticateCalls)
	})

	t.Run("Success_WithoutCustomFunc", func(t *testing.T) {
		mock := &MockSplunkService{}

		err := mock.Authenticate(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.AuthenticateCalls)
	})

	t.Run("Error_CustomFuncReturnsError", func(t *testing.T) {
		expectedErr := errors.New("authentication failed")
		mock := &MockSplunkService{
			AuthenticateFunc: func(ctx context.Context) error {
				return expectedErr
			},
		}

		err := mock.Authenticate(context.Background())
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, mock.AuthenticateCalls)
	})

	t.Run("NilContext_HandledGracefully", func(t *testing.T) {
		mock := &MockSplunkService{}

		err := mock.Authenticate(nil)
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.AuthenticateCalls)
	})

	t.Run("MultipleCalls_IncrementCounter", func(t *testing.T) {
		mock := &MockSplunkService{}

		for i := 1; i <= 5; i++ {
			err := mock.Authenticate(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, i, mock.AuthenticateCalls)
		}
	})
}

func TestMockSplunkService_CheckSalesforceAddon(t *testing.T) {
	t.Run("Success_WithCustomFunc", func(t *testing.T) {
		mock := &MockSplunkService{
			CheckSalesforceAddonFunc: func(ctx context.Context) error {
				return nil
			},
		}

		err := mock.CheckSalesforceAddon(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CheckSalesforceAddonCalls)
	})

	t.Run("Success_WithoutCustomFunc", func(t *testing.T) {
		mock := &MockSplunkService{}

		err := mock.CheckSalesforceAddon(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CheckSalesforceAddonCalls)
	})

	t.Run("Error_AddonNotFound", func(t *testing.T) {
		expectedErr := errors.New("addon not found")
		mock := &MockSplunkService{
			CheckSalesforceAddonFunc: func(ctx context.Context) error {
				return expectedErr
			},
		}

		err := mock.CheckSalesforceAddon(context.Background())
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, mock.CheckSalesforceAddonCalls)
	})

	t.Run("NilContext_HandledGracefully", func(t *testing.T) {
		mock := &MockSplunkService{}

		err := mock.CheckSalesforceAddon(nil)
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CheckSalesforceAddonCalls)
	})

	t.Run("MultipleCalls_IncrementCounter", func(t *testing.T) {
		mock := &MockSplunkService{}

		for i := 1; i <= 3; i++ {
			err := mock.CheckSalesforceAddon(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, i, mock.CheckSalesforceAddonCalls)
		}
	})
}

func TestMockSplunkService_CreateIndex(t *testing.T) {
	t.Run("Success_WithCustomFunc", func(t *testing.T) {
		mock := &MockSplunkService{
			CreateIndexFunc: func(ctx context.Context, indexName string) error {
				assert.Equal(t, "test_index", indexName)
				return nil
			},
		}

		err := mock.CreateIndex(context.Background(), "test_index")
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CreateIndexCalls)
	})

	t.Run("Success_WithoutCustomFunc", func(t *testing.T) {
		mock := &MockSplunkService{}

		err := mock.CreateIndex(context.Background(), "test_index")
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CreateIndexCalls)
	})

	t.Run("Error_IndexAlreadyExists", func(t *testing.T) {
		expectedErr := errors.New("index already exists")
		mock := &MockSplunkService{
			CreateIndexFunc: func(ctx context.Context, indexName string) error {
				return expectedErr
			},
		}

		err := mock.CreateIndex(context.Background(), "existing_index")
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, mock.CreateIndexCalls)
	})

	t.Run("EmptyIndexName_PassedToFunc", func(t *testing.T) {
		mock := &MockSplunkService{
			CreateIndexFunc: func(ctx context.Context, indexName string) error {
				assert.Empty(t, indexName)
				return errors.New("index name cannot be empty")
			},
		}

		err := mock.CreateIndex(context.Background(), "")
		assert.Error(t, err)
		assert.Equal(t, 1, mock.CreateIndexCalls)
	})

	t.Run("NilContext_HandledGracefully", func(t *testing.T) {
		mock := &MockSplunkService{}

		err := mock.CreateIndex(nil, "test_index")
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CreateIndexCalls)
	})

	t.Run("MultipleCalls_WithDifferentIndexNames", func(t *testing.T) {
		mock := &MockSplunkService{}
		indexNames := []string{"index1", "index2", "index3"}

		for i, name := range indexNames {
			err := mock.CreateIndex(context.Background(), name)
			assert.NoError(t, err)
			assert.Equal(t, i+1, mock.CreateIndexCalls)
		}
	})
}

func TestMockSplunkService_CreateSalesforceAccount(t *testing.T) {
	t.Run("Success_WithCustomFunc", func(t *testing.T) {
		mock := &MockSplunkService{
			CreateSalesforceAccountFunc: func(ctx context.Context) error {
				return nil
			},
		}

		err := mock.CreateSalesforceAccount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CreateSalesforceAccountCalls)
	})

	t.Run("Success_WithoutCustomFunc", func(t *testing.T) {
		mock := &MockSplunkService{}

		err := mock.CreateSalesforceAccount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CreateSalesforceAccountCalls)
	})

	t.Run("Error_AccountCreationFailed", func(t *testing.T) {
		expectedErr := errors.New("account creation failed")
		mock := &MockSplunkService{
			CreateSalesforceAccountFunc: func(ctx context.Context) error {
				return expectedErr
			},
		}

		err := mock.CreateSalesforceAccount(context.Background())
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, mock.CreateSalesforceAccountCalls)
	})

	t.Run("NilContext_HandledGracefully", func(t *testing.T) {
		mock := &MockSplunkService{}

		err := mock.CreateSalesforceAccount(nil)
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CreateSalesforceAccountCalls)
	})

	t.Run("MultipleCalls_IncrementCounter", func(t *testing.T) {
		mock := &MockSplunkService{}

		for i := 1; i <= 4; i++ {
			err := mock.CreateSalesforceAccount(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, i, mock.CreateSalesforceAccountCalls)
		}
	})
}

func TestMockSplunkService_CreateDataInput(t *testing.T) {
	t.Run("Success_WithCustomFunc", func(t *testing.T) {
		input := &utils.DataInput{
			Name:   "Test_Input",
			Object: "Account",
		}

		mock := &MockSplunkService{
			CreateDataInputFunc: func(ctx context.Context, di *utils.DataInput) error {
				assert.Equal(t, input.Name, di.Name)
				assert.Equal(t, input.Object, di.Object)
				return nil
			},
		}

		err := mock.CreateDataInput(context.Background(), input)
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CreateDataInputCalls)
	})

	t.Run("Success_WithoutCustomFunc", func(t *testing.T) {
		mock := &MockSplunkService{}
		input := &utils.DataInput{
			Name:   "Test_Input",
			Object: "Contact",
		}

		err := mock.CreateDataInput(context.Background(), input)
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CreateDataInputCalls)
	})

	t.Run("Error_InvalidDataInput", func(t *testing.T) {
		expectedErr := errors.New("invalid data input")
		mock := &MockSplunkService{
			CreateDataInputFunc: func(ctx context.Context, di *utils.DataInput) error {
				return expectedErr
			},
		}

		input := &utils.DataInput{}
		err := mock.CreateDataInput(context.Background(), input)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, mock.CreateDataInputCalls)
	})

	t.Run("NilDataInput_PassedToFunc", func(t *testing.T) {
		mock := &MockSplunkService{
			CreateDataInputFunc: func(ctx context.Context, di *utils.DataInput) error {
				assert.Nil(t, di)
				return errors.New("data input cannot be nil")
			},
		}

		err := mock.CreateDataInput(context.Background(), nil)
		assert.Error(t, err)
		assert.Equal(t, 1, mock.CreateDataInputCalls)
	})

	t.Run("NilContext_HandledGracefully", func(t *testing.T) {
		mock := &MockSplunkService{}
		input := &utils.DataInput{Name: "Test"}

		err := mock.CreateDataInput(nil, input)
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CreateDataInputCalls)
	})

	t.Run("MultipleCalls_WithDifferentInputs", func(t *testing.T) {
		mock := &MockSplunkService{}
		inputs := []*utils.DataInput{
			{Name: "Input1", Object: "Account"},
			{Name: "Input2", Object: "Contact"},
			{Name: "Input3", Object: "Lead"},
		}

		for i, input := range inputs {
			err := mock.CreateDataInput(context.Background(), input)
			assert.NoError(t, err)
			assert.Equal(t, i+1, mock.CreateDataInputCalls)
		}
	})
}

func TestMockSplunkService_ListDataInputs(t *testing.T) {
	t.Run("Success_WithCustomFunc_ReturnsInputs", func(t *testing.T) {
		expectedInputs := []string{"Input1", "Input2", "Input3"}
		mock := &MockSplunkService{
			ListDataInputsFunc: func(ctx context.Context) ([]string, error) {
				return expectedInputs, nil
			},
		}

		inputs, err := mock.ListDataInputs(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expectedInputs, inputs)
		assert.Equal(t, 1, mock.ListDataInputsCalls)
	})

	t.Run("Success_WithoutCustomFunc_ReturnsEmptySlice", func(t *testing.T) {
		mock := &MockSplunkService{}

		inputs, err := mock.ListDataInputs(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, inputs)
		assert.Empty(t, inputs)
		assert.Equal(t, 1, mock.ListDataInputsCalls)
	})

	t.Run("Success_WithCustomFunc_ReturnsEmptyList", func(t *testing.T) {
		mock := &MockSplunkService{
			ListDataInputsFunc: func(ctx context.Context) ([]string, error) {
				return []string{}, nil
			},
		}

		inputs, err := mock.ListDataInputs(context.Background())
		assert.NoError(t, err)
		assert.Empty(t, inputs)
		assert.Equal(t, 1, mock.ListDataInputsCalls)
	})

	t.Run("Error_ListingFailed", func(t *testing.T) {
		expectedErr := errors.New("failed to list data inputs")
		mock := &MockSplunkService{
			ListDataInputsFunc: func(ctx context.Context) ([]string, error) {
				return nil, expectedErr
			},
		}

		inputs, err := mock.ListDataInputs(context.Background())
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, inputs)
		assert.Equal(t, 1, mock.ListDataInputsCalls)
	})

	t.Run("NilContext_HandledGracefully", func(t *testing.T) {
		mock := &MockSplunkService{}

		inputs, err := mock.ListDataInputs(nil)
		assert.NoError(t, err)
		assert.Empty(t, inputs)
		assert.Equal(t, 1, mock.ListDataInputsCalls)
	})

	t.Run("MultipleCalls_IncrementCounter", func(t *testing.T) {
		mock := &MockSplunkService{
			ListDataInputsFunc: func(ctx context.Context) ([]string, error) {
				return []string{"Input1"}, nil
			},
		}

		for i := 1; i <= 3; i++ {
			inputs, err := mock.ListDataInputs(context.Background())
			assert.NoError(t, err)
			assert.Len(t, inputs, 1)
			assert.Equal(t, i, mock.ListDataInputsCalls)
		}
	})
}

func TestMockSplunkService_Reset(t *testing.T) {
	t.Run("Success_ResetsAllCounters", func(t *testing.T) {
		mock := &MockSplunkService{}

		// Call all methods multiple times
		_ = mock.Authenticate(context.Background())
		_ = mock.Authenticate(context.Background())
		_ = mock.CheckSalesforceAddon(context.Background())
		_ = mock.CreateIndex(context.Background(), "test")
		_ = mock.CreateIndex(context.Background(), "test2")
		_ = mock.CreateIndex(context.Background(), "test3")
		_ = mock.CreateSalesforceAccount(context.Background())
		_ = mock.CreateDataInput(context.Background(), &utils.DataInput{})
		_, _ = mock.ListDataInputs(context.Background())
		_, _ = mock.ListDataInputs(context.Background())

		// Verify counters are incremented
		assert.Equal(t, 2, mock.AuthenticateCalls)
		assert.Equal(t, 1, mock.CheckSalesforceAddonCalls)
		assert.Equal(t, 3, mock.CreateIndexCalls)
		assert.Equal(t, 1, mock.CreateSalesforceAccountCalls)
		assert.Equal(t, 1, mock.CreateDataInputCalls)
		assert.Equal(t, 2, mock.ListDataInputsCalls)

		// Reset all counters
		mock.Reset()

		// Verify all counters are zero
		assert.Equal(t, 0, mock.AuthenticateCalls)
		assert.Equal(t, 0, mock.CheckSalesforceAddonCalls)
		assert.Equal(t, 0, mock.CreateIndexCalls)
		assert.Equal(t, 0, mock.CreateSalesforceAccountCalls)
		assert.Equal(t, 0, mock.CreateDataInputCalls)
		assert.Equal(t, 0, mock.ListDataInputsCalls)
	})

	t.Run("Success_MultipleResets", func(t *testing.T) {
		mock := &MockSplunkService{}

		// Call some methods
		_ = mock.Authenticate(context.Background())
		_ = mock.CreateIndex(context.Background(), "test")

		// First reset
		mock.Reset()
		assert.Equal(t, 0, mock.AuthenticateCalls)
		assert.Equal(t, 0, mock.CreateIndexCalls)

		// Call again
		_ = mock.Authenticate(context.Background())
		assert.Equal(t, 1, mock.AuthenticateCalls)

		// Second reset
		mock.Reset()
		assert.Equal(t, 0, mock.AuthenticateCalls)
	})

	t.Run("Success_ResetWithoutPriorCalls", func(t *testing.T) {
		mock := &MockSplunkService{}

		// Reset without any prior calls
		mock.Reset()

		// Verify all counters are zero
		assert.Equal(t, 0, mock.AuthenticateCalls)
		assert.Equal(t, 0, mock.CheckSalesforceAddonCalls)
		assert.Equal(t, 0, mock.CreateIndexCalls)
		assert.Equal(t, 0, mock.CreateSalesforceAccountCalls)
		assert.Equal(t, 0, mock.CreateDataInputCalls)
		assert.Equal(t, 0, mock.ListDataInputsCalls)
	})
}

func TestMockSplunkService_Integration(t *testing.T) {
	t.Run("Success_CompleteWorkflow", func(t *testing.T) {
		mock := &MockSplunkService{
			AuthenticateFunc: func(ctx context.Context) error {
				return nil
			},
			CheckSalesforceAddonFunc: func(ctx context.Context) error {
				return nil
			},
			CreateIndexFunc: func(ctx context.Context, indexName string) error {
				require.NotEmpty(t, indexName)
				return nil
			},
			CreateSalesforceAccountFunc: func(ctx context.Context) error {
				return nil
			},
			CreateDataInputFunc: func(ctx context.Context, input *utils.DataInput) error {
				require.NotNil(t, input)
				return nil
			},
			ListDataInputsFunc: func(ctx context.Context) ([]string, error) {
				return []string{"Input1"}, nil
			},
		}

		ctx := context.Background()

		// Execute complete workflow
		err := mock.Authenticate(ctx)
		assert.NoError(t, err)

		err = mock.CheckSalesforceAddon(ctx)
		assert.NoError(t, err)

		err = mock.CreateIndex(ctx, "salesforce_index")
		assert.NoError(t, err)

		err = mock.CreateSalesforceAccount(ctx)
		assert.NoError(t, err)

		err = mock.CreateDataInput(ctx, &utils.DataInput{Name: "Test", Object: "Account"})
		assert.NoError(t, err)

		inputs, err := mock.ListDataInputs(ctx)
		assert.NoError(t, err)
		assert.Len(t, inputs, 1)

		// Verify all calls were tracked
		assert.Equal(t, 1, mock.AuthenticateCalls)
		assert.Equal(t, 1, mock.CheckSalesforceAddonCalls)
		assert.Equal(t, 1, mock.CreateIndexCalls)
		assert.Equal(t, 1, mock.CreateSalesforceAccountCalls)
		assert.Equal(t, 1, mock.CreateDataInputCalls)
		assert.Equal(t, 1, mock.ListDataInputsCalls)
	})

	t.Run("Error_WorkflowFailsAtAuthentication", func(t *testing.T) {
		expectedErr := errors.New("authentication failed")
		mock := &MockSplunkService{
			AuthenticateFunc: func(ctx context.Context) error {
				return expectedErr
			},
		}

		err := mock.Authenticate(context.Background())
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, mock.AuthenticateCalls)

		// Other methods not called
		assert.Equal(t, 0, mock.CheckSalesforceAddonCalls)
		assert.Equal(t, 0, mock.CreateIndexCalls)
	})
}

func TestMockSplunkService_GetAuthToken(t *testing.T) {
	t.Run("Success_WithCustomFunc", func(t *testing.T) {
		expectedToken := "custom-test-token"
		mock := &MockSplunkService{
			GetAuthTokenFunc: func() string {
				return expectedToken
			},
		}

		token := mock.GetAuthToken()
		assert.Equal(t, expectedToken, token)
		assert.Equal(t, 1, mock.GetAuthTokenCalls)
	})

	t.Run("Success_WithAuthTokenValue", func(t *testing.T) {
		expectedToken := "mock-auth-token-123"
		mock := &MockSplunkService{
			AuthTokenValue: expectedToken,
		}

		token := mock.GetAuthToken()
		assert.Equal(t, expectedToken, token)
		assert.Equal(t, 1, mock.GetAuthTokenCalls)
	})

	t.Run("Success_WithoutCustomFunc_ReturnsDefaultToken", func(t *testing.T) {
		mock := &MockSplunkService{}

		token := mock.GetAuthToken()
		assert.Equal(t, "mock-token", token)
		assert.Equal(t, 1, mock.GetAuthTokenCalls)
	})

	t.Run("Success_MultipleCalls_IncrementCounter", func(t *testing.T) {
		mock := &MockSplunkService{
			AuthTokenValue: "test-token",
		}

		for i := 1; i <= 5; i++ {
			token := mock.GetAuthToken()
			assert.Equal(t, "test-token", token)
			assert.Equal(t, i, mock.GetAuthTokenCalls)
		}
	})

	t.Run("Success_EmptyAuthTokenValue", func(t *testing.T) {
		mock := &MockSplunkService{
			AuthTokenValue: "",
		}

		token := mock.GetAuthToken()
		assert.Equal(t, "mock-token", token) // Falls back to default
		assert.Equal(t, 1, mock.GetAuthTokenCalls)
	})
}

func TestMockSplunkService_CheckIndexExists(t *testing.T) {
	t.Run("Success_IndexExists", func(t *testing.T) {
		mock := &MockSplunkService{
			CheckIndexExistsFunc: func(ctx context.Context, indexName string) (bool, error) {
				assert.Equal(t, "test_index", indexName)
				return true, nil
			},
		}

		exists, err := mock.CheckIndexExists(context.Background(), "test_index")
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, 1, mock.CheckIndexExistsCalls)
	})

	t.Run("Success_IndexDoesNotExist", func(t *testing.T) {
		mock := &MockSplunkService{
			CheckIndexExistsFunc: func(ctx context.Context, indexName string) (bool, error) {
				return false, nil
			},
		}

		exists, err := mock.CheckIndexExists(context.Background(), "nonexistent")
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.Equal(t, 1, mock.CheckIndexExistsCalls)
	})

	t.Run("Success_WithoutCustomFunc_ReturnsFalse", func(t *testing.T) {
		mock := &MockSplunkService{}

		exists, err := mock.CheckIndexExists(context.Background(), "test_index")
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.Equal(t, 1, mock.CheckIndexExistsCalls)
	})

	t.Run("Error_CheckFailed", func(t *testing.T) {
		expectedErr := errors.New("check failed")
		mock := &MockSplunkService{
			CheckIndexExistsFunc: func(ctx context.Context, indexName string) (bool, error) {
				return false, expectedErr
			},
		}

		exists, err := mock.CheckIndexExists(context.Background(), "test_index")
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.False(t, exists)
		assert.Equal(t, 1, mock.CheckIndexExistsCalls)
	})

	t.Run("Success_MultipleCalls_WithDifferentIndexNames", func(t *testing.T) {
		mock := &MockSplunkService{
			CheckIndexExistsFunc: func(ctx context.Context, indexName string) (bool, error) {
				return indexName == "existing_index", nil
			},
		}

		indexNames := []string{"existing_index", "nonexistent", "another_index"}
		for i, name := range indexNames {
			exists, err := mock.CheckIndexExists(context.Background(), name)
			assert.NoError(t, err)
			assert.Equal(t, name == "existing_index", exists)
			assert.Equal(t, i+1, mock.CheckIndexExistsCalls)
		}
	})
}

func TestMockSplunkService_UpdateIndex(t *testing.T) {
	t.Run("Success_WithCustomFunc", func(t *testing.T) {
		mock := &MockSplunkService{
			UpdateIndexFunc: func(ctx context.Context, indexName string) error {
				assert.Equal(t, "test_index", indexName)
				return nil
			},
		}

		err := mock.UpdateIndex(context.Background(), "test_index")
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.UpdateIndexCalls)
	})

	t.Run("Success_WithoutCustomFunc", func(t *testing.T) {
		mock := &MockSplunkService{}

		err := mock.UpdateIndex(context.Background(), "test_index")
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.UpdateIndexCalls)
	})

	t.Run("Error_UpdateFailed", func(t *testing.T) {
		expectedErr := errors.New("update failed")
		mock := &MockSplunkService{
			UpdateIndexFunc: func(ctx context.Context, indexName string) error {
				return expectedErr
			},
		}

		err := mock.UpdateIndex(context.Background(), "test_index")
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, mock.UpdateIndexCalls)
	})

	t.Run("Success_MultipleCalls_IncrementCounter", func(t *testing.T) {
		mock := &MockSplunkService{}

		for i := 1; i <= 3; i++ {
			err := mock.UpdateIndex(context.Background(), "test_index")
			assert.NoError(t, err)
			assert.Equal(t, i, mock.UpdateIndexCalls)
		}
	})
}

func TestMockSplunkService_CheckSalesforceAccountExists(t *testing.T) {
	t.Run("Success_AccountExists", func(t *testing.T) {
		mock := &MockSplunkService{
			CheckSalesforceAccountExistsFunc: func(ctx context.Context) (bool, error) {
				return true, nil
			},
		}

		exists, err := mock.CheckSalesforceAccountExists(context.Background())
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, 1, mock.CheckSalesforceAccountExistsCalls)
	})

	t.Run("Success_AccountDoesNotExist", func(t *testing.T) {
		mock := &MockSplunkService{
			CheckSalesforceAccountExistsFunc: func(ctx context.Context) (bool, error) {
				return false, nil
			},
		}

		exists, err := mock.CheckSalesforceAccountExists(context.Background())
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.Equal(t, 1, mock.CheckSalesforceAccountExistsCalls)
	})

	t.Run("Success_WithoutCustomFunc_ReturnsFalse", func(t *testing.T) {
		mock := &MockSplunkService{}

		exists, err := mock.CheckSalesforceAccountExists(context.Background())
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.Equal(t, 1, mock.CheckSalesforceAccountExistsCalls)
	})

	t.Run("Error_CheckFailed", func(t *testing.T) {
		expectedErr := errors.New("check failed")
		mock := &MockSplunkService{
			CheckSalesforceAccountExistsFunc: func(ctx context.Context) (bool, error) {
				return false, expectedErr
			},
		}

		exists, err := mock.CheckSalesforceAccountExists(context.Background())
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.False(t, exists)
		assert.Equal(t, 1, mock.CheckSalesforceAccountExistsCalls)
	})

	t.Run("Success_MultipleCalls_IncrementCounter", func(t *testing.T) {
		mock := &MockSplunkService{}

		for i := 1; i <= 4; i++ {
			exists, err := mock.CheckSalesforceAccountExists(context.Background())
			assert.NoError(t, err)
			assert.False(t, exists)
			assert.Equal(t, i, mock.CheckSalesforceAccountExistsCalls)
		}
	})
}

func TestMockSplunkService_UpdateSalesforceAccount(t *testing.T) {
	t.Run("Success_WithCustomFunc", func(t *testing.T) {
		mock := &MockSplunkService{
			UpdateSalesforceAccountFunc: func(ctx context.Context) error {
				return nil
			},
		}

		err := mock.UpdateSalesforceAccount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.UpdateSalesforceAccountCalls)
	})

	t.Run("Success_WithoutCustomFunc", func(t *testing.T) {
		mock := &MockSplunkService{}

		err := mock.UpdateSalesforceAccount(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.UpdateSalesforceAccountCalls)
	})

	t.Run("Error_UpdateFailed", func(t *testing.T) {
		expectedErr := errors.New("update account failed")
		mock := &MockSplunkService{
			UpdateSalesforceAccountFunc: func(ctx context.Context) error {
				return expectedErr
			},
		}

		err := mock.UpdateSalesforceAccount(context.Background())
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, mock.UpdateSalesforceAccountCalls)
	})

	t.Run("Success_MultipleCalls_IncrementCounter", func(t *testing.T) {
		mock := &MockSplunkService{}

		for i := 1; i <= 3; i++ {
			err := mock.UpdateSalesforceAccount(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, i, mock.UpdateSalesforceAccountCalls)
		}
	})
}

func TestMockSplunkService_UpdateDataInput(t *testing.T) {
	t.Run("Success_WithCustomFunc", func(t *testing.T) {
		input := &utils.DataInput{
			Name:   "Test_Input",
			Object: "Account",
		}

		mock := &MockSplunkService{
			UpdateDataInputFunc: func(ctx context.Context, di *utils.DataInput) error {
				assert.Equal(t, input.Name, di.Name)
				assert.Equal(t, input.Object, di.Object)
				return nil
			},
		}

		err := mock.UpdateDataInput(context.Background(), input)
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.UpdateDataInputCalls)
	})

	t.Run("Success_WithoutCustomFunc", func(t *testing.T) {
		mock := &MockSplunkService{}
		input := &utils.DataInput{
			Name:   "Test_Input",
			Object: "Contact",
		}

		err := mock.UpdateDataInput(context.Background(), input)
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.UpdateDataInputCalls)
	})

	t.Run("Error_UpdateFailed", func(t *testing.T) {
		expectedErr := errors.New("update failed")
		mock := &MockSplunkService{
			UpdateDataInputFunc: func(ctx context.Context, di *utils.DataInput) error {
				return expectedErr
			},
		}

		input := &utils.DataInput{}
		err := mock.UpdateDataInput(context.Background(), input)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, mock.UpdateDataInputCalls)
	})

	t.Run("Success_NilDataInput", func(t *testing.T) {
		mock := &MockSplunkService{
			UpdateDataInputFunc: func(ctx context.Context, di *utils.DataInput) error {
				assert.Nil(t, di)
				return nil
			},
		}

		err := mock.UpdateDataInput(context.Background(), nil)
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.UpdateDataInputCalls)
	})

	t.Run("Success_MultipleCalls_WithDifferentInputs", func(t *testing.T) {
		mock := &MockSplunkService{}
		inputs := []*utils.DataInput{
			{Name: "Input1", Object: "Account"},
			{Name: "Input2", Object: "Contact"},
		}

		for i, input := range inputs {
			err := mock.UpdateDataInput(context.Background(), input)
			assert.NoError(t, err)
			assert.Equal(t, i+1, mock.UpdateDataInputCalls)
		}
	})
}

func TestMockSplunkService_CheckDataInputExists(t *testing.T) {
	t.Run("Success_InputExists", func(t *testing.T) {
		mock := &MockSplunkService{
			CheckDataInputExistsFunc: func(ctx context.Context, inputName string) (bool, error) {
				assert.Equal(t, "Test_Input", inputName)
				return true, nil
			},
		}

		exists, err := mock.CheckDataInputExists(context.Background(), "Test_Input")
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, 1, mock.CheckDataInputExistsCalls)
	})

	t.Run("Success_InputDoesNotExist", func(t *testing.T) {
		mock := &MockSplunkService{
			CheckDataInputExistsFunc: func(ctx context.Context, inputName string) (bool, error) {
				return false, nil
			},
		}

		exists, err := mock.CheckDataInputExists(context.Background(), "Nonexistent_Input")
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.Equal(t, 1, mock.CheckDataInputExistsCalls)
	})

	t.Run("Success_WithoutCustomFunc_ReturnsFalse", func(t *testing.T) {
		mock := &MockSplunkService{}

		exists, err := mock.CheckDataInputExists(context.Background(), "Test_Input")
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.Equal(t, 1, mock.CheckDataInputExistsCalls)
	})

	t.Run("Error_CheckFailed", func(t *testing.T) {
		expectedErr := errors.New("check failed")
		mock := &MockSplunkService{
			CheckDataInputExistsFunc: func(ctx context.Context, inputName string) (bool, error) {
				return false, expectedErr
			},
		}

		exists, err := mock.CheckDataInputExists(context.Background(), "Test_Input")
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.False(t, exists)
		assert.Equal(t, 1, mock.CheckDataInputExistsCalls)
	})

	t.Run("Success_MultipleCalls_WithDifferentInputNames", func(t *testing.T) {
		mock := &MockSplunkService{
			CheckDataInputExistsFunc: func(ctx context.Context, inputName string) (bool, error) {
				return inputName == "Existing_Input", nil
			},
		}

		inputNames := []string{"Existing_Input", "Nonexistent", "Another_Input"}
		for i, name := range inputNames {
			exists, err := mock.CheckDataInputExists(context.Background(), name)
			assert.NoError(t, err)
			assert.Equal(t, name == "Existing_Input", exists)
			assert.Equal(t, i+1, mock.CheckDataInputExistsCalls)
		}
	})
}
