package agents

import (
	"testing"
	"time"

	"github.com/Omkar0612/nexus-ai/internal/memory"
)

// --- DriftDetector ---

func newMemStoreWithEntries(t *testing.T, entries []memory.Memory) *memory.Store {
	t.Helper()
	store, err := memory.New(t.TempDir())
	if err != nil {
		t.Fatalf("memory.New: %v", err)
	}
	for _, e := range entries {
		_ = store.Save(e)
	}
	return store
}

func TestDriftDetectorRepetitiveFailures(t *testing.T) {
	now := time.Now()
	entries := []memory.Memory{
		{UserID: "u1", Role: "user", Content: "the API is broken again", CreatedAt: now.Add(-3 * time.Hour)},
		{UserID: "u1", Role: "user", Content: "still broken after the fix", CreatedAt: now.Add(-2 * time.Hour)},
		{UserID: "u1", Role: "user", Content: "broken for the third time today", CreatedAt: now.Add(-1 * time.Hour)},
	}
	store := newMemStoreWithEntries(t, entries)
	detector := NewDriftDetector(store, "u1")

	signals, err := detector.Scan(t.Context())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	var found bool
	for _, s := range signals {
		if s.Type == "repetitive_failure" && s.Severity == "high" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected high-severity repetitive_failure signal, got: %+v", signals)
	}
}

func TestDriftDetectorMissedFollowup(t *testing.T) {
	oldTime := time.Now().Add(-72 * time.Hour)
	entries := []memory.Memory{
		{UserID: "u1", Role: "user", Content: "need to follow up with the client about invoice", CreatedAt: oldTime},
	}
	store := newMemStoreWithEntries(t, entries)
	detector := NewDriftDetector(store, "u1")

	signals, err := detector.Scan(t.Context())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	var found bool
	for _, s := range signals {
		if s.Type == "missed_followup" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missed_followup signal, got: %+v", signals)
	}
}

func TestDriftFormatReportEmpty(t *testing.T) {
	store, _ := memory.New(t.TempDir())
	detector := NewDriftDetector(store, "u1")
	// No scan, no signals
	report := detector.FormatReport()
	if report == "" {
		t.Error("FormatReport should return non-empty string even with no signals")
	}
	if report != "✅ No drift detected — all work looks on track." {
		t.Errorf("unexpected clean report: %s", report)
	}
}

// --- fmtAge helper ---

func TestFmtAgeDays(t *testing.T) {
	result := fmtAge(50 * time.Hour)
	if result != "2 days" {
		t.Errorf("expected '2 days', got %q", result)
	}
}

func TestFmtAgeHours(t *testing.T) {
	result := fmtAge(3 * time.Hour)
	if result != "3 hours" {
		t.Errorf("expected '3 hours', got %q", result)
	}
}

func TestFmtAgeMinutes(t *testing.T) {
	result := fmtAge(45 * time.Minute)
	if result != "45 minutes" {
		t.Errorf("expected '45 minutes', got %q", result)
	}
}

// --- SessionBriefing ---

func TestSessionBriefingAfterAbsence(t *testing.T) {
	store, _ := memory.New(t.TempDir())
	sb := NewSessionBriefing(store, "u1")

	// Simulate absence: last seen 2 hours ago
	lastSeen := time.Now().Add(-2 * time.Hour)
	brief := sb.Brief(lastSeen)

	if brief == "" {
		t.Error("expected a non-empty session brief after 2h absence")
	}
}

func TestSessionBriefingRecentSession(t *testing.T) {
	store, _ := memory.New(t.TempDir())
	sb := NewSessionBriefing(store, "u1")

	// Just now — no absence
	brief := sb.Brief(time.Now())
	// Should return empty or minimal — no "welcome back" needed
	_ = brief // just ensure no panic
}
