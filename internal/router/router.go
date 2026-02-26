// Package router provides an LLM provider router with automatic fallback
// and a simple circuit-breaker (3 consecutive failures → mark unhealthy).
package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Omkar0612/nexus-ai/internal/types"
	"github.com/rs/zerolog/log"
)

// circuitThreshold is the number of consecutive errors before a provider is
// marked unhealthy and skipped by the router.
const circuitThreshold = 3

// sharedTransport is a tuned http.Transport reused by all router instances.
var sharedTransport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 20,
	IdleConnTimeout:     90 * time.Second,
	TLSHandshakeTimeout: 10 * time.Second,
	DisableCompression:  false,
}

// SecretString wraps an API key and masks it in all fmt/log output.
type SecretString struct{ v string }

// NewSecret wraps a plaintext value as a SecretString.
func NewSecret(s string) SecretString { return SecretString{v: s} }

// Value returns the raw key (only call when building HTTP headers).
func (s SecretString) Value() string { return s.v }

// String implements fmt.Stringer — always returns "[REDACTED]".
func (s SecretString) String() string { return "[REDACTED]" }

// GoString prevents leakage via %#v.
func (s SecretString) GoString() string { return "router.SecretString([REDACTED])" }

// Provider is a registered LLM backend.
type Provider struct {
	Name     string
	BaseURL  string
	APIKey   SecretString // masked in logs and fmt output
	Model    string
	Healthy  bool
	failures atomic.Int32 // consecutive failure counter — circuit breaker
}

// recordFailure increments the failure counter and marks unhealthy at threshold.
func (p *Provider) recordFailure() {
	if p.failures.Add(1) >= circuitThreshold {
		p.Healthy = false
	}
}

// recordSuccess resets the circuit breaker.
func (p *Provider) recordSuccess() {
	p.failures.Store(0)
	p.Healthy = true
}

// Router selects the best available LLM provider with automatic fallback.
type Router struct {
	primary   *Provider
	fallbacks []*Provider
	client    *http.Client
}

// New creates a new LLM router from config.
func New(cfg types.LLMConfig) *Router {
	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	return &Router{
		primary:   providerFromConfig(cfg),
		fallbacks: []*Provider{},
		client: &http.Client{
			Timeout:   timeout,
			Transport: sharedTransport,
		},
	}
}

// AddFallback registers a fallback provider.
func (r *Router) AddFallback(p *Provider) {
	r.fallbacks = append(r.fallbacks, p)
}

// Complete sends a completion request, falling back on error.
func (r *Router) Complete(ctx context.Context, systemPrompt, userMsg string) (*types.AgentResult, error) {
	start := time.Now()
	providers := append([]*Provider{r.primary}, r.fallbacks...)
	var lastErr error
	for _, p := range providers {
		if !p.Healthy {
			continue
		}
		content, tokIn, tokOut, err := r.callProvider(ctx, p, systemPrompt, userMsg)
		if err != nil {
			// Log provider name only — not the APIKey.
			log.Warn().Str("provider", p.Name).Err(err).Msg("provider failed, trying fallback")
			p.recordFailure()
			lastErr = err
			continue
		}
		p.recordSuccess()
		return &types.AgentResult{
			Content:   content,
			Agent:     "router",
			Model:     p.Name + "/" + p.Model,
			LatencyMs: time.Since(start).Milliseconds(),
			TokensIn:  tokIn,
			TokensOut: tokOut,
		}, nil
	}
	return nil, fmt.Errorf("all providers failed: %w", lastErr)
}

// callProvider sends a chat completion request to a single provider.
func (r *Router) callProvider(ctx context.Context, p *Provider, system, user string) (string, int, int, error) {
	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	body := struct {
		Model     string    `json:"model"`
		Messages  []message `json:"messages"`
		MaxTokens int       `json:"max_tokens"`
	}{
		Model: p.Model,
		Messages: []message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		MaxTokens: 2048,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return "", 0, 0, fmt.Errorf("router: encode: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		p.BaseURL+"/chat/completions", &buf)
	if err != nil {
		return "", 0, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.APIKey.Value() != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey.Value())
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		// Drain body to allow connection reuse; log internally but don't propagate raw body.
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		log.Debug().Str("provider", p.Name).Int("status", resp.StatusCode).Bytes("body", b).Msg("provider error response")
		return "", 0, 0, fmt.Errorf("provider %s HTTP %d", p.Name, resp.StatusCode)
	}
	var res struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4*1024*1024)).Decode(&res); err != nil {
		return "", 0, 0, fmt.Errorf("router: decode: %w", err)
	}
	if len(res.Choices) == 0 {
		return "", 0, 0, fmt.Errorf("empty response from %s", p.Name)
	}
	return strings.TrimSpace(res.Choices[0].Message.Content),
		res.Usage.PromptTokens, res.Usage.CompletionTokens, nil
}

// HealthCheck pings all providers in parallel and marks them healthy/unhealthy.
func (r *Router) HealthCheck(ctx context.Context) {
	providers := append([]*Provider{r.primary}, r.fallbacks...)
	var wg sync.WaitGroup
	for _, p := range providers {
		p := p
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.BaseURL+"/models", nil)
			if err != nil {
				p.Healthy = false
				return
			}
			if p.APIKey.Value() != "" {
				req.Header.Set("Authorization", "Bearer "+p.APIKey.Value())
			}
			resp, err := r.client.Do(req)
			if err != nil || resp.StatusCode >= 500 {
				p.Healthy = false
				log.Warn().Str("provider", p.Name).Msg("provider unhealthy")
			} else {
				p.recordSuccess()
				log.Debug().Str("provider", p.Name).Msg("provider healthy")
			}
			if resp != nil {
				resp.Body.Close()
			}
		}()
	}
	wg.Wait()
}

func providerFromConfig(cfg types.LLMConfig) *Provider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		switch strings.ToLower(cfg.Provider) {
		case "groq":
			baseURL = "https://api.groq.com/openai/v1"
		case "anthropic":
			baseURL = "https://api.anthropic.com/v1"
		case "openai":
			baseURL = "https://api.openai.com/v1"
		case "together":
			baseURL = "https://api.together.xyz/v1"
		case "ollama":
			baseURL = "http://localhost:11434/v1"
		default:
			baseURL = "http://localhost:11434/v1"
		}
	}
	model := cfg.Model
	if model == "" {
		model = "llama3.2"
	}
	return &Provider{
		Name:    cfg.Provider,
		BaseURL: baseURL,
		APIKey:  NewSecret(cfg.APIKey),
		Model:   model,
		Healthy: true,
	}
}
