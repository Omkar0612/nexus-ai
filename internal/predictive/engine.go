package predictive

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// LLMClient represents an interface to Groq/OpenAI for generating the actual reports.
type LLMClient interface {
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

// Engine runs continuously in the background, monitoring EventSources and dispatching
// heavy LLM tasks *before* the user explicitly asks for them.
type Engine struct {
	sources []EventSource
	cache   Cache
	llm     LLMClient
	ticker  *time.Ticker
	quit    chan struct{}
}

// NewEngine initializes the Predictive Pre-Computation engine.
func NewEngine(llm LLMClient, cache Cache, interval time.Duration) *Engine {
	return &Engine{
		sources: make([]EventSource, 0),
		cache:   cache,
		llm:     llm,
		ticker:  time.NewTicker(interval),
		quit:    make(chan struct{}),
	}
}

// RegisterSource adds a new observer (e.g., CalendarPoller, GitHubWebhookReceiver)
func (e *Engine) RegisterSource(src EventSource) {
	e.sources = append(e.sources, src)
}

// Start begins the background polling and computation loop.
func (e *Engine) Start(ctx context.Context) {
	log.Info().Msg("🧠 Starting Predictive Pre-Computation Engine (Zero-Latency mode)")
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-e.quit:
				return
			case <-e.ticker.C:
				e.pollAndCompute(ctx)
			}
		}
	}()
}

// Stop safely halts the background engine.
func (e *Engine) Stop() {
	close(e.quit)
	e.ticker.Stop()
}

func (e *Engine) pollAndCompute(ctx context.Context) {
	for _, src := range e.sources {
		events, err := src.Poll()
		if err != nil {
			log.Warn().Err(err).Str("source", src.Name()).Msg("Failed to poll event source")
			continue
		}

		for _, ev := range events {
			// In a real system, check if ev.ID is already in Cache to avoid duplicate work
			log.Info().Str("event_type", string(ev.Type)).Msg("Predictive Engine triggered. Beginning background computation...")
			
			// Kick off heavy LLM work in a background goroutine
			go e.computeTask(context.Background(), ev)
		}
	}
}

func (e *Engine) computeTask(ctx context.Context, ev Event) {
	var result *PrecomputedResult
	var err error

	switch ev.Type {
	case EventUpcomingMeeting:
		result, err = e.handleMeeting(ctx, ev)
	case EventBrokenBuild:
		result, err = e.handleBrokenBuild(ctx, ev)
	default:
		log.Debug().Str("type", string(ev.Type)).Msg("No predictive handler for event type")
		return
	}

	if err != nil {
		log.Error().Err(err).Str("event_id", ev.ID).Msg("Failed background pre-computation")
		return
	}

	if err := e.cache.Store(result); err != nil {
		log.Error().Err(err).Msg("Failed to cache predictive result")
		return
	}

	log.Info().
		Str("event_id", ev.ID).
		Str("action", result.Action).
		Msg("✅ Pre-computation complete. Result cached for instant UI rendering.")
}

func (e *Engine) handleMeeting(ctx context.Context, ev Event) (*PrecomputedResult, error) {
	clientName, _ := ev.Payload["client_name"].(string)
	company, _ := ev.Payload["company"].(string)

	sys := "You are NEXUS, an expert executive assistant. Create a 3-bullet meeting brief based on the client's company. Output only the brief."
	prompt := fmt.Sprintf("I am meeting with %s from %s in 1 hour. Give me a quick brief and recent news.", clientName, company)

	// Simulate Deep Research / Web Search call here
	output, err := e.llm.Generate(ctx, sys, prompt)
	if err != nil {
		return nil, err
	}

	return &PrecomputedResult{
		EventID:   ev.ID,
		Summary:   fmt.Sprintf("Meeting Brief: %s (%s)", clientName, company),
		Action:    "Review Brief",
		Data:      []byte(output),
		CreatedAt: time.Now(),
		Viewed:    false,
	}, nil
}

func (e *Engine) handleBrokenBuild(ctx context.Context, ev Event) (*PrecomputedResult, error) {
	repo, _ := ev.Payload["repository"].(string)
	errorLog, _ := ev.Payload["error_log"].(string)

	sys := "You are NEXUS, an expert DevOps engineer. Analyze the build log and write a git patch to fix the error."
	prompt := fmt.Sprintf("Repo: %s\nError:\n%s\nProvide the code fix.", repo, errorLog)

	output, err := e.llm.Generate(ctx, sys, prompt)
	if err != nil {
		return nil, err
	}

	return &PrecomputedResult{
		EventID:   ev.ID,
		Summary:   fmt.Sprintf("Fix ready for broken build in %s", repo),
		Action:    "Apply Git Patch",
		Data:      []byte(output),
		CreatedAt: time.Now(),
		Viewed:    false,
	}, nil
}
