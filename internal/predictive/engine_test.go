package predictive

import (
	"context"
	"sync"
	"testing"
	"time"
)

// MockLLM simulates instant Groq generation
type MockLLM struct{}

func (m *MockLLM) Generate(ctx context.Context, sys, prompt string) (string, error) {
	return "Mocked background deep research output.", nil
}

// MockCache stores results in memory
type MockCache struct {
	mu      sync.Mutex
	results map[string]*PrecomputedResult
}

func (m *MockCache) Store(r *PrecomputedResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.results[r.EventID] = r
	return nil
}
func (m *MockCache) GetUnviewed() ([]*PrecomputedResult, error) { return nil, nil }
func (m *MockCache) MarkViewed(id string) error                 { return nil }

// MockSource feeds a fake meeting event
type MockSource struct {
	polled bool
}

func (m *MockSource) Name() string { return "MockCalendar" }
func (m *MockSource) Poll() ([]Event, error) {
	if m.polled {
		return nil, nil // Only send event once
	}
	m.polled = true
	return []Event{
		{
			ID:        "evt_123",
			Type:      EventUpcomingMeeting,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"client_name": "John Doe",
				"company":     "Acme Corp",
			},
		},
	}, nil
}

func TestPredictiveEngine_E2E(t *testing.T) {
	cache := &MockCache{results: make(map[string]*PrecomputedResult)}
	llm := &MockLLM{}
	source := &MockSource{}

	// Create engine with a very fast 10ms tick for the test
	engine := NewEngine(llm, cache, 10*time.Millisecond)
	engine.RegisterSource(source)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine.Start(ctx)

	// Give the engine 50ms to poll, generate the LLM response, and cache it
	time.Sleep(50 * time.Millisecond)
	engine.Stop()

	// Verify the result was pre-computed and stored in cache
	cache.mu.Lock()
	defer cache.mu.Unlock()

	result, ok := cache.results["evt_123"]
	if !ok {
		t.Fatalf("expected pre-computed result to be in cache, but it was missing")
	}

	if result.Action != "Review Brief" {
		t.Errorf("expected action 'Review Brief', got: %s", result.Action)
	}

	expectedPrefix := "Meeting Brief: John Doe"
	if len(result.Summary) < len(expectedPrefix) || result.Summary[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("unexpected summary formatting: %s", result.Summary)
	}
}
