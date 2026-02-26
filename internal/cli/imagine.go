package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/Omkar0612/nexus-ai/internal/imagegen"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var imagineCmd = &cobra.Command{
	Use:   "imagine <prompt>",
	Short: "Generate an image from a text prompt",
	Long: `Generate images using local Stable Diffusion (A1111) or Together AI FLUX.1-schnell (free credits).

Examples:
  nexus imagine "a futuristic Dubai skyline at sunset"
  nexus imagine --backend together "minimalist logo, purple gradient"
  nexus imagine --output ./cover.png --width 1024 --height 768 "epic mountain landscape"`,
	Args: cobra.MinimumNArgs(1),
	RunE: runImagine,
}

func init() {
	imagineCmd.Flags().String("backend", "stablediffusion", "Backend: stablediffusion | together | replicate")
	imagineCmd.Flags().String("output", "", "Output PNG file path (default: temp file)")
	imagineCmd.Flags().Int("width", 512, "Image width in pixels")
	imagineCmd.Flags().Int("height", 512, "Image height in pixels")
	imagineCmd.Flags().Int("steps", 20, "Diffusion steps (more = better quality, slower)")
	imagineCmd.Flags().String("negative", "", "Negative prompt")
	imagineCmd.Flags().String("sd-url", "http://127.0.0.1:7860", "Stable Diffusion API URL")
	imagineCmd.Flags().String("api-key", "", "API key for Together/Replicate backends")
	imagineCmd.Flags().String("model", "black-forest-labs/FLUX.1-schnell-Free", "Model name (Together/Replicate)")
}

func runImagine(cmd *cobra.Command, args []string) error {
	prompt := strings.Join(args, " ")

	backend, _ := cmd.Flags().GetString("backend")
	output, _ := cmd.Flags().GetString("output")
	width, _ := cmd.Flags().GetInt("width")
	height, _ := cmd.Flags().GetInt("height")
	steps, _ := cmd.Flags().GetInt("steps")
	negative, _ := cmd.Flags().GetString("negative")
	sdURL, _ := cmd.Flags().GetString("sd-url")
	apiKey, _ := cmd.Flags().GetString("api-key")
	model, _ := cmd.Flags().GetString("model")

	if apiKey == "" {
		apiKey = os.Getenv("NEXUS_TOGETHER_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("NEXUS_REPLICATE_KEY")
	}

	var opts []imagegen.Option
	switch backend {
	case "stablediffusion", "sd":
		opts = append(opts, imagegen.WithStableDiffusion(sdURL))
	case "together":
		if apiKey == "" {
			return fmt.Errorf("together backend requires --api-key or NEXUS_TOGETHER_KEY env var")
		}
		opts = append(opts, imagegen.WithTogether(apiKey, model))
	case "replicate":
		if apiKey == "" {
			return fmt.Errorf("replicate backend requires --api-key or NEXUS_REPLICATE_KEY env var")
		}
		opts = append(opts, imagegen.WithReplicate(apiKey))
	default:
		return fmt.Errorf("unknown backend %q — choose: stablediffusion, together, replicate", backend)
	}

	agent := imagegen.New(opts...)
	log.Debug().Str("backend", backend).Str("prompt", prompt).Msg("Generating image...")
	result, err := agent.Generate(cmd.Context(), imagegen.Request{
		Prompt:         prompt,
		NegativePrompt: negative,
		Width:          width,
		Height:         height,
		Steps:          steps,
		OutputPath:     output,
	})
	if err != nil {
		return fmt.Errorf("imagine: %w", err)
	}

	fmt.Printf("\n\033[32m✅ Image saved:\033[0m %s\n", result.Path)
	fmt.Printf("   Backend : %s\n", result.Backend)
	fmt.Printf("   Latency : %s\n", result.Latency)
	return nil
}
