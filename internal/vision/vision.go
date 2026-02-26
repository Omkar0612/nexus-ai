// Package vision provides image analysis via local LLaVA/Moondream (Ollama)
// or a free remote fallback (Together AI free tier).
// Zero cost: Ollama runs models locally; Together AI gives $25 free credits.
package vision

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Backend selects the inference provider.
type Backend string

const (
	BackendOllama   Backend = "ollama"    // local, free, private
	BackendTogether Backend = "together"  // free $25 credits
	BackendGroq     Backend = "groq"      // groq vision (llava)
)

// AnalysisResult holds the vision model's response.
type AnalysisResult struct {
	Description string
	Model       string
	Backend     Backend
	Latency     time.Duration
}

// Agent is the vision agent.
type Agent struct {
	backend    Backend
	model      string
	ollamaURL  string
	apiKey     string
	httpClient *http.Client
}

// Option configures the agent.
type Option func(*Agent)

func WithOllama(baseURL, model string) Option {
	return func(a *Agent) {
		a.backend = BackendOllama
		a.ollamaURL = strings.TrimRight(baseURL, "/")
		a.model = model
	}
}

func WithTogether(apiKey, model string) Option {
	return func(a *Agent) {
		a.backend = BackendTogether
		a.apiKey = apiKey
		a.model = model
	}
}

// New creates a vision agent. Defaults to local Ollama + llava model.
func New(opts ...Option) *Agent {
	a := &Agent{
		backend:    BackendOllama,
		model:      "llava",
		ollamaURL:  "http://localhost:11434",
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

// AnalyseFile reads an image from disk and analyses it with the given prompt.
func (a *Agent) AnalyseFile(ctx context.Context, imagePath, prompt string) (*AnalysisResult, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("vision: read file: %w", err)
	}
	return a.AnalyseBytes(ctx, data, prompt)
}

// AnalyseBytes analyses raw image bytes with the given prompt.
func (a *Agent) AnalyseBytes(ctx context.Context, imageData []byte, prompt string) (*AnalysisResult, error) {
	switch a.backend {
	case BackendOllama:
		return a.analyseOllama(ctx, imageData, prompt)
	case BackendTogether:
		return a.analyseTogether(ctx, imageData, prompt)
	default:
		return nil, fmt.Errorf("vision: unsupported backend: %s", a.backend)
	}
}

// --- Ollama backend ---

type ollamaRequest struct {
	Model  string   `json:"model"`
	Prompt string   `json:"prompt"`
	Images []string `json:"images"` // base64-encoded
	Stream bool     `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Model    string `json:"model"`
}

func (a *Agent) analyseOllama(ctx context.Context, imageData []byte, prompt string) (*AnalysisResult, error) {
	start := time.Now()
	b64 := base64.StdEncoding.EncodeToString(imageData)
	reqBody := ollamaRequest{
		Model:  a.model,
		Prompt: prompt,
		Images: []string{b64},
		Stream: false,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.ollamaURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vision[ollama]: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vision[ollama]: status %d", resp.StatusCode)
	}
	var result ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("vision[ollama]: decode: %w", err)
	}
	return &AnalysisResult{
		Description: result.Response,
		Model:       a.model,
		Backend:     BackendOllama,
		Latency:     time.Since(start),
	}, nil
}

// --- Together AI backend ---

type togetherRequest struct {
	Model    string             `json:"model"`
	Messages []togetherMessage  `json:"messages"`
	MaxTokens int               `json:"max_tokens"`
}

type togetherMessage struct {
	Role    string        `json:"role"`
	Content []togetherContent `json:"content"`
}

type togetherContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL *struct {
		URL string `json:"url"`
	} `json:"image_url,omitempty"`
}

type togetherResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Model string `json:"model"`
}

func (a *Agent) analyseTogether(ctx context.Context, imageData []byte, prompt string) (*AnalysisResult, error) {
	start := time.Now()
	b64 := base64.StdEncoding.EncodeToString(imageData)
	dataURL := "data:image/png;base64," + b64
	reqBody := togetherRequest{
		Model:     a.model,
		MaxTokens: 1024,
		Messages: []togetherMessage{
			{
				Role: "user",
				Content: []togetherContent{
					{Type: "image_url", ImageURL: &struct{ URL string `json:"url"` }{URL: dataURL}},
					{Type: "text", Text: prompt},
				},
			},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.together.xyz/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vision[together]: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("vision[together]: status %d: %s", resp.StatusCode, raw)
	}
	var result togetherResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("vision[together]: no choices in response")
	}
	return &AnalysisResult{
		Description: result.Choices[0].Message.Content,
		Model:       a.model,
		Backend:     BackendTogether,
		Latency:     time.Since(start),
	}, nil
}
