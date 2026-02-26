// Package tts provides AI voice synthesis for NEXUS v1.7.
// Backends:
//   - Coqui TTS (local, free, offline) — http://localhost:5002
//   - ElevenLabs free tier — 10,000 chars/month free
//   - System TTS fallback (espeak/say) — zero cost, always available
package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// Backend selects the TTS provider.
type Backend string

const (
	BackendCoqui      Backend = "coqui"      // local, free, offline
	BackendElevenLabs Backend = "elevenlabs" // 10k chars/month free
	BackendSystem     Backend = "system"     // espeak / say — always free
)

// Request is a speech synthesis request.
type Request struct {
	Text       string
	Voice      string // voice ID or name (provider-specific)
	OutputPath string // WAV/MP3 output file path
	Speed      float64 // 1.0 = normal
}

// Result holds synthesis output.
type Result struct {
	Path    string
	Backend Backend
	Latency time.Duration
}

// Agent is the TTS agent.
type Agent struct {
	backend   Backend
	coquiURL  string
	apiKey    string // ElevenLabs
	voiceID   string // ElevenLabs default voice
	client    *http.Client
}

// Option configures the TTS agent.
type Option func(*Agent)

// WithCoqui uses a local Coqui TTS server.
func WithCoqui(baseURL string) Option {
	return func(a *Agent) { a.backend = BackendCoqui; a.coquiURL = baseURL }
}

// WithElevenLabs uses the ElevenLabs API (free tier: 10k chars/month).
func WithElevenLabs(apiKey, voiceID string) Option {
	return func(a *Agent) {
		a.backend = BackendElevenLabs
		a.apiKey = apiKey
		a.voiceID = voiceID
	}
}

// WithSystem uses the OS built-in TTS (espeak on Linux, say on macOS).
func WithSystem() Option {
	return func(a *Agent) { a.backend = BackendSystem }
}

// New creates a TTS agent. Defaults to system TTS (always available).
func New(opts ...Option) *Agent {
	a := &Agent{
		backend:  BackendSystem,
		coquiURL: "http://localhost:5002",
		client:   &http.Client{Timeout: 60 * time.Second},
	}
	for _, o := range opts {
		o(a)
	}
	return a
}

// Speak synthesises text to a file (or plays it directly for system TTS).
func (a *Agent) Speak(ctx context.Context, req Request) (*Result, error) {
	if req.Speed == 0 { req.Speed = 1.0 }
	switch a.backend {
	case BackendCoqui:
		return a.speakCoqui(ctx, req)
	case BackendElevenLabs:
		return a.speakElevenLabs(ctx, req)
	case BackendSystem:
		return a.speakSystem(ctx, req)
	default:
		return nil, fmt.Errorf("tts: unsupported backend: %s", a.backend)
	}
}

// --- Coqui TTS backend ---

func (a *Agent) speakCoqui(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	url := fmt.Sprintf("%s/api/tts?text=%s", a.coquiURL, req.Text)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil { return nil, err }
	resp, err := a.client.Do(httpReq)
	if err != nil { return nil, fmt.Errorf("tts[coqui]: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("tts[coqui]: status %d", resp.StatusCode)
	}
	outPath := req.OutputPath
	if outPath == "" { outPath = tempWAV() }
	data, err := io.ReadAll(resp.Body)
	if err != nil { return nil, err }
	if err := os.WriteFile(outPath, data, 0644); err != nil { return nil, err }
	return &Result{Path: outPath, Backend: BackendCoqui, Latency: time.Since(start)}, nil
}

// --- ElevenLabs backend (free tier) ---

type elevenLabsRequest struct {
	Text          string                 `json:"text"`
	ModelID       string                 `json:"model_id"`
	VoiceSettings map[string]interface{} `json:"voice_settings"`
}

func (a *Agent) speakElevenLabs(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	body, _ := json.Marshal(elevenLabsRequest{
		Text:    req.Text,
		ModelID: "eleven_monolingual_v1",
		VoiceSettings: map[string]interface{}{"stability": 0.5, "similarity_boost": 0.75},
	})
	voiceID := a.voiceID
	if req.Voice != "" { voiceID = req.Voice }
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil { return nil, err }
	httpReq.Header.Set("xi-api-key", a.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(httpReq)
	if err != nil { return nil, fmt.Errorf("tts[elevenlabs]: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tts[elevenlabs]: status %d: %s", resp.StatusCode, raw)
	}
	outPath := req.OutputPath
	if outPath == "" { outPath = tempWAV() }
	data, err := io.ReadAll(resp.Body)
	if err != nil { return nil, err }
	if err := os.WriteFile(outPath, data, 0644); err != nil { return nil, err }
	return &Result{Path: outPath, Backend: BackendElevenLabs, Latency: time.Since(start)}, nil
}

// --- System TTS backend ---

func (a *Agent) speakSystem(_ context.Context, req Request) (*Result, error) {
	start := time.Now()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("say", req.Text)
	case "windows":
		cmd = exec.Command("powershell", "-Command",
			fmt.Sprintf(`Add-Type -AssemblyName System.Speech; (New-Object System.Speech.Synthesis.SpeechSynthesizer).Speak('%s')`, req.Text))
	default:
		cmd = exec.Command("espeak", req.Text)
	}
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tts[system]: %w", err)
	}
	return &Result{Backend: BackendSystem, Latency: time.Since(start)}, nil
}

func tempWAV() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("nexus-tts-%d.wav", time.Now().UnixNano()))
}
