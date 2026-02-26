package imagegen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateSD(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "expected POST", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/sdapi/v1/txt2img" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// base64("hello")
		json.NewEncoder(w).Encode(SDResponse{Images: []string{"aGVsbG8="}})
	}))
	defer ts.Close()

	a := New(WithStableDiffusion(ts.URL))
	result, err := a.Generate(context.Background(), Request{
		Prompt:     "a cat on a keyboard",
		OutputPath: t.TempDir() + "/out.png",
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if result.Backend != BackendSD {
		t.Errorf("expected BackendSD, got %s", result.Backend)
	}
	if result.Path == "" {
		t.Error("expected non-empty output path")
	}
}

func TestDefaultsApplied(t *testing.T) {
	req := Request{Prompt: "test"}
	// ensure defaults don't panic in Generate path — use stub server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SDResponse{Images: []string{"aGVsbG8="}})
	}))
	defer ts.Close()
	a := New(WithStableDiffusion(ts.URL))
	req.OutputPath = t.TempDir() + "/default.png"
	_, err := a.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate with defaults: %v", err)
	}
}
