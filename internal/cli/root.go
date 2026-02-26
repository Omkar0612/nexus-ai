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
NEXUS AI v1.7 — Free forever. Self-hosted. Production ready.

Core capabilities:
  nexus start      — Start the daemon (Web UI at :7070)
  nexus chat       — Interactive chat with your local LLM

Creative Studio (v1.7):
  nexus imagine    — AI image generation (Stable Diffusion / Together AI FLUX)
  nexus speak      — Voice synthesis (Coqui / ElevenLabs / System TTS)
  nexus write      — Writing studio (draft, rewrite, proofread, translate)
  nexus music      — Music generation (AudioCraft / Replicate MusicGen)

Agents:
  nexus calendar   — Calendar agent (today, week, conflicts, free slots)
  nexus skills     — Plugin registry (list, run)

Run 'nexus <command> --help' for details on each command.`,
}

// Execute is the CLI entrypoint.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Core
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(chatCmd)

	// v1.7 Creative Studio
	rootCmd.AddCommand(imagineCmd)
	rootCmd.AddCommand(speakCmd)
	rootCmd.AddCommand(writeCmd)
	rootCmd.AddCommand(musicCmd)

	// Agents
	rootCmd.AddCommand(calendarCmd)
	rootCmd.AddCommand(skillsCmd)

	// Global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "Config file (default: ~/.nexus/nexus.toml)")
	rootCmd.PersistentFlags().StringP("user", "u", "default", "User ID for memory isolation")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging")
}
