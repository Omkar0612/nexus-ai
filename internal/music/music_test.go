package music

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMusicStub(t *testing.T) {
	a := New() // defaults to stub
	tmp := t.TempDir() + "/test.wav"
	result, err := a.Generate(context.Background(), Request{
		Prompt:     "calm lo-fi beats",
		Duration:   5 * time.Second,
		OutputPath: tmp,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if result.Backend != BackendStub {
		t.Errorf("expected stub backend, got %s", result.Backend)
	}
	info, err := os.Stat(tmp)
	if err != nil {
		t.Fatalf("output file missing: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
	// Validate WAV magic bytes
	data, _ := os.ReadFile(tmp)
	if len(data) < 4 || string(data[:4]) != "RIFF" {
		t.Error("output is not a valid WAV file")
	}
}

func TestMusicEmptyPrompt(t *testing.T) {
	a := New()
	_, err := a.Generate(context.Background(), Request{Prompt: ""})
	if err == nil {
		t.Error("expected error for empty prompt")
	}
}

func TestMusicDurationDefault(t *testing.T) {
	a := New()
	tmp := t.TempDir() + "/dur.wav"
	// No duration set — should default to 10s without panicking
	_, err := a.Generate(context.Background(), Request{
		Prompt:     "test",
		OutputPath: tmp,
	})
	if err != nil {
		t.Fatalf("Generate with default duration: %v", err)
	}
}
