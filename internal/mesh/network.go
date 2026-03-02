package mesh

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Network manages peer discovery and intelligent task routing across the local network.
type Network struct {
	mu        sync.RWMutex
	localNode *Node
	peers     map[string]*Node
	client    NodeClient
}

// NodeClient handles HTTP communication between peers
type NodeClient interface {
	Dispatch(ctx context.Context, targetAddress string, req *TaskRequest) (*TaskResponse, error)
}

// NewNetwork initializes the P2P Mesh engine.
func NewNetwork(local *Node, client NodeClient) *Network {
	return &Network{
		localNode: local,
		peers:     make(map[string]*Node),
		client:    client,
	}
}

// RegisterPeer adds or updates a node discovered via mDNS or manual IP config.
func (n *Network) RegisterPeer(peer *Node) {
	n.mu.Lock()
	defer n.mu.Unlock()

	peer.LastSeen = time.Now()
	if _, exists := n.peers[peer.ID]; !exists {
		log.Info().
			Str("peer_id", peer.ID).
			Str("address", peer.Address).
			Bool("has_gpu", peer.Profile.HasGPU).
			Msg("🌐 New NEXUS Node joined the Hive-Mind Mesh!")
	}
	n.peers[peer.ID] = peer
}

// PruneDeadPeers removes nodes that haven't heartbeated in the last 60 seconds.
func (n *Network) PruneDeadPeers() {
	n.mu.Lock()
	defer n.mu.Unlock()

	cutoff := time.Now().Add(-60 * time.Second)
	for id, peer := range n.peers {
		if peer.LastSeen.Before(cutoff) {
			log.Warn().Str("peer_id", id).Msg("🕸️ Peer disconnected from mesh.")
			delete(n.peers, id)
		}
	}
}

// RouteTask intelligently decides whether to execute a task locally or offload it
// to a more powerful peer in the mesh.
func (n *Network) RouteTask(ctx context.Context, req *TaskRequest) (*TaskResponse, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	// Hardware routing logic
	var bestPeer *Node

	switch req.Type {
	case "IMAGE_GEN", "LOCAL_LLM":
		// These require a GPU. If local node lacks a GPU, find a peer that has one.
		if !n.localNode.Profile.HasGPU {
			for _, peer := range n.peers {
				if peer.Profile.HasGPU && peer.Profile.LoadAverage < 0.8 {
					bestPeer = peer
					break
				}
			}
		}
	default:
		// For standard tasks, route to the node with the lowest CPU load.
		lowestLoad := n.localNode.Profile.LoadAverage
		for _, peer := range n.peers {
			if peer.Profile.LoadAverage < lowestLoad {
				lowestLoad = peer.Profile.LoadAverage
				bestPeer = peer
			}
		}
	}

	if bestPeer == nil {
		log.Info().Str("task", req.Type).Msg("Executing task locally.")
		return n.executeLocally(ctx, req)
	}

	log.Info().
		Str("task", req.Type).
		Str("target_peer", bestPeer.ID).
		Msg("🚀 Offloading heavy compute to remote peer in the mesh.")

	return n.client.Dispatch(ctx, bestPeer.Address, req)
}

// executeLocally is a stub representing the local execution pipeline.
func (n *Network) executeLocally(ctx context.Context, req *TaskRequest) (*TaskResponse, error) {
	// In production, this wires into the standard internal/agents execution pool.
	return &TaskResponse{Result: []byte(fmt.Sprintf("executed locally by %s", n.localNode.ID))}, nil
}
