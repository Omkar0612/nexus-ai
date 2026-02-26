// Package imagegen provides AI image generation for NEXUS v1.7.
// Backends:
//   - Stable Diffusion via Automatic1111/ComfyUI (local, free, private)
//   - Together AI FLUX.1-schnell (free $25 credits)
//   - Replicate SDXL (limited free runs)
package imagegen

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Backend selects the generation provider.
type Backend string

const (
	BackendSD        Backend = "stablediffusion" // local Automatic1111 or ComfyUI
	BackendTogether  Backend = "together"         // free $25 credits — FLUX.1-schnell
	BackendReplicate Backend = "replicate"         // free limited runs — SDXL
)

// Request describes an image generation request.
type Request struct {
	Prompt         string
	NegativePrompt string
	Width          int    // default 512
	Height         int    // default 512
	Steps          int    // default 20
	OutputPath     string // save PNG to file; empty = base64 only
}

// Result holds the generation output.
type Result struct {
	Base64  string
	Path    string
	Backend Backend
	Latency time.Duration
}

// Agent is the image generation agent.
type Agent struct {
	backend  Backend
	sdURL   string // Automatic1111 base URL
	apiKey  string // Together or Replicate
	model   string
	client  *http.Client
}

// Option configures the agent.
type Option func(*Agent)

// WithStableDiffusion uses a local Automatic1111/ComfyUI server.
func WithStableDiffusion(baseURL string) Option {
	return func(a *Agent) {
		a.backend = BackendSD
		a.sdURL = baseURL
	}
}

// WithTogether uses the Together AI free-tier image API (FLUX.1-schnell).
func WithTogether(apiKey, model string) Option {
	return func(a *Agent) {
		a.backend = BackendTogether
		a.apiKey = apiKey
		a.model = model
	}
}

// WithReplicate uses the Replicate API (SDXL, free limited runs).
func WithReplicate(apiKey string) Option {
	return func(a *Agent) {
		a.backend = BackendReplicate
		a.apiKey = apiKey
	}
}

// New creates an image generation agent.
// Defaults to local Stable Diffusion at http://127.0.0.1:7860.
func New(opts ...Option) *Agent {
	a := &Agent{
		backend: BackendSD,
		sdURL:  "http://127.0.0.1:7860",
		model:  "black-forest-labs/FLUX.1-schnell-Free",
		client: &http.Client{Timeout: 120 * time.Second},
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

// Generate creates an image from the request.
func (a *Agent) Generate(ctx context.Context, req Request) (*Result, error) {
	if req.Width == 0 {
		req.Width = 512
	}
	if req.Height == 0 {
		req.Height = 512
	}
	if req.Steps == 0 {
		req.Steps = 20
	}
	if req.OutputPath == "" {
		req.OutputPath = filepath.Join(os.TempDir(),
			fmt.Sprintf("nexus-img-%d.png", time.Now().UnixNano()))
	}

	var result *Result
	var err error
	switch a.backend {
	case BackendSD:
		result, err = a.generateSD(ctx, req)
	case BackendTogether:
		result, err = a.generateTogether(ctx, req)
	case BackendReplicate:
		result, err = a.generateReplicate(ctx, req)
	default:
		return nil, fmt.Errorf("imagegen: unsupported backend: %s", a.backend)
	}
	if err != nil {
		return nil, err
	}
	// Persist to disk if base64 was returned
	if result.Base64 != "" && result.Path == "" {
		if err := saveBase64(result.Base64, req.OutputPath); err != nil {
			return nil, fmt.Errorf("imagegen: save: %w", err)
		}
		result.Path = req.OutputPath
	}
	return result, nil
}

// --- Stable Diffusion (Automatic1111) ---

type sdRequest struct {
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	Steps          int    `json:"steps"`
}

// SDResponse is exported for use in tests.
type SDResponse struct {
	Images []string `json:"images"`
}

func (a *Agent) generateSD(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	body, err := json.Marshal(sdRequest{
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Width:          req.Width,
		Height:         req.Height,
		Steps:          req.Steps,
	})
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.sdURL+"/sdapi/v1/txt2img", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("imagegen[sd]: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("imagegen[sd]: status %d: %s", resp.StatusCode, raw)
	}
	var sdResp SDResponse
	if err := json.NewDecoder(resp.Body).Decode(&sdResp); err != nil {
		return nil, fmt.Errorf("imagegen[sd]: decode: %w", err)
	}
	if len(sdResp.Images) == 0 {
		return nil, fmt.Errorf("imagegen[sd]: no images returned")
	}
	return &Result{Base64: sdResp.Images[0], Backend: BackendSD, Latency: time.Since(start)}, nil
}

// --- Together AI (FLUX.1-schnell, free credits) ---

type togetherImgRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Steps  int    `json:"steps"`
	N      int    `json:"n"`
}

type togetherImgResponse struct {
	Data []struct {
		B64JSON string `json:"b64_json"`
	} `json:"data"`
}

func (a *Agent) generateTogether(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	body, err := json.Marshal(togetherImgRequest{
		Model:  a.model,
		Prompt: req.Prompt,
		Width:  req.Width,
		Height: req.Height,
		Steps:  req.Steps,
		N:      1,
	})
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.together.xyz/v1/images/generations", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("imagegen[together]: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("imagegen[together]: status %d: %s", resp.StatusCode, raw)
	}
	var tResp togetherImgResponse
	if err := json.NewDecoder(resp.Body).Decode(&tResp); err != nil {
		return nil, fmt.Errorf("imagegen[together]: decode: %w", err)
	}
	if len(tResp.Data) == 0 {
		return nil, fmt.Errorf("imagegen[together]: no images returned")
	}
	return &Result{Base64: tResp.Data[0].B64JSON, Backend: BackendTogether, Latency: time.Since(start)}, nil
}

// --- Replicate (SDXL, free limited runs) ---

type replicateInput struct {
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
	NumSteps       int    `json:"num_inference_steps"`
}

type replicatePrediction struct {
	ID     string   `json:"id"`
	Output []string `json:"output"`
	Status string   `json:"status"`
	Error  string   `json:"error"`
}

func (a *Agent) generateReplicate(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	body, err := json.Marshal(map[string]interface{}{
		"version": "39ed52f2a78e934b3ba6e2a89f5b1c712de7dfea535525255b1aa35c5565e08b",
		"input": replicateInput{
			Prompt:         req.Prompt,
			NegativePrompt: req.NegativePrompt,
			Width:          req.Width,
			Height:         req.Height,
			NumSteps:       req.Steps,
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
		return nil, fmt.Errorf("imagegen[replicate]: %w", err)
	}
	defer resp.Body.Close()
	var pred replicatePrediction
	if err := json.NewDecoder(resp.Body).Decode(&pred); err != nil {
		return nil, fmt.Errorf("imagegen[replicate]: decode: %w", err)
	}
	if pred.Error != "" {
		return nil, fmt.Errorf("imagegen[replicate]: %s", pred.Error)
	}
	// Replicate returns a URL, not base64; caller can download separately
	path := ""
	if len(pred.Output) > 0 {
		path = pred.Output[0]
	}
	return &Result{Path: path, Backend: BackendReplicate, Latency: time.Since(start)}, nil
}

func saveBase64(b64, path string) error {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return fmt.Errorf("imagegen: base64 decode: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("imagegen: mkdir: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
