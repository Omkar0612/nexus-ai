package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// AgentEvent is broadcast to all SSE clients when an agent changes state.
type AgentEvent struct {
	Agent   string `json:"agent"`
	Status  string `json:"status"` // "running" | "done" | "error"
	Message string `json:"message,omitempty"`
}

// SSEHub manages all active SSE client connections.
type SSEHub struct {
	mu      sync.RWMutex
	clients map[chan AgentEvent]struct{}
}

func newSSEHub() *SSEHub {
	return &SSEHub{clients: make(map[chan AgentEvent]struct{})}
}

// Publish broadcasts an event to all connected SSE clients.
// Safe to call from any goroutine.
func (h *SSEHub) Publish(evt AgentEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- evt:
		default: // drop if client is slow — non-blocking
		}
	}
}

// ServeHTTP is the GET /api/events handler — streams agent activity.
func (h *SSEHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan AgentEvent, 16)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, ch)
		h.mu.Unlock()
	}()

	rc := http.NewResponseController(w)
	// Send initial ping so the client knows the stream is alive
	fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
	rc.Flush()

	for {
		select {
		case evt := <-ch:
			b, _ := json.Marshal(evt)
			fmt.Fprintf(w, "data: %s\n\n", b)
			rc.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
