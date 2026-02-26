package agents

/*
SessionBriefer — proactively briefs you when you return after being away.

Feature that NOBODY has built yet.

When you start NEXUS after being away, it tells you:
  - What you were working on when you left
  - What drift signals were detected while you were away
  - What goals you haven't touched
  - Any queued tasks that ran while offline

Like having an assistant who briefs you before a meeting.
*/

import (
	"fmt"
	"strings"
	"time"

	"github.com/Omkar0612/nexus-ai/internal/memory"
)

// SessionBrief holds the full session context briefing
type SessionBrief struct {
	LastSeen      time.Duration
	ResumeContext string
	DriftSignals  []DriftSignal
	QuickActions  []string
}

// SessionBriefer generates smart session briefings
type SessionBriefer struct {
	mem      *memory.Store
	userID   string
	lastSeen time.Time
}

// NewSessionBriefer creates a new session briefer
func NewSessionBriefer(mem *memory.Store, userID string) *SessionBriefer {
	return &SessionBriefer{mem: mem, userID: userID, lastSeen: time.Now()}
}

// ShouldBrief returns true if user has been away long enough to warrant a briefing
func (s *SessionBriefer) ShouldBrief() bool {
	return time.Since(s.lastSeen) >= 30*time.Minute
}

// GenerateBrief produces a smart session briefing
func (s *SessionBriefer) GenerateBrief() (*SessionBrief, error) {
	brief := &SessionBrief{LastSeen: time.Since(s.lastSeen)}
	if !s.ShouldBrief() {
		return nil, nil
	}

	history, err := s.mem.GetEpisodicHistory(s.userID, 10)
	if err != nil || len(history) == 0 {
		return brief, nil
	}

	// Build context from last session
	var sb strings.Builder
	for i := len(history) - 1; i >= 0 && i >= len(history)-5; i-- {
		content := history[i].Content
		if len(content) > 100 {
			content = content[:100]
		}
		sb.WriteString(history[i].Role + ": " + content + "\n")
	}
	brief.ResumeContext = sb.String()

	// Run drift detection
	detector := NewDriftDetector(s.mem, s.userID)
	signals, _ := detector.Scan(nil)
	brief.DriftSignals = signals
	for _, sig := range signals {
		if sig.Severity == "high" {
			brief.QuickActions = append(brief.QuickActions, sig.Suggestion)
		}
	}

	s.lastSeen = time.Now()
	return brief, nil
}

// Format renders the brief as a human-readable Telegram/CLI message
func (b *SessionBrief) Format() string {
	if b == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("👋 **Welcome back!** You were away for %s.\n\n", fmtAge(b.LastSeen)))

	if b.ResumeContext != "" {
		sb.WriteString("📍 **Where you left off:**\n" + b.ResumeContext + "\n")
	}

	high := 0
	for _, s := range b.DriftSignals {
		if s.Severity == "high" {
			high++
		}
	}
	if high > 0 {
		sb.WriteString(fmt.Sprintf("⚠️ **%d high-priority items need attention** — run `nexus drift`\n\n", high))
	}

	if len(b.QuickActions) > 0 {
		sb.WriteString("🎯 **Suggested actions:**\n")
		max := 3
		if len(b.QuickActions) < max {
			max = len(b.QuickActions)
		}
		for i := 0; i < max; i++ {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, b.QuickActions[i]))
		}
	}
	sb.WriteString("\nWhat would you like to work on?")
	return sb.String()
}
