package cli

import (
	"fmt"
	"time"

	"github.com/Omkar0612/nexus-ai/internal/music"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var musicCmd = &cobra.Command{
	Use:   "music <prompt>",
	Short: "Generate music from a text prompt",
	Long: `Generate music using local Meta AudioCraft or Replicate MusicGen (free tier).

Examples:
  nexus music "calm lo-fi piano for focus work" --duration 30s
  nexus music --backend replicate "cinematic orchestral" --duration 15s --out score.wav
  nexus music "upbeat jazz, 120 bpm" --out track.wav`,
	Args:  cobra.MinimumNArgs(1),
	RunE:  runMusic,
}

func init() {
	musicCmd.Flags().String("backend", "stub", "Backend: stub | audiocraft | replicate")
	musicCmd.Flags().Duration("duration", 10*time.Second, "Duration of generated audio (e.g. 30s, 1m)")
	musicCmd.Flags().String("out", "", "Output WAV file path")
	musicCmd.Flags().String("api-key", "", "API key for Replicate")
	musicCmd.Flags().String("audiocraft-url", "http://localhost:8765", "AudioCraft bridge URL")
}

func runMusic(cmd *cobra.Command, args []string) error {
	prompt := ""
	for i, a := range args {
		if i > 0 {
			prompt += " "
		}
		prompt += a
	}

	backend, _ := cmd.Flags().GetString("backend")
	duration, _ := cmd.Flags().GetDuration("duration")
	out, _ := cmd.Flags().GetString("out")
	apiKey, _ := cmd.Flags().GetString("api-key")
	acURL, _ := cmd.Flags().GetString("audiocraft-url")

	var opts []music.Option
	switch backend {
	case "stub":
		// default, no option needed
	case "audiocraft":
		opts = append(opts, music.WithAudioCraft(acURL))
	case "replicate":
		if apiKey == "" {
			return fmt.Errorf("replicate backend requires --api-key")
		}
		opts = append(opts, music.WithReplicate(apiKey))
	default:
		return fmt.Errorf("unknown backend %q — choose: stub, audiocraft, replicate", backend)
	}

	agent := music.New(opts...)
	req := music.Request{
		Prompt:     prompt,
		Duration:   duration,
		OutputPath: out,
	}

	log.Info().Str("backend", backend).Str("prompt", prompt).Dur("duration", duration).Msg("Generating music...")
	result, err := agent.Generate(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("music: %w", err)
	}

	if result.Path != "" {
		fmt.Printf("\n\033[32m✅ Music saved:\033[0m %s\n", result.Path)
	} else {
		fmt.Printf("\n\033[32m✅ Music generation submitted\033[0m\n")
	}
	fmt.Printf("   Backend : %s\n", result.Backend)
	fmt.Printf("   Latency : %s\n", result.Latency)
	return nil
}
