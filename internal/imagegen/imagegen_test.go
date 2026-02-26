package imagegen

import (
	"net/http"
	"net/http/httptest"
	"context"
	"encoding/json"
	"testing"
)

func TestGenerateSD(t *testing.T) {
	// Fake SD server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sdapi/v1/txt2img" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(sdResponse{Images: []string{"aGVsbG8="}})
	}))
	defer ts.Close()

	a := New(WithStableDiffusion(ts.URL))
	result, err := a.Generate(context.Background(), Request{Prompt: "a cat on a keyboard"})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if result.Base64 == "" {
		t.Error("expected base64 data")
	}
	if result.Backend != BackendSD {
		t.Errorf("expected SD backend, got %s", result.Backend)
	}
}
