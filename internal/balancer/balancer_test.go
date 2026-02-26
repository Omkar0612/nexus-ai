package balancer

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPickLeastConn(t *testing.T) {
	lb := New([]string{"http://localhost:7701", "http://localhost:7702", "http://localhost:7703"}, "least_conn")
	node, err := lb.Pick()
	if err != nil {
		t.Fatalf("Pick: %v", err)
	}
	if node == nil {
		t.Fatal("expected a node, got nil")
	}
}

func TestPickRoundRobin(t *testing.T) {
	lb := New([]string{"http://localhost:7701", "http://localhost:7702"}, "round_robin")
	for i := 0; i < 6; i++ {
		_, err := lb.Pick()
		if err != nil {
			t.Fatalf("round robin pick %d failed: %v", i, err)
		}
	}
}

func TestNoHealthyNodes(t *testing.T) {
	lb := New([]string{"http://localhost:7701"}, "least_conn")
	lb.nodes[0].Healthy = false
	_, err := lb.Pick()
	if err == nil {
		t.Fatal("expected error when no healthy nodes")
	}
}

func TestStats(t *testing.T) {
	lb := New([]string{"http://localhost:7701", "http://localhost:7702"}, "least_conn")
	stats := lb.Stats()
	if len(stats) != 2 {
		t.Errorf("expected 2 stats, got %d", len(stats))
	}
}

func TestServeHTTPNoNodes(t *testing.T) {
	lb := New([]string{"http://localhost:7701"}, "least_conn")
	lb.nodes[0].Healthy = false
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	lb.ServeHTTP(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}
