package observe

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Trace represents a complete execution timeline for an agent task.
type Trace struct {
	ID          string
	AgentName   string
	StartTime   time.Time
	EndTime     time.Time
	Steps       []TraceStep
	TokensUsed  int
	CostUSD     float64
	Status      TraceStatus
	ErrorReason string
}

type TraceStatus string

const (
	StatusRunning  TraceStatus = "RUNNING"
	StatusSuccess  TraceStatus = "SUCCESS"
	StatusFailed   TraceStatus = "FAILED"
	StatusKilled   TraceStatus = "KILLED"
	StatusLooped   TraceStatus = "HALTED_LOOP_DETECTED"
)

// TraceStep represents a single atomic action in the agent's reasoning chain.
type TraceStep struct {
	StepID      int
	Timestamp   time.Time
	Action      string // e.g. "tool_call", "llm_reasoning", "memory_lookup"
	ToolName    string
	ToolArgs    map[string]interface{}
	ToolOutput  string
	LatencyMs   int64
	Success     bool
	RetryCount  int
}

// Tracer collects structured logs for agent execution timelines.
type Tracer struct {
	mu     sync.RWMutex
	traces map[string]*Trace
}

// NewTracer initializes the observability tracer.
func NewTracer() *Tracer {
	return &Tracer{
		traces: make(map[string]*Trace),
	}
}

// StartTrace begins a new execution timeline.
func (t *Tracer) StartTrace(ctx context.Context, agentName, traceID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.traces[traceID] = &Trace{
		ID:        traceID,
		AgentName: agentName,
		StartTime: time.Now(),
		Status:    StatusRunning,
		Steps:     make([]TraceStep, 0),
	}

	log.Info().Str("trace_id", traceID).Str("agent", agentName).Msg("🔍 Trace started")
}

// RecordStep logs a single agent action with full context.
func (t *Tracer) RecordStep(traceID string, step TraceStep) {
	t.mu.Lock()
	defer t.mu.Unlock()

	trace, exists := t.traces[traceID]
	if !exists {
		return
	}

	step.StepID = len(trace.Steps) + 1
	step.Timestamp = time.Now()
	trace.Steps = append(trace.Steps, step)

	// Detect hallucination loops (same tool call repeated 3+ times)
	if t.detectLoop(trace) {
		trace.Status = StatusLooped
		trace.ErrorReason = fmt.Sprintf("Loop detected: tool '%s' called %d times consecutively", step.ToolName, t.countRepeats(trace, step.ToolName))
		log.Warn().Str("trace_id", traceID).Str("tool", step.ToolName).Msg("⚠️ Hallucination loop detected")
	}
}

// EndTrace finalizes the execution and calculates total cost.
func (t *Tracer) EndTrace(traceID string, status TraceStatus, tokensUsed int, costUSD float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	trace, exists := t.traces[traceID]
	if !exists {
		return
	}

	trace.EndTime = time.Now()
	trace.Status = status
	trace.TokensUsed = tokensUsed
	trace.CostUSD = costUSD

	duration := trace.EndTime.Sub(trace.StartTime)
	log.Info().
		Str("trace_id", traceID).
		Str("status", string(status)).
		Int("steps", len(trace.Steps)).
		Int("tokens", tokensUsed).
		Float64("cost_usd", costUSD).
		Str("duration", duration.String()).
		Msg("✅ Trace completed")
}

// GetTrace retrieves the full execution timeline for debugging.
func (t *Tracer) GetTrace(traceID string) (*Trace, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	trace, exists := t.traces[traceID]
	return trace, exists
}

func (t *Tracer) detectLoop(trace *Trace) bool {
	if len(trace.Steps) < 3 {
		return false
	}

	// Check last 3 steps for identical tool calls
	steps := trace.Steps
	n := len(steps)
	last := steps[n-1]

	if n >= 3 && steps[n-2].ToolName == last.ToolName && steps[n-3].ToolName == last.ToolName {
		return true
	}
	return false
}

func (t *Tracer) countRepeats(trace *Trace, toolName string) int {
	count := 0
	for i := len(trace.Steps) - 1; i >= 0; i-- {
		if trace.Steps[i].ToolName == toolName {
			count++
		} else {
			break
		}
	}
	return count
}
