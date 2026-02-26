package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	// maxBodyBytes caps request body at 1 MB to prevent OOM DoS.
	maxBodyBytes = 1 << 20 // 1 MB
	// maxMessageLen caps the LLM prompt length to prevent token-stuffing.
	maxMessageLen = 32_000
)

// agentAllowlist prevents log injection via crafted agent field values.
var agentAllowlist = map[string]bool{
	"router": true, "memory": true, "drift": true, "goal": true,
	"emotional": true, "calendar": true, "browser": true, "email": true,
	"github": true, "writing": true, "imagegen": true, "tts": true,
	"music": true, "plugin": true, "vault": true,
}

type chatReq struct {
	Message string `json:"message"`
	Agent   string `json:"agent,omitempty"`
}

// handleChat accepts a JSON body and streams the LLM response as SSE.
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	// 1. Cap request body size — prevents OOM from unbounded upload.
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req chatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// 2. Validate message.
	if req.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}
	if len(req.Message) > maxMessageLen {
		http.Error(w, "message too long", http.StatusRequestEntityTooLarge)
		return
	}

	// 3. Validate agent field against allowlist to prevent log injection.
	agentName := "router"
	if req.Agent != "" {
		if !agentAllowlist[strings.ToLower(req.Agent)] {
			http.Error(w, "invalid agent", http.StatusBadRequest)
			return
		}
		agentName = strings.ToLower(req.Agent)
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Note: CORS header is set by corsMiddleware in server.go, not here.

	rc := http.NewResponseController(w)

	s.hub.Publish(AgentEvent{
		Agent:   agentName,
		Status:  "running",
		Message: req.Message,
	})

	var content string
	if s.router != nil {
		result, err := s.router.Complete(r.Context(), "You are NEXUS, a helpful AI assistant.", req.Message)
		if err != nil {
			// Log full error server-side; return opaque message to client.
			s.logger.Error().Err(err).Msg("chat: llm error")
			s.hub.Publish(AgentEvent{Agent: agentName, Status: "error", Message: "llm unavailable"})
			fmt.Fprintf(w, "data: Sorry, the AI is temporarily unavailable. Please try again.\n\n")
			rc.Flush()
			return
		}
		content = result.Content
	} else {
		content = "NEXUS v1.7 — connect a real LLM provider to get live answers."
	}

	for _, chunk := range splitWords(content) {
		fmt.Fprintf(w, "data: %s\n\n", chunk)
		rc.Flush()
	}

	s.hub.Publish(AgentEvent{Agent: agentName, Status: "done"})
}

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func splitWords(s string) []string {
	var chunks []string
	runes := []rune(s)
	const chunkSize = 20
	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	return chunks
}
