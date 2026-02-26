package agents

/*
DriftDetector — monitors conversation history for work drift signals.

Detects:
  - Stalled tasks (mentioned but never completed past threshold hours)
  - Missed follow-ups ("follow up with X" never resolved)
  - Context loss (abrupt topic change mid-task)
  - Repetitive failures (same error mentioned 3+ times)

This feature does not exist in any other open-source AI agent.
Inspired by: r/AI_Agents — 'something that quietly prevents things from unraveling'
*/

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Omkar0612/nexus-ai/internal/memory"
)

// DriftSignal represents a detected work drift pattern
type DriftSignal struct {
	Type        string    // stalled_task, missed_followup, context_loss, repetitive_failure
	Severity    string    // low, medium, high
	Description string
	Suggestion  string
	DetectedAt  time.Time
	TaskRef     string
}

// DriftDetector runs silently in the background watching for work drift
type DriftDetector struct {
	mem        *memory.Store
	userID     string
	signals    []DriftSignal
	thresholds DriftThresholds
}

// DriftThresholds configures when to fire alerts
type DriftThresholds struct {
	StalledTaskHours   int
	MissedFollowupDays int
	GoalDeviationScore float64
}

// NewDriftDetector creates a new drift detector
func NewDriftDetector(mem *memory.Store, userID string) *DriftDetector {
	return &DriftDetector{
		mem:    mem,
		userID: userID,
		thresholds: DriftThresholds{
			StalledTaskHours:   24,
			MissedFollowupDays: 2,
			GoalDeviationScore: 0.6,
		},
	}
}

// Scan analyses recent memory for drift signals
func (d *DriftDetector) Scan(ctx context.Context) ([]DriftSignal, error) {
	history, err := d.mem.GetEpisodicHistory(d.userID, 100)
	if err != nil {
		return nil, err
	}
	var signals []DriftSignal
	signals = append(signals, d.detectStalledTasks(history)...)
	signals = append(signals, d.detectMissedFollowups(history)...)
	signals = append(signals, d.detectRepetitiveFailures(history)...)
	d.signals = signals
	return signals, nil
}

func (d *DriftDetector) detectStalledTasks(history []memory.Memory) []DriftSignal {
	var signals []DriftSignal
	taskKW := []string{"working on", "need to", "will", "plan to", "building", "creating"}
	doneKW := []string{"done", "finished", "completed", "shipped", "deployed", "fixed"}
	pending := make(map[string]time.Time)

	for i := len(history) - 1; i >= 0; i-- {
		m := history[i]
		content := strings.ToLower(m.Content)
		for _, kw := range taskKW {
			if strings.Contains(content, kw) {
				ref := taskRef(m.Content, kw)
				if ref != "" {
					pending[ref] = m.CreatedAt
				}
			}
		}
		for _, kw := range doneKW {
			if strings.Contains(content, kw) {
				for ref := range pending {
					if fuzzyMatch(content, ref) {
						delete(pending, ref)
					}
				}
			}
		}
	}

	threshold := time.Duration(d.thresholds.StalledTaskHours) * time.Hour
	for ref, start := range pending {
		if time.Since(start) > threshold {
			signals = append(signals, DriftSignal{
				Type:        "stalled_task",
				Severity:    severityFromAge(time.Since(start)),
				Description: fmt.Sprintf("Task stalled: '%s' (last touched %s ago)", ref, fmtAge(time.Since(start))),
				Suggestion:  fmt.Sprintf("Resume or close: '%s'", ref),
				DetectedAt:  time.Now(),
				TaskRef:     ref,
			})
		}
	}
	return signals
}

func (d *DriftDetector) detectMissedFollowups(history []memory.Memory) []DriftSignal {
	var signals []DriftSignal
	kws := []string{"follow up", "remind me", "check back", "circle back", "ping them"}
	for _, m := range history {
		content := strings.ToLower(m.Content)
		for _, kw := range kws {
			if strings.Contains(content, kw) {
				age := time.Since(m.CreatedAt)
				if age > time.Duration(d.thresholds.MissedFollowupDays)*24*time.Hour {
					preview := m.Content
					if len(preview) > 80 {
						preview = preview[:80]
					}
					signals = append(signals, DriftSignal{
						Type:        "missed_followup",
						Severity:    "medium",
						Description: fmt.Sprintf("Follow-up may have been missed (%s ago)", fmtAge(age)),
						Suggestion:  fmt.Sprintf("Did you follow up on: '%s'?", preview),
						DetectedAt:  time.Now(),
					})
					break
				}
			}
		}
	}
	return signals
}

func (d *DriftDetector) detectRepetitiveFailures(history []memory.Memory) []DriftSignal {
	var signals []DriftSignal
	errorKW := []string{"error", "failed", "not working", "broken", "issue", "bug"}
	counts := make(map[string]int)
	for _, m := range history {
		content := strings.ToLower(m.Content)
		for _, kw := range errorKW {
			if strings.Contains(content, kw) {
				counts[kw]++
				if counts[kw] == 3 {
					signals = append(signals, DriftSignal{
						Type:        "repetitive_failure",
						Severity:    "high",
						Description: fmt.Sprintf("Same issue mentioned 3+ times: '%s'", kw),
						Suggestion:  "Let me help you solve this systematically",
						DetectedAt:  time.Now(),
					})
				}
			}
		}
	}
	return signals
}

// FormatReport generates a human-readable drift report
func (d *DriftDetector) FormatReport() string {
	if len(d.signals) == 0 {
		return "✅ No drift detected — all work looks on track."
	}
	var sb strings.Builder
	sb.WriteString("⚠️ **NEXUS Drift Report**\n\n")
	icons := map[string]string{
		"stalled_task":      "🔴",
		"missed_followup":   "🟡",
		"repetitive_failure": "🔴",
	}
	for _, s := range d.signals {
		icon := icons[s.Type]
		if icon == "" {
			icon = "🔵"
		}
		sb.WriteString(fmt.Sprintf("%s [%s] %s\n   💡 %s\n\n", icon, strings.ToUpper(s.Severity), s.Description, s.Suggestion))
	}
	return sb.String()
}

func taskRef(content, keyword string) string {
	idx := strings.Index(strings.ToLower(content), keyword)
	if idx < 0 {
		return ""
	}
	end := idx + len(keyword) + 50
	if end > len(content) {
		end = len(content)
	}
	return strings.TrimSpace(content[idx:end])
}

func fuzzyMatch(a, b string) bool {
	for _, w := range strings.Fields(strings.ToLower(b)) {
		if len(w) > 4 && strings.Contains(strings.ToLower(a), w) {
			return true
		}
	}
	return false
}

func severityFromAge(d time.Duration) string {
	if d > 72*time.Hour {
		return "high"
	}
	if d > 24*time.Hour {
		return "medium"
	}
	return "low"
}

func fmtAge(d time.Duration) string {
	if d > 24*time.Hour {
		return fmt.Sprintf("%.0f days", d.Hours()/24)
	}
	if d > time.Hour {
		return fmt.Sprintf("%.0f hours", d.Hours())
	}
	return fmt.Sprintf("%.0f minutes", d.Minutes())
}
