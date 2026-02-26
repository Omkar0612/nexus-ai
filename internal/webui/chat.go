package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type chatReq struct {
	Message string `json:"message"`
	Agent   string `json:"agent,omitempty"` // optional agent override
}

// handleChat accepts a JSON body and streams an LLM response as SSE.
// It delegates to internal/router for all agent dispatch logic.
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

	// Publish activity start event
	s.hub.Publish(AgentEvent{
		Agent:   req.Agent,
		Status:  "running",
		Message: req.Message,
	})

	// TODO: replace stub with router.Stream(r.Context(), req.Message, req.Agent, w, rc)
	chunks := []string{"Hello", " from", " NEXUS", " v1.6!"}
	for _, chunk := range chunks {
		fmt.Fprintf(w, "data: %s\n\n", chunk)
		rc.Flush()
	}

	// Publish activity done event
	s.hub.Publish(AgentEvent{
		Agent:  req.Agent,
		Status: "done",
	})
}
