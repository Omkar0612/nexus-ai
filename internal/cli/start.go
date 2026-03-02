package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Omkar0612/nexus-ai/internal/router"
	"github.com/Omkar0612/nexus-ai/internal/types"
	"github.com/Omkar0612/nexus-ai/internal/webui"
	
	// v1.8 feature imports
	"github.com/Omkar0612/nexus-ai/internal/routing"
	"github.com/Omkar0612/nexus-ai/internal/mesh"
	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the NEXUS daemon",
	Long:  `Starts the NEXUS background daemon with all agents, memory, Web UI and gateway.`,
	RunE:  runStart,
}

func init() {
	startCmd.Flags().IntP("port", "p", 7700, "Gateway port")
	startCmd.Flags().String("host", "127.0.0.1", "Bind host")
	startCmd.Flags().String("webui-addr", ":7070", "Web UI listen address (e.g. :7070)")
	startCmd.Flags().BoolP("no-tui", "n", false, "Disable terminal UI")
	startCmd.Flags().Bool("no-webui", false, "Disable web UI")
}

func runStart(cmd *cobra.Command, _ []string) error {
	debug, _ := cmd.Root().PersistentFlags().GetBool("debug")
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	port, _ := cmd.Flags().GetInt("port")
	host, _ := cmd.Flags().GetString("host")
	webuiAddr, _ := cmd.Flags().GetString("webui-addr")
	noWebUI, _ := cmd.Flags().GetBool("no-webui")

	fmt.Printf("\n\033[35m  NEXUS AI v1.8 — Autonomous OS\033[0m\n")
	fmt.Printf("  Gateway : %s:%d\n", host, port)
	if !noWebUI {
		fmt.Printf("  Web UI  : http://localhost%s\n", webuiAddr)
	}
	fmt.Printf("  Skills  : run 'nexus skills list' to see all plugins\n")
	fmt.Println()

	// 1. Initialize LLM Base Router (v1.7)
	llmCfg := types.LLMConfig{
		Provider:   getEnvOrDefault("NEXUS_LLM_PROVIDER", "ollama"),
		Model:      getEnvOrDefault("NEXUS_LLM_MODEL", "llama3.2"),
		BaseURL:    getEnvOrDefault("NEXUS_LLM_BASE_URL", "http://localhost:11434/v1"),
		APIKey:     os.Getenv("NEXUS_LLM_API_KEY"),
		TimeoutSec: 120,
	}
	r := router.New(llmCfg)
	
	// Create context for daemon lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 2. Initialize Token Stock Market (Dynamic Routing)
	market := routing.NewMarket(60 * time.Second)
	market.Start(ctx)
	defer market.Stop()

	// 3. Initialize Hive-Mind Mesh P2P
	localNode := &mesh.Node{
		ID:      fmt.Sprintf("nexus-%s", os.Getenv("USER")),
		Address: fmt.Sprintf("%s:%d", host, port),
		Profile: mesh.HardwareProfile{
			HasGPU: os.Getenv("NEXUS_HAS_GPU") == "true",
		},
	}
	meshNet := mesh.NewNetwork(localNode, nil) // Transport client injection pending
	discovery := mesh.NewDiscovery(meshNet, localNode)
	
	// Start mDNS broadcast (errors logged but non-fatal if offline)
	if err := discovery.Start(ctx, port); err != nil {
		log.Warn().Err(err).Msg("Mesh discovery disabled (mDNS failed)")
	} else {
		defer discovery.Stop()
	}

	// 4. Initialize the Web UI
	if !noWebUI {
		srv := webui.New(webuiAddr, log.Logger, r)
		go func() {
			log.Info().Str("addr", webuiAddr).Msg("Web UI started — open http://localhost" + webuiAddr)
			if err := srv.Start(); err != nil {
				log.Error().Err(err).Msg("webui error")
			}
		}()
	}

	log.Info().Int("port", port).Str("host", host).Msg("NEXUS daemon running — Ctrl+C to stop")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Info().Msg("NEXUS shutting down gracefully")
	return nil
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
