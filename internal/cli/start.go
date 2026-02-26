package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the NEXUS daemon",
	Long:  `Starts the NEXUS background daemon with all agents, memory, and gateway.`,
	RunE:  runStart,
}

func init() {
	startCmd.Flags().IntP("port", "p", 7700, "Gateway port")
	startCmd.Flags().String("host", "127.0.0.1", "Bind host")
	startCmd.Flags().BoolP("no-tui", "n", false, "Disable terminal UI")
}

func runStart(cmd *cobra.Command, args []string) error {
	debug, _ := cmd.Flags().GetBool("debug")
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	port, _ := cmd.Flags().GetInt("port")
	host, _ := cmd.Flags().GetString("host")

	fmt.Printf("\n🧠  NEXUS AI Agent\n")
	fmt.Printf("   Starting on %s:%d\n\n", host, port)

	log.Info().Int("port", port).Str("host", host).Msg("NEXUS daemon starting")

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("NEXUS shutting down gracefully")
	return nil
}
