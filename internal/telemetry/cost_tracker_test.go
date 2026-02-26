package telemetry

import (
	"testing"
)

func TestCalculateCostKnownModel(t *testing.T) {
	ct := &CostTracker{}
	cost := ct.calculateCost("groq", "llama-3.3-70b-versatile", 1000, 500)
	if cost <= 0 {
		t.Errorf("expected positive cost, got %f", cost)
	}
}

func TestCalculateCostFreeModel(t *testing.T) {
	ct := &CostTracker{}
	cost := ct.calculateCost("ollama", "llama3.2", 100000, 100000)
	if cost != 0 {
		t.Errorf("expected 0 cost for free model, got %f", cost)
	}
}

func TestCalculateCostUnknownModel(t *testing.T) {
	ct := &CostTracker{}
	cost := ct.calculateCost("unknown", "model-x", 1_000_000, 1_000_000)
	if cost <= 0 {
		t.Errorf("expected fallback cost estimate, got %f", cost)
	}
}

func TestCostTrackerRecordAndStatus(t *testing.T) {
	ct, err := New(t.TempDir(), 1.00, 10.00)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer ct.Close()

	cost, err := ct.Record("user1", "groq", "llama-3.3-70b-versatile", "chat", "sess-1", 50000, 10000)
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	if cost < 0 {
		t.Errorf("expected non-negative cost, got %f", cost)
	}

	status, err := ct.GetStatus("user1")
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if status.DailySpent < 0 {
		t.Errorf("unexpected negative daily spent")
	}
}

func TestBudgetAlert(t *testing.T) {
	ct, err := New(t.TempDir(), 0.00001, 1.00)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer ct.Close()

	alerted := false
	ct.SetAlertCallback(func(msg string) { alerted = true })

	// Large call that should exceed the tiny daily limit
	_, err = ct.Record("user1", "groq", "llama-3.3-70b-versatile", "test", "s1", 1_000_000, 1_000_000)
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	if !alerted {
		t.Error("expected budget alert to fire")
	}
}

func TestSuggestCheaperModel(t *testing.T) {
	suggestion := SuggestCheaperModel("anthropic", "claude-3-opus")
	if suggestion == "" {
		t.Error("expected a cheaper model suggestion for expensive model")
	}
	none := SuggestCheaperModel("ollama", "llama3.2")
	if none != "" {
		t.Error("expected no suggestion for free model")
	}
}
