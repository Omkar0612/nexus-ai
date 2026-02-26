// Package tts provides AI voice synthesis for NEXUS v1.7.
// Backends:
//   - Coqui TTS (local, free, offline) — http://localhost:5002
//   - ElevenLabs free tier — 10,000 chars/month free
//   - System TTS fallback (espeak / say / PowerShell) — zero cost
package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	Voice      string  // voice ID or name (provider-specific)
	OutputPath string  // WAV/MP3 output file; empty = temp file
	Speed      float64 // 1.0 = normal; only used where supported
}

// Result holds synthesis output.
type Result struct {
	Path    string
	Backend Backend
	Latency time.Duration
}

// Agent is the TTS agent.
type Agent struct {
	backend  Backend
	coquiURL string
	apiKey  string // ElevenLabs
	voiceID string // ElevenLabs default voice
	client  *http.Client
}

// Option configures the TTS agent.
type Option func(*Agent)

// WithCoqui uses a local Coqui TTS server.
func WithCoqui(baseURL string) Option {
	return func(a *Agent) {
		a.backend = BackendCoqui
		a.coquiURL = baseURL
	}
}

// WithElevenLabs uses the ElevenLabs API (free tier: 10k chars/month).
func WithElevenLabs(apiKey, voiceID string) Option {
	return func(a *Agent) {
		a.backend = BackendElevenLabs
		a.apiKey = apiKey
		a.voiceID = voiceID
	}
}

// WithSystem uses the OS built-in TTS (espeak on Linux, say on macOS, SAPI on Windows).
func WithSystem() Option {
	return func(a *Agent) { a.backend = BackendSystem }
}

// New creates a TTS agent. Defaults to system TTS (always available, no setup).
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

// Speak synthesises text and saves to file (or plays it for system TTS).
func (a *Agent) Speak(ctx context.Context, req Request) (*Result, error) {
	if req.Text == "" {
		return nil, fmt.Errorf("tts: text must not be empty")
	}
	if req.Speed == 0 {
		req.Speed = 1.0
	}
	switch a.backend {
	case BackendCoqui:
		return a.speakCoqui(ctx, req)
	case BackendElevenLabs:
		return a.speakElevenLabs(ctx, req)
	case BackendSystem:
		return a.speakSystem(req)
	default:
		return nil, fmt.Errorf("tts: unsupported backend: %s", a.backend)
	}
}

// --- Coqui TTS ---

func (a *Agent) speakCoqui(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	// Properly encode the text as a query parameter
	params := url.Values{}
	params.Set("text", req.Text)
	if req.Voice != "" {
		params.Set("speaker_id", req.Voice)
	}
	fullURL := a.coquiURL + "/api/tts?" + params.Encode()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("tts[coqui]: build request: %w", err)
	}
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("tts[coqui]: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return nil, fmt.Errorf("tts[coqui]: status %d: %s", resp.StatusCode, raw)
	}
	outPath := req.OutputPath
	if outPath == "" {
		outPath = tempAudio("wav")
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tts[coqui]: read body: %w", err)
	}
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return nil, fmt.Errorf("tts[coqui]: write file: %w", err)
	}
	return &Result{Path: outPath, Backend: BackendCoqui, Latency: time.Since(start)}, nil
}

// --- ElevenLabs (free tier: 10k chars/month) ---

type elevenLabsRequest struct {
	Text          string                 `json:"text"`
	ModelID       string                 `json:"model_id"`
	VoiceSettings map[string]interface{} `json:"voice_settings"`
}

func (a *Agent) speakElevenLabs(ctx context.Context, req Request) (*Result, error) {
	start := time.Now()
	body, err := json.Marshal(elevenLabsRequest{
		Text:    req.Text,
		ModelID: "eleven_monolingual_v1",
		VoiceSettings: map[string]interface{}{
			"stability":        0.5,
			"similarity_boost": 0.75,
		},
	})
	if err != nil {
		return nil, err
	}
	voiceID := a.voiceID
	if req.Voice != "" {
		voiceID = req.Voice
	}
	if voiceID == "" {
		voiceID = "21m00Tcm4TlvDq8ikWAM" // ElevenLabs default — Rachel
	}
	endpoint := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", voiceID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("xi-api-key", a.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "audio/mpeg")
	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("tts[elevenlabs]: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("tts[elevenlabs]: status %d: %s", resp.StatusCode, raw)
	}
	outPath := req.OutputPath
	if outPath == "" {
		outPath = tempAudio("mp3")
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tts[elevenlabs]: read body: %w", err)
	}
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return nil, fmt.Errorf("tts[elevenlabs]: write file: %w", err)
	}
	return &Result{Path: outPath, Backend: BackendElevenLabs, Latency: time.Since(start)}, nil
}

// --- System TTS (no deps, always works) ---

func (a *Agent) speakSystem(req Request) (*Result, error) {
	start := time.Now()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		args := []string{req.Text}
		if req.Voice != "" {
			args = append([]string{"-v", req.Voice}, args...)
		}
		cmd = exec.Command("say", args...)
	case "windows":
		script := fmt.Sprintf(
			`Add-Type -AssemblyName System.Speech; `+
				`$s = New-Object System.Speech.Synthesis.SpeechSynthesizer; `+
				`$s.Speak('%s')`,
			req.Text)
		cmd = exec.Command("powershell", "-NoProfile", "-Command", script)
	default: // Linux + others
		cmd = exec.Command("espeak", req.Text)
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("tts[system]: %w — output: %s", err, out)
	}
	return &Result{Backend: BackendSystem, Latency: time.Since(start)}, nil
}

func tempAudio(ext string) string {
	return filepath.Join(os.TempDir(),
		fmt.Sprintf("nexus-tts-%d.%s", time.Now().UnixNano(), ext))
}
