package shadow

import (
	"context"
	"strings"
	"testing"
	"time"
)

type mockGate struct {
	lastPrompt  string
	willApprove bool
	called      bool
}

func (m *mockGate) AskPermission(ctx context.Context, riskLevel, prompt string) bool {
	m.called = true
	m.lastPrompt = prompt
	return m.willApprove
}

func TestShadowEvolution_CostSavings(t *testing.T) {
	gate := &mockGate{willApprove: true}
	engine := New(gate)

	baseline := func(ctx context.Context) (*Metrics, error) {
		return &Metrics{
			Cost:    0.10, // $0.10 baseline cost
			Latency: 2 * time.Second,
			Output:  "Here is your morning digest with 5 news articles and weather.",
		}, nil
	}

	shadow := func(ctx context.Context) (*Metrics, error) {
		return &Metrics{
			Cost:    0.06, // $0.06 shadow cost (40% cheaper)
			Latency: 2 * time.Second,
			Output:  "Here is your morning digest with 5 news articles and weather.",
		}, nil
	}

	err := engine.Evaluate(context.Background(), "morning digest", baseline, shadow)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !gate.called {
		t.Fatal("expected HITL gate to be called for a 40% improvement")
	}

	expectedSubstr := "save 40% API costs"
	if !strings.Contains(gate.lastPrompt, expectedSubstr) {
		t.Errorf("expected prompt to contain %q, got: %s", expectedSubstr, gate.lastPrompt)
	}
}

func TestShadowEvolution_DiscardDegradedOutput(t *testing.T) {
	gate := &mockGate{willApprove: true}
	engine := New(gate)

	baseline := func(ctx context.Context) (*Metrics, error) {
		return &Metrics{
			Cost:    0.10,
			Output:  "A very long detailed report containing 10 sections of deep research.",
		}, nil
	}

	shadow := func(ctx context.Context) (*Metrics, error) {
		return &Metrics{
			Cost:    0.01, // 90% cheaper! But...
			Output:  "Too short.", // Degraded output quality (<80% length)
		}, nil
	}

	err := engine.Evaluate(context.Background(), "deep research", baseline, shadow)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if gate.called {
		t.Fatal("expected HITL gate NOT to be called because shadow output was degraded")
	}
}
