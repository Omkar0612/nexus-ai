package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Omkar0612/nexus-ai/internal/types"
	"github.com/rs/zerolog/log"
)

// Provider is a registered LLM backend
type Provider struct {
	Name    string
	BaseURL string
	APIKey  string
	Model   string
	Healthy bool
}

// Router selects the best available LLM provider and falls back automatically
type Router struct {
	primary   *Provider
	fallbacks []*Provider
	client    *http.Client
}

// New creates a new LLM router from config
func New(cfg types.LLMConfig) *Router {
	primary := providerFromConfig(cfg)
	return &Router{
		primary:   primary,
		fallbacks: []*Provider{},
		client:    &http.Client{Timeout: time.Duration(cfg.TimeoutSec) * time.Second},
	}
}

// AddFallback registers a fallback provider
func (r *Router) AddFallback(p *Provider) {
	r.fallbacks = append(r.fallbacks, p)
}

// Complete sends a completion request, falling back on error
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
			log.Warn().Str("provider", p.Name).Err(err).Msg("provider failed, trying fallback")
			p.Healthy = false
			lastErr = err
			continue
		}
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

// callProvider sends a chat completion request to a single provider
func (r *Router) callProvider(ctx context.Context, p *Provider, system, user string) (string, int, int, error) {
	body := map[string]interface{}{
		"model": p.Model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"max_tokens": 2048,
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", 0, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", 0, 0, fmt.Errorf("provider %s HTTP %d: %s", p.Name, resp.StatusCode, string(b))
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
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", 0, 0, err
	}
	if len(res.Choices) == 0 {
		return "", 0, 0, fmt.Errorf("empty response from %s", p.Name)
	}
	return strings.TrimSpace(res.Choices[0].Message.Content),
		res.Usage.PromptTokens, res.Usage.CompletionTokens, nil
}

// HealthCheck pings all providers and marks them healthy/unhealthy
func (r *Router) HealthCheck(ctx context.Context) {
	providers := append([]*Provider{r.primary}, r.fallbacks...)
	for _, p := range providers {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.BaseURL+"/models", nil)
		if err != nil {
			p.Healthy = false
			continue
		}
		if p.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+p.APIKey)
		}
		resp, err := r.client.Do(req)
		if err != nil || resp.StatusCode >= 500 {
			p.Healthy = false
			log.Warn().Str("provider", p.Name).Msg("provider unhealthy")
		} else {
			p.Healthy = true
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
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
	return &Provider{Name: cfg.Provider, BaseURL: baseURL, APIKey: cfg.APIKey, Model: model, Healthy: true}
}
