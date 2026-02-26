package voice

import (
	"testing"
)

func TestVoiceStartStop(t *testing.T) {
	v := New(DefaultConfig())
	if err := v.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !v.IsListening() {
		t.Error("expected IsListening to be true after Start")
	}
	v.Stop()
	if v.IsListening() {
		t.Error("expected IsListening to be false after Stop")
	}
}

func TestVoiceDoubleStart(t *testing.T) {
	v := New(DefaultConfig())
	_ = v.Start()
	if err := v.Start(); err == nil {
		t.Error("expected error on double Start")
	}
	v.Stop()
}

func TestVoiceSimulateInput(t *testing.T) {
	v := New(DefaultConfig())
	_ = v.Start()

	received := ""
	v.SetTranscriptHandler(func(evt TranscriptEvent) {
		received = evt.Text
	})

	v.SimulateInput("nexus run drift scan")
	if received == "" {
		t.Error("expected transcript to be received")
	}
}

func TestVoiceWakeWordDetection(t *testing.T) {
	v := New(DefaultConfig())
	_ = v.Start()

	var wakeDetected bool
	v.SetTranscriptHandler(func(evt TranscriptEvent) {
		wakeDetected = evt.WakeWordDetected
	})

	v.SimulateInput("hey nexus what are my goals?")
	if !wakeDetected {
		t.Error("expected wake word to be detected")
	}
}

func TestVoiceSpeakSilent(t *testing.T) {
	v := New(DefaultConfig()) // TTS = silent by default
	if err := v.Speak("Hello Omkar, your drift scan is complete."); err != nil {
		t.Errorf("Speak (silent): %v", err)
	}
}

func TestVoiceStatus(t *testing.T) {
	v := New(DefaultConfig())
	status := v.Status()
	if status == "" {
		t.Error("expected non-empty status")
	}
}
