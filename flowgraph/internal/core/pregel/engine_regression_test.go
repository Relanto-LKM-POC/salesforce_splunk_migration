package pregel

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type singleSendVertex struct {
	target string
}

func (v *singleSendVertex) Compute(vertexID string, state map[string]interface{}, messages []*Message) (map[string]interface{}, []*Message, bool, error) {
	sent, _ := state["sent"].(bool)
	newState := map[string]interface{}{"sent": true}
	if !sent {
		msg := &Message{From: vertexID, To: v.target, Value: "ping"}
		return newState, []*Message{msg}, true, nil
	}
	return newState, nil, true, nil
}

type reactivatedVertex struct{}

func (v *reactivatedVertex) Compute(vertexID string, state map[string]interface{}, messages []*Message) (map[string]interface{}, []*Message, bool, error) {
	runs, _ := state["runs"].(int)
	received, _ := state["received"].(bool)

	runs++
	if len(messages) > 0 {
		received = true
	}

	return map[string]interface{}{
		"runs":     runs,
		"received": received,
	}, nil, true, nil
}

func TestEngine_MessagePersistsBetweenSupersteps(t *testing.T) {
	vertices := map[string]VertexProgram{
		"A": &singleSendVertex{target: "B"},
		"B": &reactivatedVertex{},
	}
	initialStates := map[string]map[string]interface{}{
		"A": {"sent": false},
		"B": {"runs": 0, "received": false},
	}

	engine := NewEngine(vertices, initialStates, Config{
		MaxSupersteps: 4,
		Parallelism:   1,
	})

	require.NoError(t, engine.Run(context.Background()))

	state := engine.VertexStates["B"]
	require.Equal(t, 2, state["runs"])
	require.Equal(t, true, state["received"])
}

type flakyReceiver struct {
	runs     int
	failed   bool
	attempts []int
}

func (f *flakyReceiver) Compute(vertexID string, state map[string]interface{}, messages []*Message) (map[string]interface{}, []*Message, bool, error) {
	f.runs++

	if len(messages) == 0 {
		return map[string]interface{}{
			"succeeded":  false,
			"last_count": 0,
		}, nil, true, nil
	}

	f.attempts = append(f.attempts, len(messages))

	if !f.failed {
		f.failed = true
		return map[string]interface{}{
			"succeeded":  false,
			"last_count": len(messages),
		}, nil, true, fmt.Errorf("transient failure")
	}

	return map[string]interface{}{
		"succeeded":  true,
		"last_count": len(messages),
	}, nil, true, nil
}

func TestEngine_RetryKeepsInputMessages(t *testing.T) {
	receiver := &flakyReceiver{}
	vertices := map[string]VertexProgram{
		"A": &singleSendVertex{target: "B"},
		"B": receiver,
	}

	initialStates := map[string]map[string]interface{}{
		"A": {"sent": false},
		"B": {},
	}

	engine := NewEngine(vertices, initialStates, Config{
		MaxSupersteps: 4,
		Parallelism:   1,
		RetryPolicy: RetryPolicy{
			MaxRetries: 1,
		},
	})

	require.NoError(t, engine.Run(context.Background()))

	state := engine.VertexStates["B"]
	require.Equal(t, true, state["succeeded"])
	require.Equal(t, 1, state["last_count"])
	require.Equal(t, 3, receiver.runs)
	require.Equal(t, []int{1, 1}, receiver.attempts)
}
