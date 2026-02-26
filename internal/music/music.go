// Package music provides AI music generation for NEXUS v1.7.
// Backends:
//   - AudioCraft (local via Python bridge) — Meta's free local model
//   - Replicate API (MusicGen) — limited free runs
//   - Stub mode — returns silence WAV for testing
package music

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Backend selects the music generation provider.
type Backend string

const (
	BackendAudioCraft Backend = "audiocraft" // local Meta AudioCraft
	BackendReplicate  Backend = "replicate"  // free limited runs
	BackendStub       Backend = "stub"       // silent WAV for tests
)

// Request describes a music generation request.
type Request struct {
	Prompt     string        // e.g. "upbeat jazz piano, 120 bpm"
	Duration   time.Duration // e.g. 10 * time.Second
	OutputPath string        // output WAV/MP3 file path
}

// Result holds the generation output.
type Result struct {
	Path    string
	Backend Backend
	Latency time.Duration
}

// Agent is the music generation agent.
type Agent struct {
	backend    Backend
	acURL      string // AudioCraft bridge URL
	apiKey     string // Replicate
	client     *http.Client
}

// Option configures the agent.
type Option func(*Agent)

// WithAudioCraft uses a local AudioCraft Python bridge server.
func WithAudioCraft(bridgeURL string) Option {
	return func(a *Agent) { a.backend = BackendAudioCraft; a.acURL = bridgeURL }
}

// WithReplicate uses the Replicate API (free tier: limited runs/month).
func WithReplicate(apiKey string) Option {
	return func(a *Agent) { a.backend = BackendReplicate; a.apiKey = apiKey }
}

// New creates a music generation agent. Defaults to stub for safe CI.
func New(opts ...Option) *Agent {
	a := &Agent{
		backend: BackendStub,
		acURL:   "http://localhost:8765",
		client:  &http.Client{Timeout: 120 * time.Second},
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

// Generate creates music from the prompt.
func (a *Agent) Generate(ctx context.Context, req Request) (*Result, error) {
	if req.Duration == 0 { req.Duration = 10 * time.Second }
	if req.OutputPath == "" {
		req.OutputPath = fmt.Sprintf("/tmp/nexus-music-%d.wav", time.Now().UnixNano())
	}
	switch a.backend {
	case BackendAudioCraft:
		return a.generateAudioCraft(ctx, req)
	case BackendReplicate:
		return a.generateReplicate(ctx, req)
	case BackendStub:
		return a.generateStub(req)
	default:
		return nil, fmt.Errorf("music: unsupported backend: %s", a.backend)
	}
}

// --- AudioCraft backend ---

type acRequest struct {
	Prompt   string  `json:"prompt"`
	Duration float64 `json:"duration"`
}

func (a *Agent) generateAudioCraft(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	body, _ := json.Marshal(acRequest{Prompt: req.Prompt, Duration: req.Duration.Seconds()})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.acURL+"/generate", strings.NewReader(string(body)))
	if err != nil { return nil, err }
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(httpReq)
	if err != nil { return nil, fmt.Errorf("music[audiocraft]: %w", err) }
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil { return nil, err }
	if err := os.WriteFile(req.OutputPath, data, 0644); err != nil { return nil, err }
	return &Result{Path: req.OutputPath, Backend: BackendAudioCraft, Latency: time.Since(start)}, nil
}

// --- Replicate backend ---

func (a *Agent) generateReplicate(ctx context.Context, req Request) (*Result, error) {
	// Replicate API: POST /v1/predictions with model facebook/musicgen
	start := time.Now()
	body, _ := json.Marshal(map[string]any{
		"version": "671ac645ce5e552cc63a54a2bbff63fcf798043055d2dac5fc9e36a837eedcfb",
		"input": map[string]any{
			"prompt":             req.Prompt,
			"duration":           int(req.Duration.Seconds()),
			"output_format":      "wav",
			"continuation":       false,
		},
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.replicate.com/v1/predictions", strings.NewReader(string(body)))
	if err != nil { return nil, err }
	httpReq.Header.Set("Authorization", "Token "+a.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(httpReq)
	if err != nil { return nil, fmt.Errorf("music[replicate]: %w", err) }
	defer resp.Body.Close()
	var prediction map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil { return nil, err }
	return &Result{Path: "", Backend: BackendReplicate, Latency: time.Since(start)}, nil
}

// --- Stub backend (silent WAV, CI safe) ---

var silentWAV = []byte{
	0x52, 0x49, 0x46, 0x46, 0x24, 0x00, 0x00, 0x00,
	0x57, 0x41, 0x56, 0x45, 0x66, 0x6D, 0x74, 0x20,
	0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
	0x44, 0xAC, 0x00, 0x00, 0x88, 0x58, 0x01, 0x00,
	0x02, 0x00, 0x10, 0x00, 0x64, 0x61, 0x74, 0x61,
	0x00, 0x00, 0x00, 0x00,
}

func (a *Agent) generateStub(req Request) (*Result, error) {
	if err := os.WriteFile(req.OutputPath, silentWAV, 0644); err != nil {
		return nil, err
	}
	return &Result{Path: req.OutputPath, Backend: BackendStub, Latency: 0}, nil
}
