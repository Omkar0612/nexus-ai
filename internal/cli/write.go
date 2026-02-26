package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Omkar0612/nexus-ai/internal/router"
	"github.com/Omkar0612/nexus-ai/internal/types"
	"github.com/Omkar0612/nexus-ai/internal/writing"
	"github.com/spf13/cobra"
)

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "AI Writing Studio — draft, rewrite, proofread, translate and more",
	Long: `AI-powered writing tools backed by your local LLM. Zero additional cost.

Subcommands:
  draft       Generate a new piece of writing
  rewrite     Rewrite text in a different style
  summarise   Condense text to a target word count
  proofread   Check and correct grammar, style, clarity
  expand      Expand an outline into full prose
  translate   Translate text to another language`,
}

// -- draft --

var writeDraftCmd = &cobra.Command{
	Use:   "draft",
	Short: "Draft a piece of writing from a topic",
	Example: `  nexus write draft --topic "AI agents in 2026" --style professional --words 500
  nexus write draft --topic "Why Dubai is a tech hub" --style casual --out article.md`,
	RunE: runWriteDraft,
}

func init() {
	writeDraftCmd.Flags().String("topic", "", "Topic to write about (required)")
	writeDraftCmd.Flags().String("style", "professional", "Style: professional | casual | persuasive | academic | creative")
	writeDraftCmd.Flags().Int("words", 300, "Target word count")
	writeDraftCmd.Flags().String("out", "", "Save output to file")
	_ = writeDraftCmd.MarkFlagRequired("topic")
}

func runWriteDraft(cmd *cobra.Command, _ []string) error {
	topic, _ := cmd.Flags().GetString("topic")
	styleStr, _ := cmd.Flags().GetString("style")
	words, _ := cmd.Flags().GetInt("words")
	out, _ := cmd.Flags().GetString("out")
	a, err := newWritingAgent()
	if err != nil {
		return err
	}
	result, err := a.Draft(cmd.Context(), topic, writing.Style(styleStr), words)
	if err != nil {
		return fmt.Errorf("write draft: %w", err)
	}
	return writeOutput(result, out)
}

// -- rewrite --

var writeRewriteCmd = &cobra.Command{
	Use:   "rewrite",
	Short: "Rewrite text in a different style",
	Example: `  nexus write rewrite --style casual --file email-draft.txt
  echo "The utilisation of AI..." | nexus write rewrite --style casual`,
	RunE: runWriteRewrite,
}

func init() {
	writeRewriteCmd.Flags().String("style", "professional", "Target style")
	writeRewriteCmd.Flags().String("file", "", "Input file (omit to read from stdin)")
	writeRewriteCmd.Flags().String("out", "", "Save output to file")
}

func runWriteRewrite(cmd *cobra.Command, _ []string) error {
	styleStr, _ := cmd.Flags().GetString("style")
	file, _ := cmd.Flags().GetString("file")
	out, _ := cmd.Flags().GetString("out")
	text, err := readInput(file)
	if err != nil {
		return err
	}
	a, err := newWritingAgent()
	if err != nil {
		return err
	}
	result, err := a.Rewrite(cmd.Context(), text, writing.Style(styleStr))
	if err != nil {
		return fmt.Errorf("write rewrite: %w", err)
	}
	return writeOutput(result, out)
}

// -- summarise --

var writeSummariseCmd = &cobra.Command{
	Use:   "summarise",
	Short: "Summarise text to a target word count",
	Example: `  nexus write summarise --file meeting-notes.txt --words 100`,
	RunE:    runWriteSummarise,
}

func init() {
	writeSummariseCmd.Flags().String("file", "", "Input file")
	writeSummariseCmd.Flags().Int("words", 100, "Target word count")
	writeSummariseCmd.Flags().String("out", "", "Save output to file")
}

func runWriteSummarise(cmd *cobra.Command, _ []string) error {
	file, _ := cmd.Flags().GetString("file")
	words, _ := cmd.Flags().GetInt("words")
	out, _ := cmd.Flags().GetString("out")
	text, err := readInput(file)
	if err != nil {
		return err
	}
	a, err := newWritingAgent()
	if err != nil {
		return err
	}
	result, err := a.Summarise(cmd.Context(), text, words)
	if err != nil {
		return fmt.Errorf("write summarise: %w", err)
	}
	return writeOutput(result, out)
}

// -- proofread --

var writeProofreadCmd = &cobra.Command{
	Use:   "proofread",
	Short: "Proofread and correct text",
	Example: `  nexus write proofread --file report.md
  echo "She dont know." | nexus write proofread`,
	RunE: runWriteProofread,
}

func init() {
	writeProofreadCmd.Flags().String("file", "", "Input file")
	writeProofreadCmd.Flags().String("out", "", "Save corrected text to file")
}

func runWriteProofread(cmd *cobra.Command, _ []string) error {
	file, _ := cmd.Flags().GetString("file")
	out, _ := cmd.Flags().GetString("out")
	text, err := readInput(file)
	if err != nil {
		return err
	}
	a, err := newWritingAgent()
	if err != nil {
		return err
	}
	corrected, issues, err := a.Proofread(cmd.Context(), text)
	if err != nil {
		return fmt.Errorf("write proofread: %w", err)
	}
	if len(issues) > 0 {
		fmt.Fprintf(os.Stderr, "\n\033[33mIssues found:\033[0m\n")
		for _, issue := range issues {
			fmt.Fprintf(os.Stderr, "  • %s\n", issue)
		}
		fmt.Fprintln(os.Stderr)
	}
	return writeOutput(corrected, out)
}

// -- expand --

var writeExpandCmd = &cobra.Command{
	Use:   "expand",
	Short: "Expand an outline or bullet list into full prose",
	RunE:  runWriteExpand,
}

func init() {
	writeExpandCmd.Flags().String("file", "", "Input file containing outline")
	writeExpandCmd.Flags().String("style", "professional", "Writing style")
	writeExpandCmd.Flags().String("out", "", "Save output to file")
}

func runWriteExpand(cmd *cobra.Command, _ []string) error {
	file, _ := cmd.Flags().GetString("file")
	styleStr, _ := cmd.Flags().GetString("style")
	out, _ := cmd.Flags().GetString("out")
	outline, err := readInput(file)
	if err != nil {
		return err
	}
	a, err := newWritingAgent()
	if err != nil {
		return err
	}
	result, err := a.Expand(cmd.Context(), outline, writing.Style(styleStr))
	if err != nil {
		return fmt.Errorf("write expand: %w", err)
	}
	return writeOutput(result, out)
}

// -- translate --

var writeTranslateCmd = &cobra.Command{
	Use:   "translate",
	Short: "Translate text to another language",
	Example: `  nexus write translate --lang Arabic --file announcement.txt
  nexus write translate --lang Spanish --file article.md --out article-es.md`,
	RunE: runWriteTranslate,
}

func init() {
	writeTranslateCmd.Flags().String("file", "", "Input file")
	writeTranslateCmd.Flags().String("lang", "", "Target language (required)")
	writeTranslateCmd.Flags().String("out", "", "Save translated text to file")
	_ = writeTranslateCmd.MarkFlagRequired("lang")
}

func runWriteTranslate(cmd *cobra.Command, _ []string) error {
	file, _ := cmd.Flags().GetString("file")
	lang, _ := cmd.Flags().GetString("lang")
	out, _ := cmd.Flags().GetString("out")
	text, err := readInput(file)
	if err != nil {
		return err
	}
	a, err := newWritingAgent()
	if err != nil {
		return err
	}
	result, err := a.Translate(cmd.Context(), text, lang)
	if err != nil {
		return fmt.Errorf("write translate: %w", err)
	}
	return writeOutput(result, out)
}

// -- shared helpers --

// newWritingAgent builds a writing agent from environment variables.
// Supported env vars: NEXUS_LLM_PROVIDER, NEXUS_LLM_MODEL, NEXUS_LLM_BASE_URL, NEXUS_LLM_API_KEY.
func newWritingAgent() (*writing.Agent, error) {
	r := router.New(types.LLMConfig{
		Provider:   getEnvOrDefault("NEXUS_LLM_PROVIDER", "ollama"),
		Model:      getEnvOrDefault("NEXUS_LLM_MODEL", "llama3.2"),
		BaseURL:    getEnvOrDefault("NEXUS_LLM_BASE_URL", "http://localhost:11434/v1"),
		APIKey:     os.Getenv("NEXUS_LLM_API_KEY"),
		TimeoutSec: 120,
	})
	return writing.New(r), nil
}

// readInput reads text from a file or stdin.
// Uses os.Stdin directly — works on all platforms including Windows.
func readInput(file string) (string, error) {
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read file %q: %w", file, err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// writeOutput prints text to stdout or saves it to a file.
func writeOutput(text, path string) error {
	if path == "" {
		fmt.Println(text)
		return nil
	}
	if err := os.WriteFile(path, []byte(text+"\n"), 0o644); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	fmt.Printf("\033[32m✅ Saved:\033[0m %s\n", path)
	return nil
}

func init() {
	writeCmd.AddCommand(writeDraftCmd)
	writeCmd.AddCommand(writeRewriteCmd)
	writeCmd.AddCommand(writeSummariseCmd)
	writeCmd.AddCommand(writeProofreadCmd)
	writeCmd.AddCommand(writeExpandCmd)
	writeCmd.AddCommand(writeTranslateCmd)
}
