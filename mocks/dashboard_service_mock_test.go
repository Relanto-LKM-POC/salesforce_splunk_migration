package mocks

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockDashboardService_CreateDashboardsFromDirectory(t *testing.T) {
	t.Run("Success_WithCustomFunc", func(t *testing.T) {
		mock := &MockDashboardService{
			CreateDashboardsFromDirectoryFunc: func(ctx context.Context, dashboardDir string) error {
				assert.Equal(t, "/path/to/dashboards", dashboardDir)
				return nil
			},
		}

		err := mock.CreateDashboardsFromDirectory(context.Background(), "/path/to/dashboards")
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CreateDashboardsFromDirectoryCalls)
	})

	t.Run("Success_WithoutCustomFunc", func(t *testing.T) {
		mock := &MockDashboardService{}

		err := mock.CreateDashboardsFromDirectory(context.Background(), "/path/to/dashboards")
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CreateDashboardsFromDirectoryCalls)
	})

	t.Run("Error_DirectoryNotFound", func(t *testing.T) {
		expectedErr := errors.New("directory not found")
		mock := &MockDashboardService{
			CreateDashboardsFromDirectoryFunc: func(ctx context.Context, dashboardDir string) error {
				return expectedErr
			},
		}

		err := mock.CreateDashboardsFromDirectory(context.Background(), "/nonexistent")
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, mock.CreateDashboardsFromDirectoryCalls)
	})

	t.Run("Success_EmptyDirectory", func(t *testing.T) {
		mock := &MockDashboardService{
			CreateDashboardsFromDirectoryFunc: func(ctx context.Context, dashboardDir string) error {
				assert.Empty(t, dashboardDir)
				return nil
			},
		}

		err := mock.CreateDashboardsFromDirectory(context.Background(), "")
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CreateDashboardsFromDirectoryCalls)
	})

	t.Run("Success_NilContext", func(t *testing.T) {
		mock := &MockDashboardService{}

		err := mock.CreateDashboardsFromDirectory(nil, "/path/to/dashboards")
		assert.NoError(t, err)
		assert.Equal(t, 1, mock.CreateDashboardsFromDirectoryCalls)
	})

	t.Run("Success_MultipleCalls_IncrementCounter", func(t *testing.T) {
		mock := &MockDashboardService{}

		directories := []string{"/path1", "/path2", "/path3"}
		for i, dir := range directories {
			err := mock.CreateDashboardsFromDirectory(context.Background(), dir)
			assert.NoError(t, err)
			assert.Equal(t, i+1, mock.CreateDashboardsFromDirectoryCalls)
		}
	})

	t.Run("Error_CustomFuncReturnsError", func(t *testing.T) {
		expectedErr := errors.New("failed to create dashboards")
		mock := &MockDashboardService{
			CreateDashboardsFromDirectoryFunc: func(ctx context.Context, dashboardDir string) error {
				return expectedErr
			},
		}

		err := mock.CreateDashboardsFromDirectory(context.Background(), "/path")
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, mock.CreateDashboardsFromDirectoryCalls)
	})

	t.Run("Success_ConcurrentCalls_ThreadSafe", func(t *testing.T) {
		mock := &MockDashboardService{}

		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				_ = mock.CreateDashboardsFromDirectory(context.Background(), "/path")
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}

		assert.Equal(t, 10, mock.CreateDashboardsFromDirectoryCalls)
	})
}

func TestMockDashboardService_Reset(t *testing.T) {
	t.Run("Success_ResetsCallCounter", func(t *testing.T) {
		mock := &MockDashboardService{}

		// Call multiple times
		for i := 0; i < 5; i++ {
			_ = mock.CreateDashboardsFromDirectory(context.Background(), "/path")
		}

		assert.Equal(t, 5, mock.CreateDashboardsFromDirectoryCalls)

		// Reset
		mock.Reset()

		// Verify counter is reset
		assert.Equal(t, 0, mock.CreateDashboardsFromDirectoryCalls)
	})

	t.Run("Success_MultipleResets", func(t *testing.T) {
		mock := &MockDashboardService{}

		// First call and reset
		_ = mock.CreateDashboardsFromDirectory(context.Background(), "/path")
		assert.Equal(t, 1, mock.CreateDashboardsFromDirectoryCalls)
		mock.Reset()
		assert.Equal(t, 0, mock.CreateDashboardsFromDirectoryCalls)

		// Second call and reset
		_ = mock.CreateDashboardsFromDirectory(context.Background(), "/path")
		assert.Equal(t, 1, mock.CreateDashboardsFromDirectoryCalls)
		mock.Reset()
		assert.Equal(t, 0, mock.CreateDashboardsFromDirectoryCalls)
	})

	t.Run("Success_ResetWithoutPriorCalls", func(t *testing.T) {
		mock := &MockDashboardService{}

		// Reset without any prior calls
		mock.Reset()

		// Verify counter is zero
		assert.Equal(t, 0, mock.CreateDashboardsFromDirectoryCalls)
	})

	t.Run("Success_ConcurrentReset_ThreadSafe", func(t *testing.T) {
		mock := &MockDashboardService{}

		// Call multiple times
		for i := 0; i < 5; i++ {
			_ = mock.CreateDashboardsFromDirectory(context.Background(), "/path")
		}

		done := make(chan bool)
		// Concurrent resets
		for i := 0; i < 3; i++ {
			go func() {
				mock.Reset()
				done <- true
			}()
		}

		for i := 0; i < 3; i++ {
			<-done
		}

		assert.Equal(t, 0, mock.CreateDashboardsFromDirectoryCalls)
	})
}

func TestMockDashboardService_Integration(t *testing.T) {
	t.Run("Success_CompleteWorkflow", func(t *testing.T) {
		callOrder := []string{}
		mock := &MockDashboardService{
			CreateDashboardsFromDirectoryFunc: func(ctx context.Context, dashboardDir string) error {
				callOrder = append(callOrder, dashboardDir)
				return nil
			},
		}

		ctx := context.Background()

		// Execute workflow
		_ = mock.CreateDashboardsFromDirectory(ctx, "/resources/dashboards")
		_ = mock.CreateDashboardsFromDirectory(ctx, "/additional/dashboards")

		assert.Equal(t, 2, mock.CreateDashboardsFromDirectoryCalls)
		assert.Equal(t, []string{"/resources/dashboards", "/additional/dashboards"}, callOrder)
	})

	t.Run("Error_WorkflowFailsOnFirstCall", func(t *testing.T) {
		expectedErr := errors.New("first call failed")
		mock := &MockDashboardService{
			CreateDashboardsFromDirectoryFunc: func(ctx context.Context, dashboardDir string) error {
				return expectedErr
			},
		}

		err := mock.CreateDashboardsFromDirectory(context.Background(), "/path")
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, mock.CreateDashboardsFromDirectoryCalls)
	})
}
