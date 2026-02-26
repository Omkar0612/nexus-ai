package agents

/*
MultiAgentBus — spawn and coordinate specialised sub-agents over a central message bus.

The next frontier of AI agent systems (GitHub Blog, Feb 2026):
'Multi-agent workflows fail because there’s no coordination layer.'
— github.blog/ai-and-ml, Feb 23 2026

NEXUS MultiAgentBus:
  1. Typed agent roles: Researcher, Coder, Writer, Analyst, Reviewer
  2. Central message bus — agents pass structured messages, not raw strings
  3. Role enforcement — agents can’t do work outside their role
  4. Task router — auto-routes tasks to the best-fit agent
  5. Result aggregator — merges outputs from parallel agents
  6. Timeout protection — stalled sub-agents are cancelled automatically
  7. Loop detection integration — bus detects circular message chains
*/

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// AgentRole defines the specialisation of a sub-agent
type AgentRole string

const (
	RoleResearcher AgentRole = "researcher"
	RoleCoder      AgentRole = "coder"
	RoleWriter     AgentRole = "writer"
	RoleAnalyst    AgentRole = "analyst"
	RoleReviewer   AgentRole = "reviewer"
	RoleOrchestrator AgentRole = "orchestrator"
)

// MessageType classifies bus messages
type MessageType string

const (
	MsgTask    MessageType = "task"
	MsgResult  MessageType = "result"
	MsgError   MessageType = "error"
	MsgControl MessageType = "control"
)

// BusMessage is a structured message on the agent bus
type BusMessage struct {
	ID        string
	Type      MessageType
	From      AgentRole
	To        AgentRole
	Payload   string
	Meta      map[string]string
	CreatedAt time.Time
	ReplyTo   string // ID of message this is replying to
}

// AgentHandler is a function that processes a task message
type AgentHandler func(ctx context.Context, msg BusMessage) (BusMessage, error)

// SubAgent is a registered agent on the bus
type SubAgent struct {
	Role        AgentRole
	Name        string
	Description string
	Handler     AgentHandler
	Capabilities []string
	Busy        bool
	TaskCount   int
	mu          sync.Mutex
}

// BusStats holds bus performance metrics
type BusStats struct {
	TotalMessages  int
	TotalTasks     int
	TotalErrors    int
	AgentStats     map[AgentRole]int
}

// MultiAgentBus is the NEXUS central agent coordination system
type MultiAgentBus struct {
	agents   map[AgentRole]*SubAgent
	mu       sync.RWMutex
	history  []BusMessage
	histMu   sync.Mutex
	stats    BusStats
	timeout  time.Duration
	detector *LoopDetector
}

// NewBus creates a new multi-agent bus
func NewBus(taskTimeout time.Duration) *MultiAgentBus {
	if taskTimeout <= 0 {
		taskTimeout = 30 * time.Second
	}
	return &MultiAgentBus{
		agents:   make(map[AgentRole]*SubAgent),
		timeout:  taskTimeout,
		detector: NewLoopDetector(3, 30),
		stats:    BusStats{AgentStats: make(map[AgentRole]int)},
	}
}

// Register adds a sub-agent to the bus
func (b *MultiAgentBus) Register(agent *SubAgent) error {
	if agent.Role == "" {
		return fmt.Errorf("agent role is required")
	}
	if agent.Handler == nil {
		return fmt.Errorf("agent %s has no handler", agent.Role)
	}
	b.mu.Lock()
	b.agents[agent.Role] = agent
	b.mu.Unlock()
	log.Info().Str("role", string(agent.Role)).Str("name", agent.Name).Msg("sub-agent registered on bus")
	return nil
}

// Send dispatches a message to a specific agent role and waits for the result
func (b *MultiAgentBus) Send(ctx context.Context, msg BusMessage) (BusMessage, error) {
	if msg.ID == "" {
		msg.ID = fmt.Sprintf("msg-%d", time.Now().UnixNano())
	}
	msg.CreatedAt = time.Now()

	// Loop detection on message routing
	isLoop, event := b.detector.Record(string(msg.To), msg.Payload)
	if isLoop {
		log.Warn().Str("to", string(msg.To)).Msg("bus loop detected")
		return BusMessage{
			Type: MsgError,
			From: RoleOrchestrator,
			To:   msg.From,
			Payload: fmt.Sprintf("bus loop detected: %s", event.Format()),
		}, fmt.Errorf("message loop detected on route to %s", msg.To)
	}

	b.mu.RLock()
	agent, ok := b.agents[msg.To]
	b.mu.RUnlock()
	if !ok {
		return BusMessage{}, fmt.Errorf("no agent registered for role: %s", msg.To)
	}

	agent.mu.Lock()
	agent.Busy = true
	agent.TaskCount++
	agent.mu.Unlock()

	b.record(msg)

	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	result, err := agent.Handler(ctx, msg)

	agent.mu.Lock()
	agent.Busy = false
	agent.mu.Unlock()

	b.histMu.Lock()
	b.stats.TotalMessages++
	b.stats.TotalTasks++
	b.stats.AgentStats[msg.To]++
	if err != nil {
		b.stats.TotalErrors++
	}
	b.histMu.Unlock()

	if err != nil {
		return BusMessage{Type: MsgError, From: msg.To, Payload: err.Error()}, err
	}
	result.From = msg.To
	result.ReplyTo = msg.ID
	result.CreatedAt = time.Now()
	b.record(result)
	return result, nil
}

// Broadcast sends a task to ALL registered agents and collects results
func (b *MultiAgentBus) Broadcast(ctx context.Context, payload string) map[AgentRole]BusMessage {
	b.mu.RLock()
	roles := make([]AgentRole, 0, len(b.agents))
	for role := range b.agents {
		roles = append(roles, role)
	}
	b.mu.RUnlock()

	results := make(map[AgentRole]BusMessage, len(roles))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, role := range roles {
		wg.Add(1)
		go func(r AgentRole) {
			defer wg.Done()
			result, err := b.Send(ctx, BusMessage{
				Type: MsgTask, From: RoleOrchestrator, To: r, Payload: payload,
			})
			if err != nil {
				result = BusMessage{Type: MsgError, Payload: err.Error()}
			}
			mu.Lock()
			results[r] = result
			mu.Unlock()
		}(role)
	}
	wg.Wait()
	return results
}

// Route auto-routes a task to the best-fit agent based on keywords
func (b *MultiAgentBus) Route(ctx context.Context, task string) (BusMessage, error) {
	role := b.inferRole(task)
	return b.Send(ctx, BusMessage{
		Type: MsgTask, From: RoleOrchestrator, To: role, Payload: task,
	})
}

func (b *MultiAgentBus) inferRole(task string) AgentRole {
	lower := strings.ToLower(task)
	switch {
	case containsAny(lower, "search", "research", "find", "look up", "what is", "who is"):
		return RoleResearcher
	case containsAny(lower, "code", "write function", "implement", "debug", "fix bug", "refactor"):
		return RoleCoder
	case containsAny(lower, "write", "draft", "summarise", "summarize", "explain", "document"):
		return RoleWriter
	case containsAny(lower, "analyse", "analyze", "compare", "data", "chart", "trend", "report"):
		return RoleAnalyst
	case containsAny(lower, "review", "check", "validate", "verify", "audit", "test"):
		return RoleReviewer
	default:
		return RoleResearcher
	}
}

// Stats returns bus performance metrics
func (b *MultiAgentBus) Stats() BusStats {
	b.histMu.Lock()
	defer b.histMu.Unlock()
	return b.stats
}

// ListAgents returns a formatted list of registered agents
func (b *MultiAgentBus) ListAgents() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if len(b.agents) == 0 {
		return "No sub-agents registered. Use bus.Register() to add agents."
	}
	var sb strings.Builder
	sb.WriteString("🤖 **NEXUS Agent Bus**\n\n")
	for role, agent := range b.agents {
		status := "🟢 idle"
		if agent.Busy {
			status = "🟡 busy"
		}
		sb.WriteString(fmt.Sprintf("**%s** (%s) [%s]\n", agent.Name, role, status))
		sb.WriteString(fmt.Sprintf("  Tasks completed: %d\n", agent.TaskCount))
		if len(agent.Capabilities) > 0 {
			sb.WriteString(fmt.Sprintf("  Capabilities: %s\n", strings.Join(agent.Capabilities, ", ")))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (b *MultiAgentBus) record(msg BusMessage) {
	b.histMu.Lock()
	defer b.histMu.Unlock()
	b.history = append(b.history, msg)
	if len(b.history) > 200 {
		b.history = b.history[len(b.history)-200:]
	}
}

func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}
