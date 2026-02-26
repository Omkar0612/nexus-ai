package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type chatReq struct {
	Message string `json:"message"`
	Agent   string `json:"agent,omitempty"`
}

// handleChat accepts a JSON body and streams the LLM response as SSE.
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	var req chatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	rc := http.NewResponseController(w)

	// Broadcast agent activity start
	s.hub.Publish(AgentEvent{
		Agent:   coalesce(req.Agent, "router"),
		Status:  "running",
		Message: req.Message,
	})

	// Route through LLM router if available, else use stub
	var content string
	if s.router != nil {
		result, err := s.router.Complete(r.Context(), "You are NEXUS, a helpful AI assistant.", req.Message)
		if err != nil {
			s.hub.Publish(AgentEvent{Agent: coalesce(req.Agent, "router"), Status: "error", Message: err.Error()})
			fmt.Fprintf(w, "data: Error: %s\n\n", err.Error())
			rc.Flush()
			return
		}
		content = result.Content
	} else {
		// Test stub — router is nil in unit tests
		content = "NEXUS v1.6 stub response — connect a real LLM provider to get live answers."
	}

	// Stream response word by word for a live feel
	words := splitWords(content)
	for _, chunk := range words {
		fmt.Fprintf(w, "data: %s\n\n", chunk)
		rc.Flush()
	}

	// Broadcast done
	s.hub.Publish(AgentEvent{
		Agent:  coalesce(req.Agent, "router"),
		Status: "done",
	})
}

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// splitWords splits a string into ~4-word chunks for streaming effect.
func splitWords(s string) []string {
	var chunks []string
	words := []rune(s)
	chunkSize := 4 * 5 // ~4 words of ~5 chars
	for i := 0; i < len(words); i += chunkSize {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunks = append(chunks, string(words[i:end]))
	}
	return chunks
}
