package writing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Omkar0612/nexus-ai/internal/router"
	"github.com/Omkar0612/nexus-ai/internal/types"
)

func mockLLMServer(response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"content": response}},
			},
			"usage": map[string]int{"prompt_tokens": 10, "completion_tokens": 20},
		})
	}))
}

func newTestAgent(serverURL string) *Agent {
	r := router.New(types.LLMConfig{
		Provider:   "test",
		BaseURL:    serverURL,
		Model:      "test-model",
		TimeoutSec: 10,
	})
	return New(r)
}

func TestDraft(t *testing.T) {
	srv := mockLLMServer("AI agents are transforming productivity in 2026.")
	defer srv.Close()
	a := newTestAgent(srv.URL)
	out, err := a.Draft(context.Background(), "AI agents in 2026", StyleProfessional, 100)
	if err != nil {
		t.Fatalf("Draft: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty draft")
	}
}

func TestRewrite(t *testing.T) {
	srv := mockLLMServer("Rewritten in casual style.")
	defer srv.Close()
	a := newTestAgent(srv.URL)
	out, err := a.Rewrite(context.Background(), "The utilisation of AI is increasing.", StyleCasual)
	if err != nil {
		t.Fatalf("Rewrite: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty rewrite")
	}
}

func TestSummarise(t *testing.T) {
	srv := mockLLMServer("Short summary here.")
	defer srv.Close()
	a := newTestAgent(srv.URL)
	out, err := a.Summarise(context.Background(), "Long article text here...", 20)
	if err != nil {
		t.Fatalf("Summarise: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty summary")
	}
}

func TestProofread(t *testing.T) {
	srv := mockLLMServer("CORRECTED: Fixed text here.\nISSUE: Missing comma in line 2.")
	defer srv.Close()
	a := newTestAgent(srv.URL)
	corrected, issues, err := a.Proofread(context.Background(), "Fixed text hear.")
	if err != nil {
		t.Fatalf("Proofread: %v", err)
	}
	if corrected == "" {
		t.Error("expected corrected text")
	}
	if len(issues) == 0 {
		t.Error("expected at least one issue")
	}
}

func TestTranslate(t *testing.T) {
	srv := mockLLMServer("مرحبا بالعالم")
	defer srv.Close()
	a := newTestAgent(srv.URL)
	out, err := a.Translate(context.Background(), "Hello world", "Arabic")
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if out == "" {
		t.Error("expected translation")
	}
}
