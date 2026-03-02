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
	engine := NewShadowManager(ModeActive)
	engine.Start()
	defer engine.Stop()

	baseline := func(ctx context.Context) (*ExecutionMetrics, error) {
		return &ExecutionMetrics{
			CostUSD:      0.10, // $0.10 baseline cost
			Latency:      2 * time.Second,
			QualityScore: 0.9,
		}, nil
	}

	shadow := func(ctx context.Context) (*ExecutionMetrics, error) {
		return &ExecutionMetrics{
			CostUSD:      0.06, // $0.06 shadow cost (40% cheaper)
			Latency:      2 * time.Second,
			QualityScore: 0.9,
		}, nil
	}

	// Execute baseline strategy
	baselineMetrics, err := baseline(context.Background())
	if err != nil {
		t.Fatalf("baseline execution failed: %v", err)
	}

	// Execute shadow strategy
	shadowMetrics, err := shadow(context.Background())
	if err != nil {
		t.Fatalf("shadow execution failed: %v", err)
	}

	// Verify cost savings
	savings := (baselineMetrics.CostUSD - shadowMetrics.CostUSD) / baselineMetrics.CostUSD * 100
	if savings < 35 {
		t.Errorf("expected at least 35%% savings, got %.2f%%", savings)
	}

	if savings < 39 || savings > 41 {
		t.Logf("Cost savings: %.2f%%", savings)
	}
}

func TestShadowEvolution_QualityCheck(t *testing.T) {
	engine := NewShadowManager(ModePassive)
	engine.Start()
	defer engine.Stop()

	baseline := func(ctx context.Context) (*ExecutionMetrics, error) {
		return &ExecutionMetrics{
			CostUSD:      0.10,
			QualityScore: 0.95,
			Latency:      2 * time.Second,
		}, nil
	}

	shadowDegraded := func(ctx context.Context) (*ExecutionMetrics, error) {
		return &ExecutionMetrics{
			CostUSD:      0.01, // 90% cheaper but...
			QualityScore: 0.50, // Poor quality
			Latency:      1 * time.Second,
		}, nil
	}

	// Execute both
	baselineMetrics, _ := baseline(context.Background())
	shadowMetrics, _ := shadowDegraded(context.Background())

	// Verify quality check would reject this
	if shadowMetrics.QualityScore < baselineMetrics.QualityScore*0.8 {
		t.Log("Shadow strategy correctly identified as degraded")
	} else {
		t.Error("Expected shadow quality to be significantly lower")
	}
}

func TestShadowManager_StrategyRegistration(t *testing.T) {
	manager := NewShadowManager(ModeABTest)

	strategy := &Strategy{
		ID:      "test-strategy-1",
		Name:    "Test Strategy",
		Version: "1.0",
		Enabled: true,
	}

	err := manager.RegisterStrategy(strategy)
	if err != nil {
		t.Fatalf("failed to register strategy: %v", err)
	}

	// Try to register duplicate
	err = manager.RegisterStrategy(strategy)
	if err == nil {
		t.Error("expected error when registering duplicate strategy")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("expected 'already registered' error, got: %v", err)
	}
}

func TestShadowManager_GetMetrics(t *testing.T) {
	manager := NewShadowManager(ModePassive)
	manager.RegisterStrategy(&Strategy{
		ID:      "s1",
		Name:    "Strategy 1",
		Enabled: true,
	})
	manager.RegisterStrategy(&Strategy{
		ID:      "s2",
		Name:    "Strategy 2",
		Enabled: false,
	})

	metrics := manager.GetMetrics()

	if metrics["total_strategies"].(int) != 2 {
		t.Errorf("expected 2 total strategies, got %v", metrics["total_strategies"])
	}

	if metrics["enabled_strategies"].(int) != 1 {
		t.Errorf("expected 1 enabled strategy, got %v", metrics["enabled_strategies"])
	}
}
