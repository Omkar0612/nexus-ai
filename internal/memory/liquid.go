package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

// LLMClient represents a generator that can distill large texts into dense concepts.
type LLMClient interface {
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

// LiquidContext Engine monitors episodic memory. When the context window reaches
// a critical threshold (e.g., 8000 tokens), it "dreams" in the background, 
// semantically clustering old episodes and compressing them into dense Concepts.
type LiquidContext struct {
	db             Storage
	llm            LLMClient
	threshold      int // Max tokens before consolidation triggers
	ticker         *time.Ticker
	quit           chan struct{}
}

// NewLiquidContext initializes the memory consolidator.
func NewLiquidContext(db Storage, llm LLMClient, maxTokens int, interval time.Duration) *LiquidContext {
	return &LiquidContext{
		db:        db,
		llm:       llm,
		threshold: maxTokens,
		ticker:    time.NewTicker(interval),
		quit:      make(chan struct{}),
	}
}

// Start begins the background consolidation loop.
func (lc *LiquidContext) Start(ctx context.Context) {
	log.Info().Msg("🧠 Starting Liquid Context (Agentic Memory Consolidation)")
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-lc.quit:
				return
			case <-lc.ticker.C:
				lc.consolidate(ctx)
			}
		}
	}()
}

// Stop safely halts the consolidator.
func (lc *LiquidContext) Stop() {
	close(lc.quit)
	lc.ticker.Stop()
}

func (lc *LiquidContext) consolidate(ctx context.Context) {
	totalTokens, err := lc.db.GetContextWindow()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to check context window size")
		return
	}

	if totalTokens < lc.threshold {
		// Context is still healthy, no need to compress
		return
	}

	log.Info().
		Int("current_tokens", totalTokens).
		Int("threshold", lc.threshold).
		Msg("Context window threshold reached. Initiating semantic memory consolidation...")

	// Grab episodes older than 24 hours (or whatever defines "old" in this DB)
	cutoff := time.Now().Add(-24 * time.Hour)
	oldEpisodes, err := lc.db.GetOldEpisodes(cutoff)
	if err != nil || len(oldEpisodes) == 0 {
		return
	}

	var rawContext string
	var idsToDelete []string
	var tokensSaved int

	for _, ep := range oldEpisodes {
		rawContext += fmt.Sprintf("[%s] %s: %s\n", ep.Timestamp.Format(time.Kitchen), ep.Role, ep.Content)
		idsToDelete = append(idsToDelete, ep.ID)
		tokensSaved += ep.Tokens
	}

	// The Prompt that does the "Liquid Context" compression
	sysPrompt := `You are the NEXUS Memory Consolidator. 
Your job is to read raw, bloated chat logs and extract ONLY the dense, factual, permanent knowledge.
Strip all conversational filler, pleasantries, and redundant text.
Output a highly compressed, factual summary (a "Concept").`

	userPrompt := fmt.Sprintf("Compress the following history:\n%s", rawContext)

	// In a real scenario, we use a cheap, fast model (like Gemini Flash) for this background task
	compressedText, err := lc.llm.Generate(ctx, sysPrompt, userPrompt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate dense concept")
		return
	}

	// Store the new dense concept
	concept := &Concept{
		ID:           fmt.Sprintf("concept_%d", time.Now().Unix()),
		Timestamp:    time.Now(),
		Topic:        "Archived Conversation",
		DenseSummary: compressedText,
		Tokens:       len(compressedText) / 4, // Rough token estimate
	}

	if err := lc.db.StoreConcept(concept); err != nil {
		log.Error().Err(err).Msg("Failed to store consolidated concept")
		return
	}

	// Delete the old, bloated episodes
	if err := lc.db.DeleteEpisodes(idsToDelete); err != nil {
		log.Error().Err(err).Msg("Failed to clear old episodes after consolidation")
		return
	}

	log.Info().
		Int("episodes_compressed", len(oldEpisodes)).
		Int("tokens_freed", tokensSaved - concept.Tokens).
		Msg("✅ Liquid Context successfully consolidated memory.")
}
