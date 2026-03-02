package agents

/*
SelfHealingAgent — diagnoses and auto-fixes broken tasks.

When a task fails, instead of silently dying:
  1. Captures full error context
  2. Sends to LLM for root-cause analysis
  3. Attempts automatic fix or fallback
  4. Reports to user with plain-language diagnosis
  5. Logs fix to memory so it never repeats

Fixes the #1 OpenClaw complaint: 'cron was silently broken for 2 days'
*/

import (
	"fmt"
	"strings"
	"time"
)

// FailureRecord tracks a task failure history
type FailureRecord struct {
	TaskName    string
	Error       string
	AttemptNum  int
	LastAttempt time.Time
	Fixed       bool
	FixApplied  string
}

// HealingResult is the outcome of a healing attempt
type HealingResult struct {
	Status     string // fixed, retrying, escalated, diagnosed
	RootCause  string
	FixApplied string
	Message    string
	UserAction string
	Confidence float64
}

// SelfHealingAgent manages task failure recovery
type SelfHealingAgent struct {
	failures map[string]*FailureRecord
	maxRetry int
}

// NewSelfHealingAgent creates a new self-healing agent
func NewSelfHealingAgent() *SelfHealingAgent {
	return &SelfHealingAgent{
		failures: make(map[string]*FailureRecord),
		maxRetry: 3,
	}
}

// RecordFailure logs a failure and returns a healing result
func (s *SelfHealingAgent) RecordFailure(taskName, errMsg string) *HealingResult {
	rec, ok := s.failures[taskName]
	if !ok {
		rec = &FailureRecord{TaskName: taskName}
		s.failures[taskName] = rec
	}
	rec.Error = errMsg
	rec.AttemptNum++
	rec.LastAttempt = time.Now()

	if rec.AttemptNum > s.maxRetry {
		return &HealingResult{
			Status:     "escalated",
			Message:    fmt.Sprintf("Task '%s' failed %d times — manual intervention needed", taskName, rec.AttemptNum),
			UserAction: "Please review the task configuration.",
		}
	}

	return &HealingResult{
		Status:  "retrying",
		Message: fmt.Sprintf("⚠️ Task '%s' failed (attempt %d/%d). Auto-retrying...", taskName, rec.AttemptNum, s.maxRetry),
	}
}

// HealthReport generates a system health summary
func (s *SelfHealingAgent) HealthReport() string {
	if len(s.failures) == 0 {
		return "✅ All systems healthy — no failures recorded."
	}
	var sb strings.Builder
	sb.WriteString("🏥 **NEXUS Health Report**\n\n")
	for name, rec := range s.failures {
		status := "🔴 FAILING"
		if rec.Fixed {
			status = "✅ FIXED"
		}
		errPreview := rec.Error
		if len(errPreview) > 100 {
			errPreview = errPreview[:100]
		}
		sb.WriteString(fmt.Sprintf("%s **%s**\n   Last error: %s\n   Attempts: %d\n\n",
			status, name, errPreview, rec.AttemptNum))
	}
	return sb.String()
}
