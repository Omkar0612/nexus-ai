package routing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Market operates like a stock exchange for LLM tokens. It continuously polls
// providers and maintains an order book of the cheapest/fastest models.
type Market struct {
	mu      sync.RWMutex
	sources []QuoteSource
	quotes  map[Provider]*MarketQuote
	ticker  *time.Ticker
	quit    chan struct{}
}

// NewMarket initializes the Token Stock Market.
func NewMarket(interval time.Duration) *Market {
	return &Market{
		sources: make([]QuoteSource, 0),
		quotes:  make(map[Provider]*MarketQuote),
		ticker:  time.NewTicker(interval),
		quit:    make(chan struct{}),
	}
}

// RegisterSource adds a new LLM provider to the market exchange.
func (m *Market) RegisterSource(src QuoteSource) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sources = append(m.sources, src)
}

// Start begins polling providers for real-time latency and rate limits.
func (m *Market) Start(ctx context.Context) {
	log.Info().Msg("📈 Starting Token Stock Market (Dynamic Cost Arbitrage)")
	go func() {
		// Initial fetch
		m.refreshQuotes(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case <-m.quit:
				return
			case <-m.ticker.C:
				m.refreshQuotes(ctx)
			}
		}
	}()
}

// Stop halts the market polling.
func (m *Market) Stop() {
	close(m.quit)
	m.ticker.Stop()
}

func (m *Market) refreshQuotes(ctx context.Context) {
	var wg sync.WaitGroup
	for _, src := range m.sources {
		wg.Add(1)
		go func(s QuoteSource) {
			defer wg.Done()
			
			// Set a strict timeout so a lagging provider doesn't stall the market
			pollCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			quote, err := s.FetchQuote(pollCtx)
			if err != nil {
				log.Debug().Err(err).Str("provider", string(s.Name())).Msg("Failed to fetch market quote")
				m.markRateLimited(s.Name())
				return
			}

			m.mu.Lock()
			m.quotes[s.Name()] = quote
			m.mu.Unlock()
			
			log.Debug().
				Str("provider", string(quote.Provider)).
				Str("model", quote.Model).
				Float64("cost_1mt", quote.CostPer1MT).
				Str("latency", quote.CurrentLatency.String()).
				Msg("Market quote updated")
		}(src)
	}
	wg.Wait()
}

func (m *Market) markRateLimited(p Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if q, exists := m.quotes[p]; exists {
		q.IsRateLimited = true
		q.CurrentLatency = 999 * time.Second // Penalize latency heavily
	}
}

// GetBestQuote evaluates the current order book and returns the optimal provider.
// It prioritizes Free/Ollama, then balances Cost vs Latency.
func (m *Market) GetBestQuote() (*MarketQuote, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.quotes) == 0 {
		return nil, fmt.Errorf("no providers available in the market")
	}

	var best *MarketQuote
	lowestScore := 999999.0 // Lower is better

	for _, quote := range m.quotes {
		if quote.IsRateLimited {
			continue
		}

		// Ollama (Local) is always preferred if running fast enough, as cost is 0.00
		if quote.Provider == ProviderOllama && quote.CurrentLatency < 5*time.Second {
			log.Info().Msg("🏎️ Routing to local Ollama (Cost: $0.00)")
			return quote, nil
		}

		// Arbitrage Formula: (Cost * Weight) + (Latency in Seconds * Weight)
		// This mathematically balances a cheap but slow API vs a fast but expensive API.
		latencySec := quote.CurrentLatency.Seconds()
		score := (quote.CostPer1MT * 100) + (latencySec * 10)

		if score < lowestScore {
			lowestScore = score
			best = quote
		}
	}

	if best == nil {
		return nil, fmt.Errorf("all providers are currently rate limited")
	}

	return best, nil
}
