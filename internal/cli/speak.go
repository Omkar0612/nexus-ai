package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/Omkar0612/nexus-ai/internal/tts"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var speakCmd = &cobra.Command{
	Use:   "speak <text>",
	Short: "Synthesise speech from text",
	Long: `Convert text to speech using local Coqui TTS, ElevenLabs (free tier), or system TTS.

Examples:
  nexus speak "Good morning, your briefing is ready"
  nexus speak --backend coqui --out briefing.wav "3 tasks today"
  nexus speak --backend elevenlabs "Meeting in 10 minutes"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runSpeak,
}

func init() {
	speakCmd.Flags().String("backend", "system", "Backend: system | coqui | elevenlabs")
	speakCmd.Flags().String("out", "", "Output audio file path (WAV/MP3)")
	speakCmd.Flags().String("voice", "", "Voice ID or name (provider-specific)")
	speakCmd.Flags().String("coqui-url", "http://localhost:5002", "Coqui TTS server URL")
	speakCmd.Flags().String("api-key", "", "API key for ElevenLabs")
}

func runSpeak(cmd *cobra.Command, args []string) error {
	text := strings.Join(args, " ")

	backend, _ := cmd.Flags().GetString("backend")
	out, _ := cmd.Flags().GetString("out")
	voice, _ := cmd.Flags().GetString("voice")
	coquiURL, _ := cmd.Flags().GetString("coqui-url")
	apiKey, _ := cmd.Flags().GetString("api-key")

	if apiKey == "" {
		apiKey = os.Getenv("NEXUS_ELEVENLABS_KEY")
	}

	var opts []tts.Option
	switch backend {
	case "system":
		opts = append(opts, tts.WithSystem())
	case "coqui":
		opts = append(opts, tts.WithCoqui(coquiURL))
	case "elevenlabs":
		if apiKey == "" {
			return fmt.Errorf("elevenlabs backend requires --api-key or NEXUS_ELEVENLABS_KEY env var")
		}
		opts = append(opts, tts.WithElevenLabs(apiKey, voice))
	default:
		return fmt.Errorf("unknown backend %q — choose: system, coqui, elevenlabs", backend)
	}

	agent := tts.New(opts...)
	req := tts.Request{
		Text:       text,
		Voice:      voice,
		OutputPath: out,
	}

	log.Info().Str("backend", backend).Str("text", text).Msg("Speaking...")
	result, err := agent.Speak(cmd.Context(), req)
	if err != nil {
		return fmt.Errorf("speak: %w", err)
	}

	if result.Path != "" {
		fmt.Printf("\n\033[32m✅ Audio saved:\033[0m %s\n", result.Path)
	} else {
		fmt.Printf("\n\033[32m✅ Spoken via %s TTS\033[0m\n", result.Backend)
	}
	fmt.Printf("   Latency : %s\n", result.Latency)
	return nil
}
