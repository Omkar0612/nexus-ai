package calendar

import (
	"context"
	"testing"
	"time"
)

// --- stub provider ---

type stubProvider struct {
	events []Event
}

func (s *stubProvider) Name() Provider { return "stub" }
func (s *stubProvider) ListEvents(_ context.Context, from, to time.Time) ([]Event, error) {
	var out []Event
	for _, e := range s.events {
		if !e.End.Before(from) && !e.Start.After(to) {
			out = append(out, e)
		}
	}
	return out, nil
}
func (s *stubProvider) CreateEvent(_ context.Context, e Event) (Event, error) { return e, nil }
func (s *stubProvider) UpdateEvent(_ context.Context, _ Event) error          { return nil }
func (s *stubProvider) DeleteEvent(_ context.Context, _ string) error         { return nil }

func makeAgent(events []Event) *Agent {
	return New(time.UTC, &stubProvider{events: events})
}

var (
	t9  = time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
	t10 = time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	t11 = time.Date(2026, 3, 1, 11, 0, 0, 0, time.UTC)
	t12 = time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	t13 = time.Date(2026, 3, 1, 13, 0, 0, 0, time.UTC)
)

func TestRangeSorted(t *testing.T) {
	events := []Event{
		{ID: "b", Title: "B", Start: t10, End: t11},
		{ID: "a", Title: "A", Start: t9, End: t10},
	}
	agent := makeAgent(events)
	got, err := agent.Range(context.Background(), t9, t12)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
	if got[0].ID != "a" {
		t.Errorf("expected first event to be 'a', got '%s'", got[0].ID)
	}
}

func TestDetectConflicts(t *testing.T) {
	events := []Event{
		{ID: "1", Start: t9, End: t11},
		{ID: "2", Start: t10, End: t12}, // overlaps with 1 by 1h
		{ID: "3", Start: t12, End: t13}, // starts when 2 ends — no overlap with anyone
	}
	agent := makeAgent(nil)
	conflicts := agent.DetectConflicts(events)
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0].Overlap != time.Hour {
		t.Errorf("expected 1h overlap, got %s", conflicts[0].Overlap)
	}
}

func TestFindFreeSlot(t *testing.T) {
	events := []Event{
		{ID: "1", Start: t9, End: t11},
		{ID: "2", Start: t11, End: t12},
	}
	agent := makeAgent(events)
	slot, err := agent.FindFreeSlot(context.Background(), 30*time.Minute, 8*time.Hour, t9)
	if err != nil {
		t.Fatal(err)
	}
	if slot.Before(t12) {
		t.Errorf("expected free slot at or after %s, got %s", t12, slot)
	}
}

func TestDigestLines(t *testing.T) {
	events := []Event{
		{Title: "Standup", Start: t9, End: t10},
		{Title: "All day conf", AllDay: true},
	}
	lines := DigestLines(events, time.UTC)
	if len(lines) < 2 {
		t.Errorf("expected at least 2 digest lines, got %d", len(lines))
	}
}
