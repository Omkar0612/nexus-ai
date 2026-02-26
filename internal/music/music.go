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
	BackendAudioCraft Backend = "audiocraft"
	BackendReplicate  Backend = "replicate"
	BackendStub       Backend = "stub"
)

// Request describes a music generation request.
type Request struct {
	Prompt     string
	Duration   time.Duration
	OutputPath string
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
	acURL   string
	apiKey  string
	client  *http.Client
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

// New creates a music generation agent. Defaults to stub for safe CI operation.
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

// doJSON is a shared helper: marshal body → POST → check status → decode into out.
func (a *Agent) doJSON(ctx context.Context, url string, body, out interface{}, authHeader string) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, raw)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// --- AudioCraft bridge ---

type acRequest struct {
	Prompt   string  `json:"prompt"`
	Duration float64 `json:"duration"`
}

type acResponse struct {
	Data string `json:"data"` // base64-encoded WAV
	Path string `json:"path"` // or direct path if bridge writes to disk
}

func (a *Agent) generateAudioCraft(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	var acResp acResponse
	if err := a.doJSON(ctx, a.acURL+"/generate",
		acRequest{Prompt: req.Prompt, Duration: req.Duration.Seconds()},
		&acResp, ""); err != nil {
		return nil, fmt.Errorf("music[audiocraft]: %w", err)
	}
	// Bridge may return either a path or base64 data.
	outPath := acResp.Path
	if outPath == "" {
		outPath = req.OutputPath
		if acResp.Data != "" {
			if err := os.WriteFile(outPath, []byte(acResp.Data), 0o644); err != nil {
				return nil, fmt.Errorf("music[audiocraft]: write: %w", err)
			}
		}
	}
	return &Result{Path: outPath, Backend: BackendAudioCraft, Latency: time.Since(start)}, nil
}

// --- Replicate MusicGen ---

type replicateMusicInput struct {
	Prompt       string `json:"prompt"`
	Duration     int    `json:"duration"`
	OutputFormat string `json:"output_format"`
	Continuation bool   `json:"continuation"`
}

type replicatePrediction struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	Output []string `json:"output"`
	Error  string   `json:"error"`
}

func (a *Agent) generateReplicate(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	var pred replicatePrediction
	if err := a.doJSON(ctx, "https://api.replicate.com/v1/predictions",
		map[string]interface{}{
			"version": "671ac645ce5e552cc63a54a2bbff63fcf798043055d2dac5fc9e36a837eedcfb",
			"input": replicateMusicInput{
				Prompt:       req.Prompt,
				Duration:     int(req.Duration.Seconds()),
				OutputFormat: "wav",
			},
		}, &pred, "Token "+a.apiKey); err != nil {
		return nil, fmt.Errorf("music[replicate]: %w", err)
	}
	if pred.Error != "" {
		return nil, fmt.Errorf("music[replicate]: %s", pred.Error)
	}
	outURL := ""
	if len(pred.Output) > 0 {
		outURL = pred.Output[0]
	}
	return &Result{Path: outURL, Backend: BackendReplicate, Latency: time.Since(start)}, nil
}

// silentWAV is a minimal valid 44-byte WAV file with 0 data samples.
var silentWAV = []byte{
	0x52, 0x49, 0x46, 0x46, 0x24, 0x00, 0x00, 0x00,
	0x57, 0x41, 0x56, 0x45, 0x66, 0x6D, 0x74, 0x20,
	0x10, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
	0x44, 0xAC, 0x00, 0x00, 0x88, 0x58, 0x01, 0x00,
	0x02, 0x00, 0x10, 0x00, 0x64, 0x61, 0x74, 0x61,
	0x00, 0x00, 0x00, 0x00,
}

func (a *Agent) generateStub(req Request) (*Result, error) {
	if err := os.WriteFile(req.OutputPath, silentWAV, 0o644); err != nil {
		return nil, fmt.Errorf("music[stub]: write: %w", err)
	}
	return &Result{Path: req.OutputPath, Backend: BackendStub}, nil
}
