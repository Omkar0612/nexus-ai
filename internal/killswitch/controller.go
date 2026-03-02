package killswitch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Controller manages the 3-layer kill-switch architecture for agents.
type Controller struct {
	mu              sync.RWMutex
	activeSessions  map[string]*Session
	credentialVault CredentialStore
	actionLog       ActionLogger
}

// Session represents an active agent execution that can be paused or killed.
type Session struct {
	ID              string
	AgentName       string
	StartTime       time.Time
	State           SessionState
	CredentialIDs   []string
	PendingActions  []RecoverableAction
}

type SessionState string

const (
	StateActive SessionState = "ACTIVE"
	StatePaused SessionState = "PAUSED"
	StateKilled SessionState = "KILLED"
)

// RecoverableAction represents a single agent action that can be rolled back.
type RecoverableAction struct {
	ID            string
	ToolName      string
	TargetSystem  string
	Payload       map[string]interface{}
	ExecutedAt    time.Time
	IdempotencyKey string
	UndoScript    string // Optional: rollback logic
}

// CredentialStore defines the interface for revoking agent credentials.
type CredentialStore interface {
	RevokeAll(credentialIDs []string) error
}

// ActionLogger persists every agent action for audit trails and rollback.
type ActionLogger interface {
	Log(action RecoverableAction) error
	GetHistory(sessionID string) ([]RecoverableAction, error)
}

// NewController initializes the kill-switch system.
func NewController(vault CredentialStore, logger ActionLogger) *Controller {
	return &Controller{
		activeSessions:  make(map[string]*Session),
		credentialVault: vault,
		actionLog:       logger,
	}
}

// RegisterSession tracks a new agent execution for kill-switch eligibility.
func (c *Controller) RegisterSession(sessionID, agentName string, credentialIDs []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.activeSessions[sessionID] = &Session{
		ID:             sessionID,
		AgentName:      agentName,
		StartTime:      time.Now(),
		State:          StateActive,
		CredentialIDs:  credentialIDs,
		PendingActions: make([]RecoverableAction, 0),
	}

	log.Info().Str("session_id", sessionID).Msg("🛡️ Session registered for kill-switch")
}

// Layer 1: Hard Stop — Instant credential revocation and queue drain.
func (c *Controller) HardStop(ctx context.Context, sessionID string, reason string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sess, exists := c.activeSessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	log.Warn().Str("session_id", sessionID).Str("reason", reason).Msg("🛑 HARD STOP triggered")

	// 1. Revoke all API credentials
	if err := c.credentialVault.RevokeAll(sess.CredentialIDs); err != nil {
		return fmt.Errorf("failed to revoke credentials: %w", err)
	}

	// 2. Mark session as killed
	sess.State = StateKilled

	// 3. Trigger post-mortem (async)
	go c.runPostMortem(sessionID, reason)

	return nil
}

// Layer 2: Soft Pause — Freeze execution but preserve state for review.
func (c *Controller) SoftPause(sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	sess, exists := c.activeSessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	sess.State = StatePaused
	log.Info().Str("session_id", sessionID).Msg("⏸️ Session paused (state preserved)")
	return nil
}

// Layer 3: Transactional Rollback — Undo agent actions idempotently.
func (c *Controller) Rollback(ctx context.Context, sessionID string) error {
	c.mu.RLock()
	_, exists := c.activeSessions[sessionID]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	actions, err := c.actionLog.GetHistory(sessionID)
	if err != nil {
		return fmt.Errorf("failed to retrieve action history: %w", err)
	}

	log.Info().Str("session_id", sessionID).Int("actions", len(actions)).Msg("↩️ Rolling back agent actions...")

	// Execute rollback in reverse chronological order
	for i := len(actions) - 1; i >= 0; i-- {
		action := actions[i]
		log.Debug().Str("action_id", action.ID).Str("tool", action.ToolName).Msg("Undoing action")

		// Placeholder: actual undo logic would call the inverse API or restore state
		// For example: if action was "create_record", rollback calls "delete_record"
	}

	log.Info().Str("session_id", sessionID).Msg("✅ Rollback complete")
	return nil
}

func (c *Controller) runPostMortem(sessionID, reason string) {
	log.Info().Str("session_id", sessionID).Msg("🔬 Post-mortem: classifying failure and updating test suite")

	// 1. Classify the failure (cost breach, loop, tool failure, etc.)
	// 2. Add the trace to a regression test dataset
	// 3. Update prompt validators or tool scopes if needed

	// Placeholder for actual implementation
}
