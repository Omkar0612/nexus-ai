package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Omkar0612/nexus-ai/internal/router"
	"github.com/Omkar0612/nexus-ai/internal/types"
	"github.com/Omkar0612/nexus-ai/internal/webui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the NEXUS daemon",
	Long:  `Starts the NEXUS background daemon with all agents, memory, webui and gateway.`,
	RunE:  runStart,
}

func init() {
	startCmd.Flags().IntP("port", "p", 7700, "Gateway port")
	startCmd.Flags().String("host", "127.0.0.1", "Bind host")
	startCmd.Flags().String("webui-addr", ":7070", "Web UI listen address (e.g. :7070)")
	startCmd.Flags().BoolP("no-tui", "n", false, "Disable terminal UI")
	startCmd.Flags().BoolP("no-webui", "", false, "Disable web UI")
}

func runStart(cmd *cobra.Command, args []string) error {
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

	fmt.Printf("\n\033[35m  NEXUS AI v1.6\033[0m\n")
	fmt.Printf("  Gateway : %s:%d\n", host, port)
	if !noWebUI {
		fmt.Printf("  Web UI  : http://localhost%s\n", webuiAddr)
	}
	fmt.Println()

	// Boot LLM router with default Ollama (user can override via config/env)
	llmCfg := types.LLMConfig{
		Provider:   "ollama",
		Model:      "llama3.2",
		BaseURL:    "http://localhost:11434/v1",
		TimeoutSec: 120,
	}
	r := router.New(llmCfg)
	log.Info().Str("provider", llmCfg.Provider).Str("model", llmCfg.Model).Msg("LLM router ready")

	// Boot Web UI
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

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("NEXUS shutting down gracefully")
	_ = context.Background() // keep context import used for future shutdown hooks
	return nil
}
