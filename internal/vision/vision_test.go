package vision

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnalyseBytesOllama(t *testing.T) {
	// Spin up a mock Ollama server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ollamaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		if req.Model == "" || len(req.Images) == 0 {
			http.Error(w, "missing model or image", 400)
			return
		}
		json.NewEncoder(w).Encode(ollamaResponse{Response: "a test image", Model: req.Model})
	}))
	defer ts.Close()

	agent := New(WithOllama(ts.URL, "llava"))
	// Use 1x1 pixel PNG (minimal valid PNG bytes encoded as base64)
	pngB64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwADhQGAWjR9awAAAABJRU5ErkJggg=="
	pngBytes, _ := base64.StdEncoding.DecodeString(pngB64)

	result, err := agent.AnalyseBytes(context.Background(), pngBytes, "what is this?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Description != "a test image" {
		t.Errorf("unexpected description: %q", result.Description)
	}
	if result.Backend != BackendOllama {
		t.Errorf("expected ollama backend")
	}
}

func TestAnalyseBytesOllamaError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	agent := New(WithOllama(ts.URL, "llava"))
	_, err := agent.AnalyseBytes(context.Background(), []byte("fake"), "prompt")
	if err == nil {
		t.Error("expected error on 500 response")
	}
}

func TestNewDefaults(t *testing.T) {
	agent := New()
	if agent.backend != BackendOllama {
		t.Errorf("expected default backend ollama, got %s", agent.backend)
	}
	if agent.model != "llava" {
		t.Errorf("expected default model llava, got %s", agent.model)
	}
}
