package shadow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Metrics tracks the performance and cost of an agent workflow run.
type Metrics struct {
	Cost    float64
	Tokens  int
	Latency time.Duration
	Output  string
}

// Task defines an executable agent workflow that returns metrics.
type Task func(ctx context.Context) (*Metrics, error)

// HITLGate abstracts the existing Human-in-the-Loop prompt system 
// allowing Shadow Mode to ask for permission before modifying configs.
type HITLGate interface {
	AskPermission(ctx context.Context, riskLevel, prompt string) bool
}

// Engine runs experimental prompts/models in a hidden sandbox, compares
// their efficiency to production, and prompts the user if it finds an optimization.
type Engine struct {
	hitl HITLGate
}

// New creates a new Shadow Mode evolution engine.
func New(hitl HITLGate) *Engine {
	return &Engine{hitl: hitl}
}

// EvaluateAsync runs the shadow task in the background. If it outperforms
// the baseline, it uses the HITL gate to request an upgrade.
func (e *Engine) EvaluateAsync(ctx context.Context, taskName string, baseline, shadow Task) {
	go func() {
		if err := e.Evaluate(ctx, taskName, baseline, shadow); err != nil {
			log.Error().Err(err).Str("task", taskName).Msg("Shadow evaluation failed")
		}
	}()
}

// Evaluate runs both tasks synchronously and computes the optimization delta.
func (e *Engine) Evaluate(ctx context.Context, taskName string, baseline, shadow Task) error {
	log.Debug().Str("task", taskName).Msg("Starting baseline vs shadow execution...")

	// 1. Run Baseline (Original Prompt/Model)
	baseMetrics, err := baseline(ctx)
	if err != nil {
		return fmt.Errorf("baseline task failed: %w", err)
	}

	// 2. Run Shadow (New Prompt/Model)
	shadowMetrics, err := shadow(ctx)
	if err != nil {
		return fmt.Errorf("shadow task failed: %w", err)
	}

	// Evaluation Criteria
	// 1. Output quality guardrail: Output must remain structurally similar.
	// (Heuristic: must be at least 80% of the length of the original. 
	// In v2, this routes to a local LLM-as-a-judge for semantic verification).
	if len(shadowMetrics.Output) < int(float64(len(baseMetrics.Output))*0.8) {
		log.Debug().Msg("Shadow output was too degraded. Discarding experiment.")
		return nil
	}

	// 2. Calculate savings
	costSaved := baseMetrics.Cost - shadowMetrics.Cost
	costSavedPct := 0.0
	if baseMetrics.Cost > 0 {
		costSavedPct = (costSaved / baseMetrics.Cost) * 100
	}

	latencySaved := baseMetrics.Latency - shadowMetrics.Latency

	var reasons []string
	if costSavedPct >= 10.0 { // At least 10% cheaper
		reasons = append(reasons, fmt.Sprintf("save %.0f%% API costs", costSavedPct))
	}
	if latencySaved > (baseMetrics.Latency / 5) { // At least 20% faster
		reasons = append(reasons, fmt.Sprintf("run %v faster", latencySaved))
	}

	// If no significant improvement, quietly discard the shadow run.
	if len(reasons) == 0 {
		log.Debug().Msg("Shadow run did not yield significant improvements.")
		return nil
	}

	// Significant improvement found! Trigger HITL Gate.
	reasonStr := strings.Join(reasons, " and ")
	diffPreview := generateDiffPreview(baseMetrics.Output, shadowMetrics.Output)

	prompt := fmt.Sprintf(
		"Hey, I found a way to %s on your '%s'.\n\n"+
			"[Baseline] Cost: $%.4f | Latency: %v\n"+
			"[Shadow]   Cost: $%.4f | Latency: %v\n\n"+
			"Output Diff Preview:\n%s\n\n"+
			"Approve upgrade? [Y/N]",
		reasonStr, taskName,
		baseMetrics.Cost, baseMetrics.Latency,
		shadowMetrics.Cost, shadowMetrics.Latency,
		diffPreview,
	)

	// Hit the Risk Gate (LOW risk because we are just modifying internal config)
	approved := e.hitl.AskPermission(ctx, "LOW", prompt)
	if approved {
		log.Info().Str("task", taskName).Msg("User approved Shadow Mode upgrade! Applying new config...")
		// TODO: Emit an event to the multi-agent bus to persist the new config
	} else {
		log.Info().Str("task", taskName).Msg("User rejected Shadow Mode upgrade.")
	}

	return nil
}

// generateDiffPreview creates a truncated visual diff for the user to review.
func generateDiffPreview(base, shadow string) string {
	truncate := func(s string) string {
		if len(s) > 50 {
			return strings.ReplaceAll(s[:47], "\n", " ") + "..."
		}
		return strings.ReplaceAll(s, "\n", " ")
	}
	return fmt.Sprintf("- %s\n+ %s", truncate(base), truncate(shadow))
}
