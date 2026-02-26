package dashboard

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMetricStoreRecordAndGet(t *testing.T) {
	store := NewMetricStore(time.Hour)
	store.Record("cost_usd", 0.005, "groq")
	store.Record("cost_usd", 0.010, "openai")
	points := store.Get("cost_usd")
	if len(points) != 2 {
		t.Errorf("expected 2 points, got %d", len(points))
	}
}

func TestMetricStoreSum(t *testing.T) {
	store := NewMetricStore(time.Hour)
	store.Record("cost_usd", 0.50, "")
	store.Record("cost_usd", 0.25, "")
	sum := store.Sum("cost_usd", time.Hour)
	if sum < 0.74 {
		t.Errorf("expected sum ~0.75, got %f", sum)
	}
}

func TestAnalyticsSnapshotEndpoint(t *testing.T) {
	a := New(9876)
	a.Record("cost_usd", 0.05, "test")
	a.UpdateAgentStats([]AgentStat{
		{Name: "Researcher", Role: "researcher", TotalTasks: 10, Failures: 1, SuccessRate: 0.90},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/snapshot", nil)
	w := httptest.NewRecorder()
	a.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.Len() < 10 {
		t.Error("expected non-empty snapshot response")
	}
}

func TestAnalyticsMetricSeriesEndpoint(t *testing.T) {
	a := New(9877)
	a.Record("loop_events", 1, "web-search")
	req := httptest.NewRequest(http.MethodGet, "/api/metrics/loop_events", nil)
	w := httptest.NewRecorder()
	a.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAnalyticsHealthEndpoint(t *testing.T) {
	a := New(9878)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	a.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
