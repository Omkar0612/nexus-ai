package mesh

import (
	"context"
	"testing"
	"time"
)

type mockClient struct {
	dispatchedTo string
}

func (m *mockClient) Dispatch(ctx context.Context, targetAddress string, req *TaskRequest) (*TaskResponse, error) {
	m.dispatchedTo = targetAddress
	return &TaskResponse{Result: []byte("remote success")}, nil
}

func TestMeshNetwork_HardwareRouting(t *testing.T) {
	// 1. Setup Local Node (e.g., iPhone) - No GPU, high load
	localPhone := &Node{
		ID:      "iphone_01",
		Address: "127.0.0.1",
		Profile: HardwareProfile{
			HasGPU:      false,
			LoadAverage: 0.9,
		},
	}

	client := &mockClient{}
	net := NewNetwork(localPhone, client)

	// 2. Register Remote Node (e.g., Desktop PC) - Has GPU, low load
	remotePC := &Node{
		ID:      "desktop_rtx4090",
		Address: "192.168.1.100:7070",
		Profile: HardwareProfile{
			HasGPU:      true,
			LoadAverage: 0.1,
		},
	}
	net.RegisterPeer(remotePC)

	// 3. Request Image Generation
	req := &TaskRequest{
		TaskType: "IMAGE_GEN",
		Payload:  []byte("A futuristic cyberpunk city"),
	}

	_, err := net.RouteTask(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 4. Verify the task was correctly offloaded to the Desktop PC
	if client.dispatchedTo != remotePC.Address {
		t.Errorf("expected task to route to %s, got %s", remotePC.Address, client.dispatchedTo)
	}
}

func TestMeshNetwork_PruneDeadPeers(t *testing.T) {
	net := NewNetwork(&Node{ID: "local"}, &mockClient{})
	
	deadPeer := &Node{ID: "dead_peer"}
	net.RegisterPeer(deadPeer)
	
	// Manually age the peer backwards past the 60s timeout
	net.peers["dead_peer"].LastSeen = time.Now().Add(-65 * time.Second)
	
	net.PruneDeadPeers()
	
	if len(net.peers) != 0 {
		t.Errorf("expected 0 peers after pruning, got %d", len(net.peers))
	}
}
