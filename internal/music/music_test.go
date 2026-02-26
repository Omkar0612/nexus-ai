package music

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMusicStub(t *testing.T) {
	a := New() // defaults to stub backend
	tmp := t.TempDir() + "/test.wav"
	result, err := a.Generate(context.Background(), Request{
		Prompt:     "calm lo-fi beats",
		Duration:   5 * time.Second,
		OutputPath: tmp,
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
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
}
