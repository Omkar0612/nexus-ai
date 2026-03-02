package mesh

import (
	"testing"
	"time"
)

func TestNewMeshManager(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			config: &Config{
				DiscoveryPort: 1000,
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: &Config{
				DiscoveryPort: 70000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gpu := GPUInfo{
				Model:       "Test GPU",
				MemoryTotal: 8192,
				MemoryFree:  4096,
			}
			_, err := NewMeshManager(gpu, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMeshManager() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSubmitTask(t *testing.T) {
	gpu := GPUInfo{
		Model:       "Test GPU",
		MemoryTotal: 8192,
		MemoryFree:  4096,
	}
	manager, err := NewMeshManager(gpu, DefaultConfig())
	if err != nil {
		t.Fatalf("Failed to create mesh manager: %v", err)
	}

	tests := []struct {
		name    string
		task    *TaskRequest
		wantErr bool
	}{
		{
			name: "valid task",
			task: &TaskRequest{
				ID:             "task-1",
				Type:           "inference",
				RequiredMemory: 1024,
			},
			wantErr: false,
		},
		{
			name:    "nil task",
			task:    nil,
			wantErr: true,
		},
		{
			name: "empty task ID",
			task: &TaskRequest{
				ID:             "",
				Type:           "inference",
				RequiredMemory: 1024,
			},
			wantErr: true,
		},
		{
			name: "negative memory",
			task: &TaskRequest{
				ID:             "task-2",
				Type:           "inference",
				RequiredMemory: -100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.SubmitTask(tt.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("SubmitTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSelectBestPeer(t *testing.T) {
	gpu := GPUInfo{
		Model:       "Test GPU",
		MemoryTotal: 8192,
		MemoryFree:  4096,
	}
	manager, _ := NewMeshManager(gpu, DefaultConfig())

	// Add test peers
	manager.peers = map[string]*PeerInfo{
		"peer1": {
			ID:        "peer1",
			Available: true,
			GPUInfo: GPUInfo{
				MemoryFree: 2048,
			},
			CurrentLoad: 0.5,
		},
		"peer2": {
			ID:        "peer2",
			Available: true,
			GPUInfo: GPUInfo{
				MemoryFree: 4096,
			},
			CurrentLoad: 0.2, // Lower load, should be selected
		},
		"peer3": {
			ID:        "peer3",
			Available: false, // Not available
			GPUInfo: GPUInfo{
				MemoryFree: 8192,
			},
			CurrentLoad: 0.1,
		},
	}

	task := &TaskRequest{
		RequiredMemory: 1024,
	}

	peer := manager.selectBestPeer(task)
	if peer == nil {
		t.Fatal("Expected to find a peer, got nil")
	}
	if peer.ID != "peer2" {
		t.Errorf("Expected peer2 (best score), got %s", peer.ID)
	}
}

func TestGetActivePeers(t *testing.T) {
	gpu := GPUInfo{Model: "Test GPU"}
	manager, _ := NewMeshManager(gpu, DefaultConfig())

	now := time.Now()
	manager.peers = map[string]*PeerInfo{
		"peer1": {
			ID:        "peer1",
			Available: true,
			LastSeen:  now,
		},
		"peer2": {
			ID:        "peer2",
			Available: true,
			LastSeen:  now.Add(-60 * time.Second), // Stale
		},
		"peer3": {
			ID:        "peer3",
			Available: false, // Not available
			LastSeen:  now,
		},
	}

	active := manager.GetActivePeers()
	if len(active) != 1 {
		t.Errorf("Expected 1 active peer, got %d", len(active))
	}
	if active[0].ID != "peer1" {
		t.Errorf("Expected peer1, got %s", active[0].ID)
	}
}

func TestGenerateSecurePeerID(t *testing.T) {
	id1 := generateSecurePeerID()
	id2 := generateSecurePeerID()

	if id1 == id2 {
		t.Error("Expected unique peer IDs, got duplicates")
	}

	if len(id1) < 10 {
		t.Errorf("Peer ID too short: %s", id1)
	}
}
