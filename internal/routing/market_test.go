package routing

import (
	"context"
	"testing"
	"time"
)

type mockSource struct {
	quote *MarketQuote
	err   error
}

func (m *mockSource) FetchQuote(ctx context.Context) (*MarketQuote, error) {
	return m.quote, m.err
}

func (m *mockSource) Name() Provider {
	return m.quote.Provider
}

func TestMarket_GetBestQuote_OllamaPreferred(t *testing.T) {
	market := NewMarket(1 * time.Minute)

	// Expensive but fast
	groq := &mockSource{
		quote: &MarketQuote{
			Provider:       ProviderGroq,
			Model:          "llama3",
			CostPer1MT:     0.50,
			CurrentLatency: 200 * time.Millisecond,
		},
	}

	// Free local
	ollama := &mockSource{
		quote: &MarketQuote{
			Provider:       ProviderOllama,
			Model:          "mistral",
			CostPer1MT:     0.00,
			CurrentLatency: 1 * time.Second,
		},
	}

	market.RegisterSource(groq)
	market.RegisterSource(ollama)

	// Manually force refresh for test
	market.refreshQuotes(context.Background())

	best, err := market.GetBestQuote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// It should route to Ollama because cost is $0.00 and latency is acceptable
	if best.Provider != ProviderOllama {
		t.Errorf("expected OLLAMA, got %s", best.Provider)
	}
}

func TestMarket_GetBestQuote_Arbitrage(t *testing.T) {
	market := NewMarket(1 * time.Minute)

	// Fast but expensive ($2.00)
	openRouter := &mockSource{
		quote: &MarketQuote{
			Provider:       ProviderOpenRouter,
			Model:          "gpt-4o",
			CostPer1MT:     2.50,
			CurrentLatency: 300 * time.Millisecond,
		},
	}

	// Cheap ($0.15) and reasonable latency
	gemini := &mockSource{
		quote: &MarketQuote{
			Provider:       ProviderGemini,
			Model:          "flash-2.0",
			CostPer1MT:     0.15,
			CurrentLatency: 800 * time.Millisecond,
		},
	}

	market.RegisterSource(openRouter)
	market.RegisterSource(gemini)
	market.refreshQuotes(context.Background())

	best, err := market.GetBestQuote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The arbitrage formula should prefer Gemini (score ~23 vs OpenRouter score ~253)
	if best.Provider != ProviderGemini {
		t.Errorf("expected GEMINI due to arbitrage, got %s", best.Provider)
	}
}

func TestMarket_GetBestQuote_RateLimitEvasion(t *testing.T) {
	market := NewMarket(1 * time.Minute)

	// Groq is normally cheapest/fastest, but rate limited
	groq := &mockSource{
		quote: &MarketQuote{
			Provider:       ProviderGroq,
			CostPer1MT:     0.10,
			CurrentLatency: 100 * time.Millisecond,
			IsRateLimited:  true,
		},
	}

	gemini := &mockSource{
		quote: &MarketQuote{
			Provider:       ProviderGemini,
			CostPer1MT:     0.15,
			CurrentLatency: 800 * time.Millisecond,
			IsRateLimited:  false,
		},
	}

	market.RegisterSource(groq)
	market.RegisterSource(gemini)
	market.refreshQuotes(context.Background())

	best, err := market.GetBestQuote()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must evade Groq and fall back to Gemini
	if best.Provider != ProviderGemini {
		t.Errorf("expected GEMINI to evade rate limit, got %s", best.Provider)
	}
}
