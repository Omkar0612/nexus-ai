package circuit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Breaker implements a circuit breaker pattern for agent tool calls.
type Breaker struct {
	mu            sync.RWMutex
	toolStates    map[string]*ToolState
	failThreshold int // Number of consecutive failures to trip circuit
}

// ToolState tracks the health of a specific agent tool integration.
type ToolState struct {
	ToolName          string
	State             CircuitState
	ConsecutiveFails  int
	LastFailTime      time.Time
	TotalCalls        int
	FailedCalls       int
}

type CircuitState string

const (
	StateClosed   CircuitState = "CLOSED"   // Normal operation
	StateOpen     CircuitState = "OPEN"     // Circuit tripped, calls blocked
	StateHalfOpen CircuitState = "HALF_OPEN" // Testing if service recovered
)

// NewBreaker initializes a circuit breaker for agent tools.
func NewBreaker(failThreshold int) *Breaker {
	return &Breaker{
		toolStates:    make(map[string]*ToolState),
		failThreshold: failThreshold,
	}
}

// Call wraps a tool execution with circuit breaker logic.
func (b *Breaker) Call(ctx context.Context, toolName string, fn func() error) error {
	b.mu.Lock()
	state, exists := b.toolStates[toolName]
	if !exists {
		state = &ToolState{
			ToolName: toolName,
			State:    StateClosed,
		}
		b.toolStates[toolName] = state
	}
	b.mu.Unlock()

	// If circuit is open, fail fast
	if state.State == StateOpen {
		// Check if we should attempt recovery
		if time.Since(state.LastFailTime) > 30*time.Second {
			state.State = StateHalfOpen
			log.Info().Str("tool", toolName).Msg("🔄 Circuit entering half-open state (testing recovery)")
		} else {
			return fmt.Errorf("circuit breaker OPEN for tool '%s' (degraded to read-only)", toolName)
		}
	}

	// Execute the tool call
	err := fn()

	b.mu.Lock()
	defer b.mu.Unlock()

	state.TotalCalls++

	if err != nil {
		state.FailedCalls++
		state.ConsecutiveFails++
		state.LastFailTime = time.Now()

		// Trip the circuit if threshold breached
		if state.ConsecutiveFails >= b.failThreshold {
			state.State = StateOpen
			log.Warn().
				Str("tool", toolName).
				Int("consecutive_fails", state.ConsecutiveFails).
				Msg("🛑 Circuit breaker OPEN (tool degraded)")
		}

		return err
	}

	// Success — reset failure counter and close circuit
	state.ConsecutiveFails = 0
	if state.State == StateHalfOpen {
		state.State = StateClosed
		log.Info().Str("tool", toolName).Msg("✅ Circuit breaker CLOSED (tool recovered)")
	}

	return nil
}

// GetStats returns the current health metrics for a tool.
func (b *Breaker) GetStats(toolName string) *ToolState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.toolStates[toolName]
}
