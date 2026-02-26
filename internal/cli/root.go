package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nexus",
	Short: "NEXUS — The open-source AI agent that learns, heals, and adapts",
	Long: `
NEXUS is a free, self-hosted AI agent with features nobody else has built:
  • Drift Detector   — spots stalled tasks before you lose them
  • Self-Healing      — auto-fixes broken workflows
  • Emotional AI      — adapts tone to your mood
  • Goal Tracker      — keeps your big picture in focus
  • Privacy Vault     — AES-256 encrypted secrets
  • Offline Mode      — works without internet

Get started: nexus start`,
}

// Execute is the CLI entrypoint
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(chatCmd)
	rootCmd.PersistentFlags().StringP("config", "c", "", "Config file path (default: ~/.nexus/nexus.toml)")
	rootCmd.PersistentFlags().StringP("user", "u", "default", "User ID for memory isolation")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging")
}
