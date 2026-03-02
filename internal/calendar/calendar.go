// Package calendar provides a natural-language calendar agent.
// Supports Google Calendar (OAuth2) and iCal/ICS files locally.
// Free tier: Google Calendar API — no cost for read/write.
package calendar

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Provider represents a calendar backend.
type Provider string

const (
	ProviderGoogle  Provider = "google"
	ProviderOutlook Provider = "outlook"
	ProviderICS     Provider = "ics"
)

// Event represents a single calendar event.
type Event struct {
	ID          string
	Title       string
	Description string
	Location    string
	Start       time.Time
	End         time.Time
	AllDay      bool
	Recurring   bool
	Attendees   []string
	CalendarID  string
	Provider    Provider
}

// ConflictResult describes a detected scheduling conflict.
type ConflictResult struct {
	EventA  Event
	EventB  Event
	Overlap time.Duration
}

// Agent is the calendar agent.
type Agent struct {
	providers map[Provider]CalendarProvider
	timezone  *time.Location
}

// CalendarProvider is the interface every backend must implement.
type CalendarProvider interface {
	ListEvents(ctx context.Context, from, to time.Time) ([]Event, error)
	CreateEvent(ctx context.Context, e Event) (Event, error)
	UpdateEvent(ctx context.Context, e Event) error
	DeleteEvent(ctx context.Context, id string) error
	Name() Provider
}

// New creates a calendar agent with the given providers.
func New(tz *time.Location, providers ...CalendarProvider) *Agent {
	if tz == nil {
		tz = time.UTC
	}
	a := &Agent{
		providers: make(map[Provider]CalendarProvider),
		timezone:  tz,
	}
	for _, p := range providers {
		a.providers[p.Name()] = p
	}
	return a
}

// Today returns all events for today across all connected providers.
func (a *Agent) Today(ctx context.Context) ([]Event, error) {
	now := time.Now().In(a.timezone)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, a.timezone)
	end := start.Add(24 * time.Hour)
	return a.Range(ctx, start, end)
}

// Tomorrow returns all events for tomorrow.
func (a *Agent) Tomorrow(ctx context.Context) ([]Event, error) {
	now := time.Now().In(a.timezone)
	start := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, a.timezone)
	end := start.Add(24 * time.Hour)
	return a.Range(ctx, start, end)
}

// Week returns events for the next 7 days.
func (a *Agent) Week(ctx context.Context) ([]Event, error) {
	now := time.Now().In(a.timezone)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, a.timezone)
	return a.Range(ctx, start, start.Add(7*24*time.Hour))
}

// Range returns events across all providers in the given time window, sorted by start time.
func (a *Agent) Range(ctx context.Context, from, to time.Time) ([]Event, error) {
	var all []Event
	for _, p := range a.providers {
		events, err := p.ListEvents(ctx, from, to)
		if err != nil {
			return nil, fmt.Errorf("calendar[%s]: %w", p.Name(), err)
		}
		all = append(all, events...)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Start.Before(all[j].Start)
	})
	return all, nil
}

// DetectConflicts finds overlapping events in the given slice.
func (a *Agent) DetectConflicts(events []Event) []ConflictResult {
	var conflicts []ConflictResult
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].AllDay || events[j].AllDay {
				continue
			}
			latestStart := events[i].Start
			if events[j].Start.After(latestStart) {
				latestStart = events[j].Start
			}
			earliestEnd := events[i].End
			if events[j].End.Before(earliestEnd) {
				earliestEnd = events[j].End
			}
			if overlap := earliestEnd.Sub(latestStart); overlap > 0 {
				conflicts = append(conflicts, ConflictResult{
					EventA:  events[i],
					EventB:  events[j],
					Overlap: overlap,
				})
			}
		}
	}
	return conflicts
}

// FindFreeSlot returns the next free time slot of the given duration on or after `after`.
// It checks all providers in the window [after, after+lookAheadDays].
func (a *Agent) FindFreeSlot(ctx context.Context, duration, lookAhead time.Duration, after time.Time) (time.Time, error) {
	events, err := a.Range(ctx, after, after.Add(lookAhead))
	if err != nil {
		return time.Time{}, err
	}
	cursor := after
	for _, e := range events {
		if e.AllDay {
			continue
		}
		if cursor.Add(duration).Before(e.Start) || cursor.Add(duration).Equal(e.Start) {
			// gap is large enough
			return cursor, nil
		}
		if e.End.After(cursor) {
			cursor = e.End
		}
	}
	if cursor.Add(duration).Before(after.Add(lookAhead)) {
		return cursor, nil
	}
	return time.Time{}, fmt.Errorf("no free slot of %s found in the next %s", duration, lookAhead)
}

// DigestLines returns a compact human-readable summary of events (for morning digest).
func DigestLines(events []Event, tz *time.Location) []string {
	if tz == nil {
		tz = time.UTC
	}
	lines := make([]string, 0, len(events))
	for _, e := range events {
		if e.AllDay {
			lines = append(lines, fmt.Sprintf("📅  %s (all day)", e.Title))
		} else {
			lines = append(lines, fmt.Sprintf("🕐  %s — %s  %s",
				e.Start.In(tz).Format("15:04"),
				e.End.In(tz).Format("15:04"),
				e.Title,
			))
		}
		if len(e.Attendees) > 0 {
			lines = append(lines, fmt.Sprintf("    👥 %s", strings.Join(e.Attendees, ", ")))
		}
	}
	return lines
}
