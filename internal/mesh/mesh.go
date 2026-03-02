package mesh

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// PeerInfo represents a discovered peer in the mesh network
type PeerInfo struct {
	ID           string    `json:"id"`
	Address      string    `json:"address"`
	Port         int       `json:"port"`
	GPUInfo      GPUInfo   `json:"gpu_info"`
	LastSeen     time.Time `json:"last_seen"`
	Available    bool      `json:"available"`
	CurrentLoad  float64   `json:"current_load"`
}

// GPUInfo holds GPU capabilities and status
type GPUInfo struct {
	Model        string  `json:"model"`
	MemoryTotal  int64   `json:"memory_total_mb"`
	MemoryFree   int64   `json:"memory_free_mb"`
	ComputeCaps  string  `json:"compute_caps"`
	DriverVer    string  `json:"driver_version"`
	Utilization  float64 `json:"utilization_percent"`
}

// TaskRequest represents a computational task to be distributed
type TaskRequest struct {
	ID              string            `json:"id"`
	Type            string            `json:"type"`
	Payload         map[string]any    `json:"payload"`
	RequiredMemory  int64             `json:"required_memory_mb"`
	Priority        int               `json:"priority"`
	TimeoutSeconds  int               `json:"timeout_seconds"`
}

// TaskResult contains the result of a distributed task
type TaskResult struct {
	TaskID      string         `json:"task_id"`
	PeerID      string         `json:"peer_id"`
	Result      map[string]any `json:"result"`
	Error       string         `json:"error,omitempty"`
	Duration    time.Duration  `json:"duration"`
	CompletedAt time.Time      `json:"completed_at"`
}

// MeshManager coordinates P2P GPU resource sharing
type MeshManager struct {
	mu              sync.RWMutex
	peers           map[string]*PeerInfo
	localPeer       *PeerInfo
	discoveryPort   int
	taskQueue       chan *TaskRequest
	resultQueue     chan *TaskResult
	ctx             context.Context
	cancel          context.CancelFunc
	discoveryTicker *time.Ticker
}

// NewMeshManager creates a new P2P mesh coordinator
func NewMeshManager(localGPUInfo GPUInfo, discoveryPort int) *MeshManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	localPeer := &PeerInfo{
		ID:          generatePeerID(),
		Address:     getLocalIP(),
		Port:        discoveryPort,
		GPUInfo:     localGPUInfo,
		LastSeen:    time.Now(),
		Available:   true,
		CurrentLoad: 0.0,
	}

	return &MeshManager{
		peers:         make(map[string]*PeerInfo),
		localPeer:     localPeer,
		discoveryPort: discoveryPort,
		taskQueue:     make(chan *TaskRequest, 100),
		resultQueue:   make(chan *TaskResult, 100),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start begins peer discovery and task distribution
func (m *MeshManager) Start() error {
	log.Info().Str("peer_id", m.localPeer.ID).Msg("Starting mesh network manager")

	// Start mDNS discovery
	go m.runDiscovery()

	// Start task scheduler
	go m.runTaskScheduler()

	// Start peer heartbeat
	go m.runPeerHealthCheck()

	return nil
}

// Stop gracefully shuts down the mesh manager
func (m *MeshManager) Stop() error {
	log.Info().Msg("Stopping mesh network manager")
	m.cancel()
	if m.discoveryTicker != nil {
		m.discoveryTicker.Stop()
	}
	close(m.taskQueue)
	close(m.resultQueue)
	return nil
}

// SubmitTask adds a computational task to the distribution queue
func (m *MeshManager) SubmitTask(task *TaskRequest) error {
	select {
	case m.taskQueue <- task:
		log.Debug().Str("task_id", task.ID).Msg("Task submitted to mesh")
		return nil
	case <-m.ctx.Done():
		return fmt.Errorf("mesh manager is shutting down")
	default:
		return fmt.Errorf("task queue is full")
	}
}

// GetResult retrieves a completed task result
func (m *MeshManager) GetResult(timeout time.Duration) (*TaskResult, error) {
	select {
	case result := <-m.resultQueue:
		return result, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for result")
	case <-m.ctx.Done():
		return nil, fmt.Errorf("mesh manager stopped")
	}
}

// GetActivePeers returns list of currently active peers
func (m *MeshManager) GetActivePeers() []*PeerInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := make([]*PeerInfo, 0, len(m.peers))
	for _, peer := range m.peers {
		if peer.Available && time.Since(peer.LastSeen) < 30*time.Second {
			active = append(active, peer)
		}
	}
	return active
}

// runDiscovery handles mDNS peer discovery
func (m *MeshManager) runDiscovery() {
	m.discoveryTicker = time.NewTicker(5 * time.Second)
	defer m.discoveryTicker.Stop()

	// Broadcast presence
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-m.discoveryTicker.C:
			m.broadcastPresence()
			m.discoverPeers()
		}
	}
}

// broadcastPresence announces this peer to the network
func (m *MeshManager) broadcastPresence() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: m.discoveryPort,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create UDP listener")
		return
	}
	defer conn.Close()

	m.mu.RLock()
	m.localPeer.LastSeen = time.Now()
	data, _ := json.Marshal(m.localPeer)
	m.mu.RUnlock()

	addr := &net.UDPAddr{
		IP:   net.IPv4(224, 0, 0, 251), // mDNS multicast
		Port: m.discoveryPort,
	}

	_, err = conn.WriteToUDP(data, addr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to broadcast presence")
	}
}

// discoverPeers listens for peer announcements
func (m *MeshManager) discoverPeers() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: m.discoveryPort,
	})
	if err != nil {
		return
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	buf := make([]byte, 4096)

	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		return
	}

	var peer PeerInfo
	if err := json.Unmarshal(buf[:n], &peer); err != nil {
		return
	}

	// Don't add self
	if peer.ID == m.localPeer.ID {
		return
	}

	m.mu.Lock()
	m.peers[peer.ID] = &peer
	m.mu.Unlock()

	log.Debug().Str("peer_id", peer.ID).Str("address", peer.Address).Msg("Discovered peer")
}

// runTaskScheduler distributes tasks to optimal peers
func (m *MeshManager) runTaskScheduler() {
	for {
		select {
		case <-m.ctx.Done():
			return
		case task := <-m.taskQueue:
			m.scheduleTask(task)
		}
	}
}

// scheduleTask finds best peer and executes task
func (m *MeshManager) scheduleTask(task *TaskRequest) {
	peer := m.selectBestPeer(task)
	if peer == nil {
		log.Warn().Str("task_id", task.ID).Msg("No suitable peer found")
		m.resultQueue <- &TaskResult{
			TaskID: task.ID,
			Error:  "no available peers",
		}
		return
	}

	log.Info().
		Str("task_id", task.ID).
		Str("peer_id", peer.ID).
		Msg("Executing task on peer")

	go m.executeTask(task, peer)
}

// selectBestPeer chooses optimal peer based on load and capabilities
func (m *MeshManager) selectBestPeer(task *TaskRequest) *PeerInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var bestPeer *PeerInfo
	bestScore := -1.0

	for _, peer := range m.peers {
		if !peer.Available {
			continue
		}

		if peer.GPUInfo.MemoryFree < task.RequiredMemory {
			continue
		}

		// Score based on load and available memory
		score := (1.0 - peer.CurrentLoad) * float64(peer.GPUInfo.MemoryFree)
		if score > bestScore {
			bestScore = score
			bestPeer = peer
		}
	}

	return bestPeer
}

// executeTask runs the task and returns result
func (m *MeshManager) executeTask(task *TaskRequest, peer *PeerInfo) {
	start := time.Now()

	// TODO: Implement actual RPC call to peer
	// For now, simulate execution
	time.Sleep(100 * time.Millisecond)

	result := &TaskResult{
		TaskID:      task.ID,
		PeerID:      peer.ID,
		Result:      map[string]any{"status": "completed"},
		Duration:    time.Since(start),
		CompletedAt: time.Now(),
	}

	m.resultQueue <- result
}

// runPeerHealthCheck removes stale peers
func (m *MeshManager) runPeerHealthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupStalePeers()
		}
	}
}

// cleanupStalePeers removes peers that haven't been seen
func (m *MeshManager) cleanupStalePeers() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, peer := range m.peers {
		if time.Since(peer.LastSeen) > 60*time.Second {
			log.Info().Str("peer_id", id).Msg("Removing stale peer")
			delete(m.peers, id)
		}
	}
}

// Helper functions

func generatePeerID() string {
	return fmt.Sprintf("peer-%d", time.Now().UnixNano())
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
