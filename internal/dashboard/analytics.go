package dashboard

/*
Analytics — web UI data layer for NEXUS.

Every production AI agent system needs observability.
This package provides the JSON API layer that powers the NEXUS
web dashboard, exposing:
  1. Cost metrics — spend by model, day, agent
  2. Audit timeline — every agent action with risk breakdown
  3. Goal progress — completion rates, stall times
  4. Agent performance — tasks, failures, latency per agent
  5. Loop detection events — tokens saved
  6. Scheduler status — job history, next runs
  7. KB stats — document count, search hit rates

HTTP handlers return JSON. Plug in any frontend (React, HTMX, Grafana).
No external analytics service needed. All data is local.
*/

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// MetricPoint is a single time-series data point
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Label     string    `json:"label,omitempty"`
}

// AgentStat holds performance metrics for one agent
type AgentStat struct {
	Name        string        `json:"name"`
	Role        string        `json:"role"`
	TotalTasks  int           `json:"total_tasks"`
	Failures    int           `json:"failures"`
	SuccessRate float64       `json:"success_rate"`
	AvgLatency  time.Duration `json:"avg_latency_ms"`
	LastActive  time.Time     `json:"last_active"`
}

// DashboardSnapshot is a full point-in-time system snapshot
type DashboardSnapshot struct {
	GeneratedAt    time.Time              `json:"generated_at"`
	CostToday      float64                `json:"cost_today_usd"`
	CostMonth      float64                `json:"cost_month_usd"`
	BudgetPct      float64                `json:"budget_pct"`
	TotalAgentTasks int                   `json:"total_agent_tasks"`
	HighRiskActions int                   `json:"high_risk_actions_24h"`
	LoopsDetected  int                    `json:"loops_detected"`
	TokensSaved    int                    `json:"tokens_saved_by_loop_detector"`
	KBDocuments    int                    `json:"kb_documents"`
	ActiveGoals    int                    `json:"active_goals"`
	StalledGoals   int                    `json:"stalled_goals"`
	Agents         []AgentStat            `json:"agents"`
	CostSeries     []MetricPoint          `json:"cost_series"`
	CustomMetrics  map[string]interface{} `json:"custom_metrics,omitempty"`
}

// MetricStore is an in-memory time-series store for dashboard metrics
type MetricStore struct {
	mu      sync.RWMutex
	series  map[string][]MetricPoint
	maxAge  time.Duration
}

// NewMetricStore creates a MetricStore with a retention window
func NewMetricStore(retention time.Duration) *MetricStore {
	if retention <= 0 {
		retention = 30 * 24 * time.Hour // 30 days default
	}
	return &MetricStore{
		series: make(map[string][]MetricPoint),
		maxAge: retention,
	}
}

// Record adds a metric data point
func (m *MetricStore) Record(name string, value float64, label string) {
	pt := MetricPoint{Timestamp: time.Now(), Value: value, Label: label}
	m.mu.Lock()
	m.series[name] = append(m.series[name], pt)
	m.mu.Unlock()
	m.prune(name)
}

func (m *MetricStore) prune(name string) {
	cutoff := time.Now().Add(-m.maxAge)
	m.mu.Lock()
	defer m.mu.Unlock()
	series := m.series[name]
	for i, pt := range series {
		if pt.Timestamp.After(cutoff) {
			m.series[name] = series[i:]
			return
		}
	}
}

// Get returns all points for a named series
func (m *MetricStore) Get(name string) []MetricPoint {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]MetricPoint, len(m.series[name]))
	copy(result, m.series[name])
	return result
}

// Sum returns the sum of all values in a series within a time window
func (m *MetricStore) Sum(name string, since time.Duration) float64 {
	cutoff := time.Now().Add(-since)
	m.mu.RLock()
	defer m.mu.RUnlock()
	var total float64
	for _, pt := range m.series[name] {
		if pt.Timestamp.After(cutoff) {
			total += pt.Value
		}
	}
	return total
}

// Analytics is the NEXUS dashboard analytics engine
type Analytics struct {
	store   *MetricStore
	agents  []AgentStat
	mu      sync.RWMutex
	mux     *http.ServeMux
	port    int
}

// New creates an Analytics instance
func New(port int) *Analytics {
	a := &Analytics{
		store: NewMetricStore(30 * 24 * time.Hour),
		port:  port,
		mux:   http.NewServeMux(),
	}
	a.registerRoutes()
	return a
}

func (a *Analytics) registerRoutes() {
	a.mux.HandleFunc("/api/snapshot", a.handleSnapshot)
	a.mux.HandleFunc("/api/metrics/", a.handleMetricSeries)
	a.mux.HandleFunc("/api/agents", a.handleAgents)
	a.mux.HandleFunc("/health", a.handleHealth)
}

// Serve starts the analytics HTTP server
func (a *Analytics) Serve() error {
	addr := fmt.Sprintf(":%d", a.port)
	fmt.Printf("📊 NEXUS Analytics dashboard: http://localhost%s\n", addr)
	return http.ListenAndServe(addr, a.mux)
}

// Record proxies to the metric store
func (a *Analytics) Record(metric string, value float64, label string) {
	a.store.Record(metric, value, label)
}

// UpdateAgentStats replaces the agent stats list
func (a *Analytics) UpdateAgentStats(stats []AgentStat) {
	a.mu.Lock()
	a.agents = stats
	a.mu.Unlock()
}

func (a *Analytics) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	agents := make([]AgentStat, len(a.agents))
	copy(agents, a.agents)
	a.mu.RUnlock()

	snapshot := DashboardSnapshot{
		GeneratedAt: time.Now(),
		CostToday:   a.store.Sum("cost_usd", 24*time.Hour),
		CostMonth:   a.store.Sum("cost_usd", 30*24*time.Hour),
		Agents:      agents,
		CostSeries:  a.store.Get("cost_usd"),
	}

	tasks := 0
	for _, ag := range agents {
		tasks += ag.TotalTasks
	}
	snapshot.TotalAgentTasks = tasks

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

func (a *Analytics) handleMetricSeries(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/metrics/")
	if name == "" {
		http.Error(w, "metric name required", http.StatusBadRequest)
		return
	}
	points := a.store.Get(name)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(points)
}

func (a *Analytics) handleAgents(w http.ResponseWriter, r *http.Request) {
	a.mu.RLock()
	agents := make([]AgentStat, len(a.agents))
	copy(agents, a.agents)
	a.mu.RUnlock()
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].TotalTasks > agents[j].TotalTasks
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

func (a *Analytics) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","time":%q}`, time.Now().Format(time.RFC3339))
}
