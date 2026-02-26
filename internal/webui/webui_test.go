package webui

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	// Pass nil router — chat.go uses stub response when router is nil
	srv := New(":0", zerolog.Nop(), nil)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/chat", srv.handleChat)
	mux.HandleFunc("GET /api/events", srv.hub.ServeHTTP)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})
	return httptest.NewServer(mux)
}

func TestHealthEndpoint(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestChatEndpointBadBody(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()
	resp, err := http.Post(ts.URL+"/api/chat", "application/json", bytes.NewBufferString("not-json"))
	if err != nil {
		t.Fatalf("chat request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestChatEndpointStreaming(t *testing.T) {
	ts := newTestServer(t)
	defer ts.Close()
	body, _ := json.Marshal(chatReq{Message: "hello"})
	resp, err := http.Post(ts.URL+"/api/chat", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("chat request failed: %v", err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Errorf("expected text/event-stream, got %s", ct)
	}
	scanner := bufio.NewScanner(resp.Body)
	var chunks []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			chunks = append(chunks, strings.TrimPrefix(line, "data: "))
		}
	}
	if len(chunks) == 0 {
		t.Error("expected at least one SSE data chunk")
	}
}

func TestSSEHubPublish(t *testing.T) {
	hub := newSSEHub()
	ts := httptest.NewServer(http.HandlerFunc(hub.ServeHTTP))
	defer ts.Close()

	clientDone := make(chan AgentEvent, 1)
	go func() {
		resp, err := http.Get(ts.URL)
		if err != nil {
			clientDone <- AgentEvent{}
			return
		}
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			var evt AgentEvent
			if err := json.Unmarshal([]byte(data), &evt); err != nil {
				continue
			}
			// Skip ping (empty agent)
			if evt.Agent == "" {
				continue
			}
			clientDone <- evt
			return
		}
	}()

	time.Sleep(100 * time.Millisecond)
	hub.Publish(AgentEvent{Agent: "calendar", Status: "running", Message: "syncing"})

	select {
	case evt := <-clientDone:
		if evt.Agent != "calendar" || evt.Status != "running" {
			t.Errorf("unexpected event: %+v", evt)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout: SSE event never received")
	}
}
