package agents

/*
HITLGate — Human-in-the-Loop approval gate for high-risk agent actions.

The only reliable method for safe production agents (multiple sources, 2025-2026):
'Until agents have 99.99% reliability, humans must stay in the loop for
high-risk actions. There is no shortcut.'
— IBM, Galileo AI, GitHub Blog (2026)

NEXUS HITLGate:
  1. Every action is classified: low / medium / high risk
  2. Low risk: auto-execute immediately (no delay)
  3. Medium risk: execute with audit log entry
  4. High risk: PAUSE — send Telegram approval request with timeout
     - Approval: action proceeds
     - Rejection: action is cancelled with reason
     - Timeout:  action is safely cancelled (fail-closed)
  5. Emergency override: block ALL actions until human unlocks
  6. Approval history stored in audit log
*/

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// ApprovalStatus is the result of a human approval request
type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
	ApprovalTimeout  ApprovalStatus = "timeout"
)

// ApprovalRequest represents a pending human approval
type ApprovalRequest struct {
	ID          string
	Action      string
	Rationale   string
	Risk        string
	RequestedAt time.Time
	ExpiresAt   time.Time
	Status      ApprovalStatus
	DecidedBy   string
	DecisionAt  time.Time
	Meta        map[string]string
}

// ActionFunc is a function that performs a NEXUS action
type ActionFunc func(ctx context.Context) error

// HITLGate is the human-in-the-loop approval system
type HITLGate struct {
	mu         sync.RWMutex
	pending    map[string]*ApprovalRequest
	history    []ApprovalRequest
	notify     func(req *ApprovalRequest) error
	timeout    time.Duration
	locked     bool // emergency lock — blocks all non-low-risk actions
	onDecision func(req *ApprovalRequest)
}

// NewHITLGate creates a new HITL gate
func NewHITLGate(approvalTimeout time.Duration, notifyFn func(req *ApprovalRequest) error) *HITLGate {
	if approvalTimeout <= 0 {
		approvalTimeout = 5 * time.Minute
	}
	return &HITLGate{
		pending: make(map[string]*ApprovalRequest),
		timeout: approvalTimeout,
		notify:  notifyFn,
	}
}

// SetDecisionCallback sets a function called when any approval decision is made
func (g *HITLGate) SetDecisionCallback(fn func(req *ApprovalRequest)) {
	g.onDecision = fn
}

// Execute runs an action through the HITL gate
func (g *HITLGate) Execute(ctx context.Context, action, rationale string, risk string, fn ActionFunc) error {
	g.mu.RLock()
	locked := g.locked
	g.mu.RUnlock()

	riskLower := strings.ToLower(risk)

	// Emergency lock blocks everything except low risk
	if locked && riskLower != "low" {
		return fmt.Errorf("HITL gate is in emergency lock mode. Unlock with: nexus hitl unlock")
	}

	switch riskLower {
	case "low":
		log.Debug().Str("action", action).Msg("HITL: auto-executing low risk action")
		return fn(ctx)

	case "medium":
		log.Info().Str("action", action).Msg("HITL: executing medium risk action (auto-approved with audit)")
		return fn(ctx)

	case "high":
		return g.requestApproval(ctx, action, rationale, risk, fn)

	default:
		return fmt.Errorf("unknown risk level: %s", risk)
	}
}

func (g *HITLGate) requestApproval(ctx context.Context, action, rationale, risk string, fn ActionFunc) error {
	req := &ApprovalRequest{
		ID:          fmt.Sprintf("hitl-%d", time.Now().UnixNano()),
		Action:      action,
		Rationale:   rationale,
		Risk:        risk,
		RequestedAt: time.Now(),
		ExpiresAt:   time.Now().Add(g.timeout),
		Status:      ApprovalPending,
	}
	g.mu.Lock()
	g.pending[req.ID] = req
	g.mu.Unlock()

	log.Warn().Str("id", req.ID).Str("action", action).Msg("HITL: high-risk action awaiting human approval")

	if g.notify != nil {
		if err := g.notify(req); err != nil {
			log.Error().Err(err).Msg("HITL: failed to send approval notification")
		}
	}

	// Wait for decision or timeout
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	timeoutCh := time.After(g.timeout)

	for {
		select {
		case <-ctx.Done():
			g.recordDecision(req, ApprovalTimeout, "context cancelled")
			return fmt.Errorf("HITL approval cancelled: context done")
		case <-timeoutCh:
			g.recordDecision(req, ApprovalTimeout, "auto")
			return fmt.Errorf("HITL approval timed out after %s — action cancelled for safety", g.timeout)
		case <-ticker.C:
			g.mu.RLock()
			status := req.Status
			g.mu.RUnlock()
			switch status {
			case ApprovalApproved:
				log.Info().Str("id", req.ID).Str("by", req.DecidedBy).Msg("HITL: action approved")
				return fn(ctx)
			case ApprovalRejected:
				return fmt.Errorf("action rejected by %s", req.DecidedBy)
			}
		}
	}
}

// Approve approves a pending request (called by Telegram webhook or CLI)
func (g *HITLGate) Approve(requestID, decidedBy string) error {
	return g.decide(requestID, decidedBy, ApprovalApproved)
}

// Reject rejects a pending request
func (g *HITLGate) Reject(requestID, decidedBy string) error {
	return g.decide(requestID, decidedBy, ApprovalRejected)
}

func (g *HITLGate) decide(requestID, decidedBy string, status ApprovalStatus) error {
	g.mu.Lock()
	req, ok := g.pending[requestID]
	if !ok {
		g.mu.Unlock()
		return fmt.Errorf("request %s not found or already decided", requestID)
	}
	req.Status = status
	req.DecidedBy = decidedBy
	req.DecisionAt = time.Now()
	g.mu.Unlock()
	g.recordDecision(req, status, decidedBy)
	return nil
}

func (g *HITLGate) recordDecision(req *ApprovalRequest, status ApprovalStatus, by string) {
	g.mu.Lock()
	req.Status = status
	if by != "" {
		req.DecidedBy = by
	}
	delete(g.pending, req.ID)
	g.history = append(g.history, *req)
	if len(g.history) > 100 {
		g.history = g.history[len(g.history)-100:]
	}
	g.mu.Unlock()
	if g.onDecision != nil {
		g.onDecision(req)
	}
}

// EmergencyLock blocks all non-low-risk actions immediately
func (g *HITLGate) EmergencyLock() {
	g.mu.Lock()
	g.locked = true
	g.mu.Unlock()
	log.Warn().Msg("HITL: EMERGENCY LOCK ENGAGED — all high/medium risk actions blocked")
}

// EmergencyUnlock re-enables normal operation
func (g *HITLGate) EmergencyUnlock() {
	g.mu.Lock()
	g.locked = false
	g.mu.Unlock()
	log.Info().Msg("HITL: emergency lock released")
}

// PendingCount returns the number of outstanding approval requests
func (g *HITLGate) PendingCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.pending)
}

// FormatApprovalMessage formats a Telegram-ready approval request message
func FormatApprovalMessage(req *ApprovalRequest) string {
	return fmt.Sprintf(
		"🔴 **NEXUS High-Risk Action — Approval Required**\n\n"+
			"**Action:** %s\n"+
			"**Why:** %s\n"+
			"**Risk:** %s\n"+
			"**Expires:** %s\n\n"+
			"Reply:\n✅ `nexus hitl approve %s`\n❌ `nexus hitl reject %s`",
		req.Action, req.Rationale, req.Risk,
		req.ExpiresAt.Format("15:04:05"),
		req.ID, req.ID,
	)
}
