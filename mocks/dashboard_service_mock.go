package mocks

import (
	"context"
	"sync"
)

// MockDashboardService is a mock implementation of DashboardServiceInterface
type MockDashboardService struct {
	CreateDashboardsFromDirectoryFunc  func(ctx context.Context, dashboardDir string) error
	CreateDashboardsFromDirectoryCalls int
	mu                                 sync.RWMutex
}

// CreateDashboardsFromDirectory mocks dashboard creation from directory
func (m *MockDashboardService) CreateDashboardsFromDirectory(ctx context.Context, dashboardDir string) error {
	m.mu.Lock()
	m.CreateDashboardsFromDirectoryCalls++
	m.mu.Unlock()

	if m.CreateDashboardsFromDirectoryFunc != nil {
		return m.CreateDashboardsFromDirectoryFunc(ctx, dashboardDir)
	}
	return nil
}

// Reset resets all call counters
func (m *MockDashboardService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CreateDashboardsFromDirectoryCalls = 0
}
