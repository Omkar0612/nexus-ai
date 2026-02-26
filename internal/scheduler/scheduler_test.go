package scheduler

import (
	"context"
	"testing"
	"time"
)

func TestSchedulerRegisterAndList(t *testing.T) {
	s := New(time.Second)
	err := s.Register(&Job{
		ID:      "test-job",
		Name:    "Test Job",
		Trigger: TriggerInterval,
		Interval: time.Minute,
		Handler: func(ctx context.Context) error { return nil },
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	list := s.ListJobs()
	if !containsStr(list, "Test Job") {
		t.Error("expected job in list output")
	}
}

func TestSchedulerRegisterNoHandler(t *testing.T) {
	s := New(time.Second)
	err := s.Register(&Job{ID: "bad", Trigger: TriggerInterval})
	if err == nil {
		t.Error("expected error for job without handler")
	}
}

func TestSchedulerEnableDisable(t *testing.T) {
	s := New(time.Second)
	_ = s.Register(&Job{
		ID: "toggle-job", Name: "Toggle",
		Trigger: TriggerInterval, Interval: time.Minute,
		Handler: func(ctx context.Context) error { return nil },
	})
	if err := s.Disable("toggle-job"); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	if err := s.Enable("toggle-job"); err != nil {
		t.Fatalf("Enable: %v", err)
	}
}

func TestSchedulerConditionSkip(t *testing.T) {
	s := New(100 * time.Millisecond)
	ran := false
	_ = s.Register(&Job{
		ID: "cond-job", Name: "Conditional",
		Trigger: TriggerInterval,
		Interval: 50 * time.Millisecond,
		Conditions: []Condition{
			func(ctx context.Context) (bool, string) {
				return false, "condition not met"
			},
		},
		Handler: func(ctx context.Context) error {
			ran = true
			return nil
		},
	})
	s.jobs["cond-job"].NextRun = time.Now().Add(-time.Second)
	s.Start()
	time.Sleep(300 * time.Millisecond)
	s.Stop()
	if ran {
		t.Error("handler should not have run when condition is false")
	}
}

func TestFileExistsCondition(t *testing.T) {
	cond := FileExistsCondition("/tmp/nexus_test_file_does_not_exist_xyz")
	ok, reason := cond(context.Background())
	if ok {
		t.Error("expected false for non-existent file")
	}
	if reason == "" {
		t.Error("expected non-empty reason")
	}
}

func TestParseCronNext(t *testing.T) {
	next := parseCronNext("0 9 * * *", "UTC")
	if next.IsZero() {
		t.Error("expected non-zero next time")
	}
	if next.Before(time.Now()) {
		t.Error("next run should be in the future")
	}
}

func containsStr(s, sub string) bool {
	return len(s) > 0 && len(sub) > 0 && (s == sub || len(s) >= len(sub) && (s[:len(sub)] == sub || containsStr(s[1:], sub)))
}
