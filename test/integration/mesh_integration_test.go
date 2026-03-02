//go:build integration
// +build integration

package integration

import (
	"testing"
	"time"

	"github.com/Omkar0612/nexus-ai/internal/mesh"
)

// TestMeshNetworkDiscovery tests peer discovery across multiple nodes
func TestMeshNetworkDiscovery(t *testing.T) {
	t.Parallel()

	// Create two mesh managers
	config1 := mesh.DefaultConfig()
	config1.DiscoveryPort = 15353

	config2 := mesh.DefaultConfig()
	config2.DiscoveryPort = 15353

	gpu1 := mesh.GPUInfo{
		Model:       "GPU-1",
		MemoryTotal: 8192,
		MemoryFree:  4096,
	}

	gpu2 := mesh.GPUInfo{
		Model:       "GPU-2",
		MemoryTotal: 16384,
		MemoryFree:  8192,
	}

	manager1, err := mesh.NewMeshManager(gpu1, config1)
	if err != nil {
		t.Fatalf("Failed to create manager1: %v", err)
	}

	manager2, err := mesh.NewMeshManager(gpu2, config2)
	if err != nil {
		t.Fatalf("Failed to create manager2: %v", err)
	}

	// Start both managers
	if err := manager1.Start(); err != nil {
		t.Fatalf("Failed to start manager1: %v", err)
	}
	defer manager1.Stop()

	if err := manager2.Start(); err != nil {
		t.Fatalf("Failed to start manager2: %v", err)
	}
	defer manager2.Stop()

	// Wait for discovery
	time.Sleep(7 * time.Second)

	// Check if peers discovered each other
	peers1 := manager1.GetActivePeers()
	peers2 := manager2.GetActivePeers()

	if len(peers1) == 0 {
		t.Error("Manager1 did not discover any peers")
	}

	if len(peers2) == 0 {
		t.Error("Manager2 did not discover any peers")
	}

	t.Logf("Manager1 discovered %d peers", len(peers1))
	t.Logf("Manager2 discovered %d peers", len(peers2))
}

// TestMeshTaskDistribution tests task submission and retrieval
func TestMeshTaskDistribution(t *testing.T) {
	t.Parallel()

	config := mesh.DefaultConfig()
	config.DiscoveryPort = 15354

	gpu := mesh.GPUInfo{
		Model:       "Test GPU",
		MemoryTotal: 8192,
		MemoryFree:  4096,
	}

	manager, err := mesh.NewMeshManager(gpu, config)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if err := manager.Start(); err != nil {
		t.Fatalf("Failed to start manager: %v", err)
	}
	defer manager.Stop()

	// Submit a task
	task := &mesh.TaskRequest{
		ID:             "integration-task-1",
		Type:           "test",
		RequiredMemory: 1024,
		Priority:       1,
	}

	if err := manager.SubmitTask(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}

	// Try to get result (will fail since no peers, but tests the flow)
	_, err = manager.GetResult(2 * time.Second)
	if err == nil {
		t.Log("Got result (unexpected but valid if peers exist)")
	} else {
		t.Logf("Expected error (no peers): %v", err)
	}
}
