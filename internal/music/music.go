// Package music provides AI music generation for NEXUS v1.7.
// Backends:
//   - AudioCraft (local Python bridge) — Meta's free local model
//   - Replicate MusicGen — limited free runs
//   - Stub — writes a valid silent WAV for CI/testing
package music

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Backend selects the music generation provider.
type Backend string

const (
	BackendAudioCraft Backend = "audiocraft" // local Meta AudioCraft
	BackendReplicate  Backend = "replicate"  // Replicate MusicGen
	BackendStub       Backend = "stub"       // silent WAV — CI safe
)

// Request describes a music generation request.
type Request struct {
	Prompt     string        // e.g. "upbeat jazz piano, 120 bpm"
	Duration   time.Duration // e.g. 10 * time.Second
	OutputPath string        // output WAV file path
}

// Result holds the generation output.
type Result struct {
	Path    string
	Backend Backend
	Latency time.Duration
}

// Agent is the music generation agent.
type Agent struct {
	backend Backend
	acURL   string // AudioCraft bridge URL
	apiKey  string // Replicate API key
	client  *http.Client
}

// Option configures the agent.
type Option func(*Agent)

// WithAudioCraft uses a local AudioCraft Python bridge server.
func WithAudioCraft(bridgeURL string) Option {
	return func(a *Agent) {
		a.backend = BackendAudioCraft
		a.acURL = bridgeURL
	}
}

// WithReplicate uses the Replicate API (free tier: limited runs/month).
func WithReplicate(apiKey string) Option {
	return func(a *Agent) {
		a.backend = BackendReplicate
		a.apiKey = apiKey
	}
}

// New creates a music generation agent. Defaults to stub for safe CI operation.
func New(opts ...Option) *Agent {
	a := &Agent{
		backend: BackendStub,
		acURL:  "http://localhost:8765",
		client: &http.Client{Timeout: 120 * time.Second},
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

// Generate creates music from the prompt and saves to OutputPath.
func (a *Agent) Generate(ctx context.Context, req Request) (*Result, error) {
	if req.Prompt == "" {
		return nil, fmt.Errorf("music: prompt must not be empty")
	}
	if req.Duration == 0 {
		req.Duration = 10 * time.Second
	}
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

// --- AudioCraft bridge ---

type acRequest struct {
	Prompt   string  `json:"prompt"`
	Duration float64 `json:"duration"`
}

func (a *Agent) generateAudioCraft(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	body, err := json.Marshal(acRequest{
		Prompt:   req.Prompt,
		Duration: req.Duration.Seconds(),
	})
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.acURL+"/generate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("music[audiocraft]: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, fmt.Errorf("music[audiocraft]: status %d: %s", resp.StatusCode, raw)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("music[audiocraft]: read: %w", err)
	}
	if err := os.WriteFile(req.OutputPath, data, 0o644); err != nil {
		return nil, fmt.Errorf("music[audiocraft]: write: %w", err)
	}
	return &Result{Path: req.OutputPath, Backend: BackendAudioCraft, Latency: time.Since(start)}, nil
}

// --- Replicate MusicGen ---

type replicateMusicInput struct {
	Prompt        string  `json:"prompt"`
	Duration      int     `json:"duration"`
	OutputFormat  string  `json:"output_format"`
	Continuation  bool    `json:"continuation"`
}

type replicatePrediction struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	Output []string `json:"output"`
	Error  string   `json:"error"`
}

func (a *Agent) generateReplicate(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	body, err := json.Marshal(map[string]interface{}{
		"version": "671ac645ce5e552cc63a54a2bbff63fcf798043055d2dac5fc9e36a837eedcfb",
		"input": replicateMusicInput{
			Prompt:       req.Prompt,
			Duration:     int(req.Duration.Seconds()),
			OutputFormat: "wav",
			Continuation: false,
		},
	})
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.replicate.com/v1/predictions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Token "+a.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("music[replicate]: %w", err)
	}
	defer resp.Body.Close()
	var pred replicatePrediction
	if err := json.NewDecoder(resp.Body).Decode(&pred); err != nil {
		return nil, fmt.Errorf("music[replicate]: decode: %w", err)
	}
	if pred.Error != "" {
		return nil, fmt.Errorf("music[replicate]: %s", pred.Error)
	}
	// Replicate returns a URL; output path is empty until user downloads
	outputURL := ""
	if len(pred.Output) > 0 {
		outputURL = pred.Output[0]
	}
	return &Result{Path: outputURL, Backend: BackendReplicate, Latency: time.Since(start)}, nil
}

// --- Stub (valid silent WAV, CI-safe) ---

// silentWAV is a minimal valid 44-byte WAV file with 0 data samples.
var silentWAV = []byte{
	// RIFF header
	0x52, 0x49, 0x46, 0x46, // "RIFF"
	0x24, 0x00, 0x00, 0x00, // chunk size = 36 (file size - 8)
	0x57, 0x41, 0x56, 0x45, // "WAVE"
	// fmt sub-chunk
	0x66, 0x6D, 0x74, 0x20, // "fmt "
	0x10, 0x00, 0x00, 0x00, // sub-chunk size = 16
	0x01, 0x00, // audio format = PCM
	0x01, 0x00, // num channels = 1
	0x44, 0xAC, 0x00, 0x00, // sample rate = 44100
	0x88, 0x58, 0x01, 0x00, // byte rate = 88200
	0x02, 0x00, // block align = 2
	0x10, 0x00, // bits per sample = 16
	// data sub-chunk
	0x64, 0x61, 0x74, 0x61, // "data"
	0x00, 0x00, 0x00, 0x00, // data size = 0
}

func (a *Agent) generateStub(req Request) (*Result, error) {
	if err := os.WriteFile(req.OutputPath, silentWAV, 0o644); err != nil {
		return nil, fmt.Errorf("music[stub]: write: %w", err)
	}
	return &Result{Path: req.OutputPath, Backend: BackendStub, Latency: 0}, nil
}
