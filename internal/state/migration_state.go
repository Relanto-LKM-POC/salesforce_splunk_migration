// Package state provides state management for migration workflows
package state

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"salesforce-splunk-migration/utils"
)

// MigrationStateManager manages the state of migration using FlowGraph patterns
type MigrationStateManager struct {
	mu            sync.RWMutex
	executionID   string
	status        ExecutionStatus
	currentStep   string
	steps         map[string]*StepState
	metadata      map[string]interface{}
	startTime     time.Time
	endTime       *time.Time
	errorMessages []string
}

// ExecutionStatus represents the current status of migration execution
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusPaused    ExecutionStatus = "paused"
)

// StepState represents the state of a single migration step
type StepState struct {
	Name       string                 `json:"name"`
	Status     ExecutionStatus        `json:"status"`
	StartTime  *time.Time             `json:"start_time,omitempty"`
	EndTime    *time.Time             `json:"end_time,omitempty"`
	Duration   time.Duration          `json:"duration,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	RetryCount int                    `json:"retry_count"`
}

// NewMigrationStateManager creates a new state manager
func NewMigrationStateManager(executionID string) *MigrationStateManager {
	return &MigrationStateManager{
		executionID:   executionID,
		status:        StatusPending,
		steps:         make(map[string]*StepState),
		metadata:      make(map[string]interface{}),
		startTime:     time.Now(),
		errorMessages: make([]string, 0),
	}
}

// StartExecution marks the execution as started
func (m *MigrationStateManager) StartExecution() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status = StatusRunning
	m.startTime = time.Now()
}

// CompleteExecution marks the execution as completed
func (m *MigrationStateManager) CompleteExecution() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	m.status = StatusCompleted
	m.endTime = &now
}

// FailExecution marks the execution as failed
func (m *MigrationStateManager) FailExecution(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	m.status = StatusFailed
	m.endTime = &now
	if err != nil {
		m.errorMessages = append(m.errorMessages, err.Error())
	}
}

// StartStep marks a step as started
func (m *MigrationStateManager) StartStep(stepName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	m.currentStep = stepName

	if _, exists := m.steps[stepName]; !exists {
		m.steps[stepName] = &StepState{
			Name:     stepName,
			Metadata: make(map[string]interface{}),
		}
	}

	m.steps[stepName].Status = StatusRunning
	m.steps[stepName].StartTime = &now
}

// CompleteStep marks a step as completed
func (m *MigrationStateManager) CompleteStep(stepName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if step, exists := m.steps[stepName]; exists {
		now := time.Now()
		step.Status = StatusCompleted
		step.EndTime = &now
		if step.StartTime != nil {
			step.Duration = now.Sub(*step.StartTime)
		}
	}
}

// FailStep marks a step as failed
func (m *MigrationStateManager) FailStep(stepName string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if step, exists := m.steps[stepName]; exists {
		now := time.Now()
		step.Status = StatusFailed
		step.EndTime = &now
		if err != nil {
			step.Error = err.Error()
		}
		if step.StartTime != nil {
			step.Duration = now.Sub(*step.StartTime)
		}
	}
}

// SetStepMetadata sets metadata for a specific step
func (m *MigrationStateManager) SetStepMetadata(stepName string, key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if step, exists := m.steps[stepName]; exists {
		if step.Metadata == nil {
			step.Metadata = make(map[string]interface{})
		}
		step.Metadata[key] = value
	}
}

// GetStepMetadata retrieves metadata for a specific step
func (m *MigrationStateManager) GetStepMetadata(stepName string, key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if step, exists := m.steps[stepName]; exists {
		val, ok := step.Metadata[key]
		return val, ok
	}
	return nil, false
}

// SetMetadata sets global metadata
func (m *MigrationStateManager) SetMetadata(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata[key] = value
}

// GetMetadata retrieves global metadata
func (m *MigrationStateManager) GetMetadata(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.metadata[key]
	return val, ok
}

// GetCurrentStatus returns the current execution status
func (m *MigrationStateManager) GetCurrentStatus() ExecutionStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

// GetCurrentStep returns the current step being executed
func (m *MigrationStateManager) GetCurrentStep() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentStep
}

// GetStepState returns the state of a specific step
func (m *MigrationStateManager) GetStepState(stepName string) (*StepState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	step, exists := m.steps[stepName]
	return step, exists
}

// GetAllSteps returns all step states
func (m *MigrationStateManager) GetAllSteps() map[string]*StepState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid race conditions
	steps := make(map[string]*StepState, len(m.steps))
	for k, v := range m.steps {
		stepCopy := *v
		steps[k] = &stepCopy
	}
	return steps
}

// GetExecutionDuration returns the total execution duration
func (m *MigrationStateManager) GetExecutionDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.endTime != nil {
		return m.endTime.Sub(m.startTime)
	}
	return time.Since(m.startTime)
}

// ToJSON serializes the state to JSON
func (m *MigrationStateManager) ToJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stateSnapshot := map[string]interface{}{
		"execution_id": m.executionID,
		"status":       m.status,
		"current_step": m.currentStep,
		"steps":        m.steps,
		"metadata":     m.metadata,
		"start_time":   m.startTime,
		"end_time":     m.endTime,
		"duration":     m.GetExecutionDuration(),
		"errors":       m.errorMessages,
	}

	return json.MarshalIndent(stateSnapshot, "", "  ")
}

// PrintSummary prints a human-readable summary of the execution state
func (m *MigrationStateManager) PrintSummary() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 Migration Execution Summary")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Execution ID: %s\n", m.executionID)
	fmt.Printf("Status: %s\n", m.status)
	fmt.Printf("Duration: %v\n", m.GetExecutionDuration().Round(time.Millisecond))
	fmt.Printf("Start Time: %s\n", m.startTime.Format(time.RFC3339))
	if m.endTime != nil {
		fmt.Printf("End Time: %s\n", m.endTime.Format(time.RFC3339))
	}

	fmt.Println("\n📝 Step Details:")
	fmt.Println(strings.Repeat("-", 60))

	stepOrder := []string{"authenticate", "create_index", "create_account", "load_data_inputs", "create_data_inputs", "verify_inputs"}
	for _, stepName := range stepOrder {
		if step, exists := m.steps[stepName]; exists {
			statusIcon := "⏳"
			switch step.Status {
			case StatusCompleted:
				statusIcon = "✅"
			case StatusFailed:
				statusIcon = "❌"
			case StatusRunning:
				statusIcon = "🔄"
			}

			fmt.Printf("%s %s (%s)", statusIcon, step.Name, step.Status)
			if step.Duration > 0 {
				fmt.Printf(" - %v", step.Duration.Round(time.Millisecond))
			}
			fmt.Println()

			if step.Error != "" {
				fmt.Printf("   Error: %s\n", step.Error)
			}

			// Print relevant metadata
			if len(step.Metadata) > 0 {
				for k, v := range step.Metadata {
					fmt.Printf("   %s: %v\n", k, v)
				}
			}
		}
	}

	if len(m.errorMessages) > 0 {
		fmt.Println("\n❌ Errors:")
		fmt.Println(strings.Repeat("-", 60))
		for i, err := range m.errorMessages {
			fmt.Printf("%d. %s\n", i+1, err)
		}
	}

	fmt.Println(strings.Repeat("=", 60))
}

// StateSnapshot represents a point-in-time snapshot of migration state
type StateSnapshot struct {
	ExecutionID   string                 `json:"execution_id"`
	Status        ExecutionStatus        `json:"status"`
	CurrentStep   string                 `json:"current_step"`
	Steps         map[string]*StepState  `json:"steps"`
	Metadata      map[string]interface{} `json:"metadata"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       *time.Time             `json:"end_time,omitempty"`
	Duration      time.Duration          `json:"duration"`
	ErrorMessages []string               `json:"errors,omitempty"`
}

// CreateSnapshot creates a point-in-time snapshot of the current state
func (m *MigrationStateManager) CreateSnapshot() *StateSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Deep copy steps
	stepsCopy := make(map[string]*StepState, len(m.steps))
	for k, v := range m.steps {
		stepCopy := *v
		stepsCopy[k] = &stepCopy
	}

	// Deep copy metadata
	metadataCopy := make(map[string]interface{}, len(m.metadata))
	for k, v := range m.metadata {
		metadataCopy[k] = v
	}

	return &StateSnapshot{
		ExecutionID:   m.executionID,
		Status:        m.status,
		CurrentStep:   m.currentStep,
		Steps:         stepsCopy,
		Metadata:      metadataCopy,
		StartTime:     m.startTime,
		EndTime:       m.endTime,
		Duration:      m.GetExecutionDuration(),
		ErrorMessages: append([]string{}, m.errorMessages...),
	}
}

// DataInputProgress tracks progress of data input creation
type DataInputProgress struct {
	Total      int                        `json:"total"`
	Completed  int                        `json:"completed"`
	Failed     int                        `json:"failed"`
	InProgress int                        `json:"in_progress"`
	Items      map[string]*DataInputState `json:"items"`
}

// DataInputState represents the state of a single data input
type DataInputState struct {
	Name      string          `json:"name"`
	Object    string          `json:"object"`
	Status    ExecutionStatus `json:"status"`
	StartTime *time.Time      `json:"start_time,omitempty"`
	EndTime   *time.Time      `json:"end_time,omitempty"`
	Duration  time.Duration   `json:"duration,omitempty"`
	Error     string          `json:"error,omitempty"`
}

// NewDataInputProgress creates a new data input progress tracker
func NewDataInputProgress(inputs []utils.DataInput) *DataInputProgress {
	items := make(map[string]*DataInputState, len(inputs))
	for _, input := range inputs {
		items[input.Name] = &DataInputState{
			Name:   input.Name,
			Object: input.Object,
			Status: StatusPending,
		}
	}

	return &DataInputProgress{
		Total: len(inputs),
		Items: items,
	}
} // StartInput marks a data input as started
func (p *DataInputProgress) StartInput(name string) {
	if item, exists := p.Items[name]; exists {
		now := time.Now()
		item.Status = StatusRunning
		item.StartTime = &now
		p.InProgress++
	}
}

// CompleteInput marks a data input as completed
func (p *DataInputProgress) CompleteInput(name string) {
	if item, exists := p.Items[name]; exists {
		now := time.Now()
		item.Status = StatusCompleted
		item.EndTime = &now
		if item.StartTime != nil {
			item.Duration = now.Sub(*item.StartTime)
		}
		p.Completed++
		p.InProgress--
	}
}

// FailInput marks a data input as failed
func (p *DataInputProgress) FailInput(name string, err error) {
	if item, exists := p.Items[name]; exists {
		now := time.Now()
		item.Status = StatusFailed
		item.EndTime = &now
		if err != nil {
			item.Error = err.Error()
		}
		if item.StartTime != nil {
			item.Duration = now.Sub(*item.StartTime)
		}
		p.Failed++
		p.InProgress--
	}
}

// GetProgress returns current progress statistics
func (p *DataInputProgress) GetProgress() (total, completed, failed, inProgress int) {
	return p.Total, p.Completed, p.Failed, p.InProgress
}

// Helper function to generate execution ID
func GenerateExecutionID() string {
	return fmt.Sprintf("migration-%d", time.Now().Unix())
}
