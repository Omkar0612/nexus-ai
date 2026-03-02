package routing

import (
	"context"
	"time"
)

// Provider represents an LLM API provider (e.g., Groq, Gemini, OpenRouter).
type Provider string

const (
	ProviderGroq       Provider = "GROQ"
	ProviderGemini     Provider = "GEMINI"
	ProviderOpenRouter Provider = "OPENROUTER"
	ProviderOllama     Provider = "OLLAMA"
)

// MarketQuote represents the current "price" and status of a specific LLM model.
type MarketQuote struct {
	Provider       Provider
	Model          string
	CostPer1MT     float64       // Cost per 1 million input tokens (in USD)
	CurrentLatency time.Duration // Real-time ping latency
	IsRateLimited  bool          // True if the provider returned 429 recently
}

// QuoteSource is an interface for a client that can fetch real-time market quotes.
type QuoteSource interface {
	FetchQuote(ctx context.Context) (*MarketQuote, error)
	Name() Provider
}
