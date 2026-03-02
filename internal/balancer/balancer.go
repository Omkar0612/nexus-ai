package balancer

import (
	"fmt"
	"math/rand" //nolint:gosec // load balancing doesn't require crypto-strength randomness
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Node is a single worker in the NEXUS cluster
type Node struct {
	ID      string
	URL     string
	Healthy bool
	Load    int64
	mu      sync.Mutex
}

// Stats holds node-level runtime stats
type Stats struct {
	NodeID  string `json:"node_id"`
	URL     string `json:"url"`
	Healthy bool   `json:"healthy"`
	Load    int64  `json:"load"`
}

// LoadBalancer distributes requests across NEXUS worker nodes
type LoadBalancer struct {
	nodes  []*Node
	mu     sync.RWMutex
	scheme string // round_robin, least_conn, random
	cursor int
}

// New creates a load balancer with the given node URLs
func New(nodeURLs []string, scheme string) *LoadBalancer {
	nodes := make([]*Node, len(nodeURLs))
	for i, u := range nodeURLs {
		nodes[i] = &Node{ID: fmt.Sprintf("node-%d", i+1), URL: u, Healthy: true}
	}
	if scheme == "" {
		scheme = "least_conn"
	}
	return &LoadBalancer{nodes: nodes, scheme: scheme}
}

// Pick selects the best available node
func (lb *LoadBalancer) Pick() (*Node, error) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	var healthy []*Node
	for _, n := range lb.nodes {
		if n.Healthy {
			healthy = append(healthy, n)
		}
	}
	if len(healthy) == 0 {
		return nil, fmt.Errorf("no healthy nodes available")
	}
	switch lb.scheme {
	case "random":
		return healthy[rand.Intn(len(healthy))], nil //nolint:gosec // non-security random selection
	case "round_robin":
		lb.cursor = (lb.cursor + 1) % len(healthy)
		return healthy[lb.cursor], nil
	default: // least_conn
		best := healthy[0]
		for _, n := range healthy[1:] {
			n.mu.Lock()
			bestLoad, nLoad := best.Load, n.Load
			n.mu.Unlock()
			if nLoad < bestLoad {
				best = n
			}
		}
		return best, nil
	}
}

// ServeHTTP proxies an incoming request to a healthy node
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	node, err := lb.Pick()
	if err != nil {
		http.Error(w, "no healthy nodes: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	node.mu.Lock()
	node.Load++
	node.mu.Unlock()
	defer func() {
		node.mu.Lock()
		node.Load--
		node.mu.Unlock()
	}()
	target, err := url.Parse(node.URL)
	if err != nil {
		http.Error(w, "invalid node URL", http.StatusInternalServerError)
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		log.Error().Err(e).Str("node", node.ID).Msg("proxy error")
		node.mu.Lock()
		node.Healthy = false
		node.mu.Unlock()
		http.Error(w, "upstream error", http.StatusBadGateway)
	}
	proxy.ServeHTTP(w, r)
}

// StartHealthChecks runs background health polling every interval
func (lb *LoadBalancer) StartHealthChecks(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			lb.mu.RLock()
			nodes := make([]*Node, len(lb.nodes))
			copy(nodes, lb.nodes)
			lb.mu.RUnlock()
			for _, node := range nodes {
				go lb.checkNode(node)
			}
		}
	}()
}

func (lb *LoadBalancer) checkNode(node *Node) {
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(node.URL + "/health")
	node.mu.Lock()
	defer node.mu.Unlock()
	if err != nil || resp.StatusCode >= 500 {
		if node.Healthy {
			log.Warn().Str("node", node.ID).Msg("node went unhealthy")
		}
		node.Healthy = false
	} else {
		if !node.Healthy {
			log.Info().Str("node", node.ID).Msg("node recovered")
		}
		node.Healthy = true
	}
	if resp != nil {
		resp.Body.Close() //nolint:gosec
	}
}

// Stats returns current node statistics
func (lb *LoadBalancer) Stats() []Stats {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	stats := make([]Stats, len(lb.nodes))
	for i, n := range lb.nodes {
		n.mu.Lock()
		stats[i] = Stats{NodeID: n.ID, URL: n.URL, Healthy: n.Healthy, Load: n.Load}
		n.mu.Unlock()
	}
	return stats
}
