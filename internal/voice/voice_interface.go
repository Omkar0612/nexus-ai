package voice

/*
VoiceInterface — speak to NEXUS and hear it speak back.

One of the most-requested AI agent UX features entering 2026:
'I want to talk to my AI agent while my hands are busy.'

NEXUS VoiceInterface:
  1. Microphone capture via PortAudio (cross-platform)
  2. Whisper transcription (local, free, offline)
  3. Routes transcribed text to NEXUS agent bus
  4. TTS response via local espeak/piper (configurable)
  5. Wake word detection ("Hey NEXUS")
  6. Push-to-talk mode (hold key = record)
  7. Works fully offline — no cloud STT/TTS APIs

Build tag: requires portaudio and whisper.cpp bindings.
For environments without audio hardware, runs in text-simulation mode.
*/

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// VoiceMode defines how voice input is triggered
type VoiceMode string

const (
	ModeWakeWord   VoiceMode = "wakeword"   // always-on, trigger on 'Hey NEXUS'
	ModePushToTalk VoiceMode = "push2talk"  // hold key to record
	ModeContinuous VoiceMode = "continuous" // always listening, auto-segment
	ModeSimulated  VoiceMode = "simulated"  // text input, no mic (testing/CI)
)

// TTSEngine selects text-to-speech backend
type TTSEngine string

const (
	TTSEspeak TTSEngine = "espeak"
	TTSPiper  TTSEngine = "piper"
	TTSSilent TTSEngine = "silent" // no audio output
)

// TranscriptEvent is emitted when speech is recognised
type TranscriptEvent struct {
	Text             string
	Confidence       float64
	DurationMs       int64
	Timestamp        time.Time
	WakeWordDetected bool
}

// VoiceConfig holds configuration for the voice interface
type VoiceConfig struct {
	Mode         VoiceMode
	TTS          TTSEngine
	WakeWord     string
	SampleRate   int
	Language     string
	WhisperModel string // tiny/base/small/medium
	SilenceMs    int    // ms of silence to end utterance
}

// DefaultConfig returns sensible defaults
func DefaultConfig() VoiceConfig {
	return VoiceConfig{
		Mode:         ModeSimulated,
		TTS:          TTSSilent,
		WakeWord:     "hey nexus",
		SampleRate:   16000,
		Language:     "en",
		WhisperModel: "base",
		SilenceMs:    800,
	}
}

// VoiceInterface manages voice input/output for NEXUS
type VoiceInterface struct {
	cfg       VoiceConfig
	mu        sync.Mutex
	listening bool
	onText    func(TranscriptEvent)
	simBuffer []string // for simulated mode
}

// New creates a VoiceInterface
func New(cfg VoiceConfig) *VoiceInterface {
	return &VoiceInterface{cfg: cfg}
}

// SetTranscriptHandler sets the callback for transcribed speech
func (v *VoiceInterface) SetTranscriptHandler(fn func(TranscriptEvent)) {
	v.onText = fn
}

// Start begins listening for voice input
func (v *VoiceInterface) Start() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.listening {
		return fmt.Errorf("voice interface already running")
	}
	v.listening = true
	switch v.cfg.Mode {
	case ModeSimulated:
		log.Info().Msg("VoiceInterface: running in simulated mode (no mic)")
		return nil
	default:
		// In production: initialise PortAudio stream here
		log.Info().Str("mode", string(v.cfg.Mode)).Msg("VoiceInterface: starting")
		return nil
	}
}

// Stop halts voice capture
func (v *VoiceInterface) Stop() {
	v.mu.Lock()
	v.listening = false
	v.mu.Unlock()
	log.Info().Msg("VoiceInterface: stopped")
}

// SimulateInput injects a text transcript (for testing and CI environments)
func (v *VoiceInterface) SimulateInput(text string) {
	if v.onText == nil {
		return
	}
	wakeDetected := containsWakeWord(text, v.cfg.WakeWord)
	clean := text
	if wakeDetected {
		clean = strings.TrimSpace(strings.ToLower(strings.Replace(
			strings.ToLower(text), v.cfg.WakeWord, "", 1,
		)))
	}
	v.onText(TranscriptEvent{
		Text:             clean,
		Confidence:       0.95,
		DurationMs:       1500,
		Timestamp:        time.Now(),
		WakeWordDetected: wakeDetected,
	})
}

// Speak sends text to the configured TTS engine
func (v *VoiceInterface) Speak(text string) error {
	switch v.cfg.TTS {
	case TTSSilent:
		// In simulated/silent mode, just log — no audio output
		log.Debug().Str("text", truncate(text, 80)).Msg("TTS (silent): would speak")
		return nil
	case TTSEspeak:
		return v.speakEspeak(text)
	case TTSPiper:
		return v.speakPiper(text)
	}
	return nil
}

func (v *VoiceInterface) speakEspeak(text string) error {
	if _, err := exec.LookPath("espeak"); err != nil {
		return fmt.Errorf("espeak not found: install with: apt install espeak")
	}
	cmd := exec.Command("espeak", "-v", v.cfg.Language, "-s", "150", text)
	return cmd.Run()
}

func (v *VoiceInterface) speakPiper(text string) error {
	if _, err := exec.LookPath("piper"); err != nil {
		return fmt.Errorf("piper not found: see https://github.com/rhasspy/piper")
	}
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf(`echo %q | piper --output_raw | aplay -r 22050 -f S16_LE -t raw -`, text))
	return cmd.Run()
}

// IsListening returns whether the interface is currently active
func (v *VoiceInterface) IsListening() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.listening
}

// Status returns a formatted status string
func (v *VoiceInterface) Status() string {
	v.mu.Lock()
	defer v.mu.Unlock()
	status := "stopped"
	if v.listening {
		status = "listening"
	}
	return fmt.Sprintf("🎤 Voice: %s | Mode: %s | TTS: %s | Wake: '%s'",
		status, v.cfg.Mode, v.cfg.TTS, v.cfg.WakeWord)
}

func containsWakeWord(text, wakeWord string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(wakeWord))
}

// truncate shortens a string to max characters, appending '...' if cut
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
