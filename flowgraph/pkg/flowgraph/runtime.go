package flowgraph

import (
	"context"
	"time"

	graphrepo "github.com/flowgraph/flowgraph/internal/adapters/repository/graph"
	memory "github.com/flowgraph/flowgraph/internal/adapters/repository/memory"
	"github.com/flowgraph/flowgraph/internal/app/dto"
	"github.com/flowgraph/flowgraph/internal/app/services"
	"github.com/flowgraph/flowgraph/internal/app/usecases"
	coregraph "github.com/flowgraph/flowgraph/internal/core/graph"
)

// Re-export core graph types for convenience
type Graph = coregraph.Graph
type Node = coregraph.Node
type Edge = coregraph.Edge
type NodeType = coregraph.NodeType

// Re-export node type constants
const (
	NodeTypeFunction    = coregraph.NodeTypeFunction
	NodeTypeTool        = coregraph.NodeTypeTool
	NodeTypeAgent       = coregraph.NodeTypeAgent
	NodeTypeConditional = coregraph.NodeTypeConditional
	NodeTypeSubgraph    = coregraph.NodeTypeSubgraph
)

// Re-export DTO types for public API
type ExecutionRequest = dto.ExecutionRequest
type ExecutionResponse = dto.ExecutionResponse
type ExecutionConfig = dto.ExecutionConfig

// Re-export interface types for extensibility
type NodeProcessor = usecases.NodeProcessor

// Runtime is a simple façade to construct and run graphs without importing
// internal packages directly. The default runtime uses in-memory components and
// is suitable for local usage and tests.
type Runtime struct {
	executor usecases.GraphExecutor
	repo     usecases.GraphRepository
}

// NewRuntime constructs a default runtime with in-memory services suitable for local usage.
func NewRuntime() *Runtime {
	nodeProcessor := usecases.NewDefaultNodeProcessor()
	return NewRuntimeWithNodeProcessor(nodeProcessor)
}

// NewRuntimeWithNodeProcessor constructs a runtime using the supplied node processor.
// This allows callers to register custom node types without reimplementing the
// rest of the runtime wiring.
func NewRuntimeWithNodeProcessor(nodeProcessor usecases.NodeProcessor) *Runtime {
	edgeEvaluator := usecases.NewDefaultEdgeEvaluator()
	stateManager := services.NewStateService()
	saver := memory.DefaultInMemorySaver()
	checkpointManager := services.NewCheckpointService(saver)
	repo := graphrepo.NewInMemoryGraphRepository()
	executor := usecases.NewDefaultGraphExecutor(nodeProcessor, edgeEvaluator, stateManager, checkpointManager, repo)

	return &Runtime{executor: executor, repo: repo}
}

// SaveGraph persists a graph to the runtime repository.
func (rt *Runtime) SaveGraph(ctx context.Context, g *coregraph.Graph) error {
	return rt.repo.Save(ctx, g)
}

// Execute runs a graph with the provided request.
func (rt *Runtime) Execute(ctx context.Context, req *dto.ExecutionRequest) (*dto.ExecutionResponse, error) {
	return rt.executor.Execute(ctx, req)
}

// RunSimple saves the graph (if not already) and executes it with a minimal
// request configuration.
func (rt *Runtime) RunSimple(ctx context.Context, g *coregraph.Graph, threadID string, input map[string]interface{}) (*dto.ExecutionResponse, error) {
	// Ensure graph is saved
	if err := rt.repo.Save(ctx, g); err != nil {
		return nil, err
	}
	req := &dto.ExecutionRequest{
		GraphID:  g.ID,
		ThreadID: threadID,
		Input:    input,
		Config: dto.ExecutionConfig{
			MaxSteps:        100,
			Timeout:         time.Minute,
			CheckpointEvery: 10,
			ValidateGraph:   true,
		},
	}
	return rt.executor.Execute(ctx, req)
}
