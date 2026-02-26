package tts

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCoquiTTS(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tts" {
			http.NotFound(w, r)
			return
		}
		text := r.URL.Query().Get("text")
		if text == "" {
			http.Error(w, "missing text", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "audio/wav")
		w.Write([]byte{0x52, 0x49, 0x46, 0x46, 0x00}) //nolint:errcheck
	}))
	defer srv.Close()

	a := New(WithCoqui(srv.URL))
	result, err := a.Speak(context.Background(), Request{
		Text:       "Hello NEXUS",
		OutputPath: t.TempDir() + "/out.wav",
	})
	if err != nil {
		t.Fatalf("Speak: %v", err)
	}
	if result.Backend != BackendCoqui {
		t.Errorf("expected coqui, got %s", result.Backend)
	}
	if result.Path == "" {
		t.Error("expected output path")
	}
}

func TestEmptyTextError(t *testing.T) {
	a := New()
	_, err := a.Speak(context.Background(), Request{Text: ""})
	if err == nil {
		t.Error("expected error for empty text")
	}
}

func TestCoquiTextEncoding(t *testing.T) {
	// Ensure special chars in text are properly URL-encoded
	var receivedText string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedText = r.URL.Query().Get("text")
		w.Header().Set("Content-Type", "audio/wav")
		w.Write([]byte{0x00}) //nolint:errcheck
	}))
	defer srv.Close()

	a := New(WithCoqui(srv.URL))
	_, err := a.Speak(context.Background(), Request{
		Text:       "Hello & World! 100%",
		OutputPath: t.TempDir() + "/enc.wav",
	})
	if err != nil {
		t.Fatalf("Speak: %v", err)
	}
	if receivedText != "Hello & World! 100%" {
		t.Errorf("text not decoded correctly: %q", receivedText)
	}
}
