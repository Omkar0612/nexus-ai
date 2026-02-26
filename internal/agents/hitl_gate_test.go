package agents

import (
	"context"
	"testing"
	"time"
)

func TestHITLLowRiskAutoExecute(t *testing.T) {
	gate := NewHITLGate(5*time.Second, nil)
	executed := false
	err := gate.Execute(context.Background(), "read file", "user requested", "low", func(ctx context.Context) error {
		executed = true
		return nil
	})
	if err != nil {
		t.Fatalf("Execute low risk: %v", err)
	}
	if !executed {
		t.Error("expected low-risk action to auto-execute")
	}
}

func TestHITLMediumRiskAutoExecute(t *testing.T) {
	gate := NewHITLGate(5*time.Second, nil)
	executed := false
	err := gate.Execute(context.Background(), "update config", "user requested", "medium", func(ctx context.Context) error {
		executed = true
		return nil
	})
	if err != nil {
		t.Fatalf("Execute medium risk: %v", err)
	}
	if !executed {
		t.Error("expected medium-risk action to auto-execute")
	}
}

func TestHITLHighRiskTimeout(t *testing.T) {
	gate := NewHITLGate(200*time.Millisecond, nil) // very short timeout
	err := gate.Execute(context.Background(), "delete database", "cleanup requested", "high",
		func(ctx context.Context) error { return nil },
	)
	if err == nil {
		t.Error("expected timeout error for high-risk action with no approver")
	}
}

func TestHITLHighRiskApprove(t *testing.T) {
	gate := NewHITLGate(5*time.Second, nil)
	done := make(chan error, 1)

	go func() {
		done <- gate.Execute(context.Background(), "deploy to production", "scheduled deploy", "high",
			func(ctx context.Context) error { return nil },
		)
	}()

	// Give the goroutine time to register the pending request
	time.Sleep(50 * time.Millisecond)

	gate.mu.RLock()
	var reqID string
	for id := range gate.pending {
		reqID = id
	}
	gate.mu.RUnlock()

	if reqID == "" {
		t.Fatal("no pending request found")
	}
	if err := gate.Approve(reqID, "omkar"); err != nil {
		t.Fatalf("Approve: %v", err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil after approval, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for approved action to complete")
	}
}

func TestHITLEmergencyLock(t *testing.T) {
	gate := NewHITLGate(5*time.Second, nil)
	gate.EmergencyLock()
	err := gate.Execute(context.Background(), "send notification", "alert", "medium",
		func(ctx context.Context) error { return nil },
	)
	if err == nil {
		t.Error("expected error when gate is emergency locked")
	}
	gate.EmergencyUnlock()
	err = gate.Execute(context.Background(), "send notification", "alert", "medium",
		func(ctx context.Context) error { return nil },
	)
	if err != nil {
		t.Errorf("expected success after unlock, got: %v", err)
	}
}
