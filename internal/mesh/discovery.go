package mesh

import (
	"context"
	"fmt"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/rs/zerolog/log"
)

const (
	ServiceName = "_nexus-mesh._tcp"
	Domain      = "local."
)

// Discovery handles broadcasting the local node and discovering peers via mDNS.
type Discovery struct {
	network   *Network
	localNode *Node
	server    *zeroconf.Server
}

// NewDiscovery initializes the mDNS service.
func NewDiscovery(net *Network, local *Node) *Discovery {
	return &Discovery{
		network:   net,
		localNode: local,
	}
}

// Start begins advertising the local node and listening for others.
func (d *Discovery) Start(ctx context.Context, port int) error {
	// 1. Advertise Local Node
	txtRecords := []string{
		fmt.Sprintf("id=%s", d.localNode.ID),
		fmt.Sprintf("gpu=%t", d.localNode.Profile.HasGPU),
		fmt.Sprintf("cpu=%s", d.localNode.Profile.CPUModel),
	}

	server, err := zeroconf.Register(
		d.localNode.ID,
		ServiceName,
		Domain,
		port,
		txtRecords,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register mDNS service: %w", err)
	}
	d.server = server
	log.Info().Str("service", ServiceName).Msg("📡 Broadcasting NEXUS Node to local mesh...")

	// 2. Discover Peers
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return fmt.Errorf("failed to initialize mDNS resolver: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			// Ignore our own broadcast
			if entry.Instance == d.localNode.ID {
				continue
			}

			// Parse TXT records for hardware profile
			hasGPU := false
			for _, txt := range entry.Text {
				if txt == "gpu=true" {
					hasGPU = true
				}
			}

			if len(entry.AddrIPv4) == 0 {
				continue
			}

			peerAddr := fmt.Sprintf("%s:%d", entry.AddrIPv4[0].String(), entry.Port)

			peer := &Node{
				ID:      entry.Instance,
				Address: peerAddr,
				Profile: HardwareProfile{
					HasGPU: hasGPU,
				},
				LastSeen: time.Now(),
			}

			d.network.RegisterPeer(peer)
		}
	}(entries)

	// Keep browsing until context is cancelled
	err = resolver.Browse(ctx, ServiceName, Domain, entries)
	if err != nil {
		return fmt.Errorf("failed to browse mDNS: %w", err)
	}

	return nil
}

// Stop halts the mDNS broadcast and discovery.
func (d *Discovery) Stop() {
	if d.server != nil {
		d.server.Shutdown()
	}
}
