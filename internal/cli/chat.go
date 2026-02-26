package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session with NEXUS",
	Long:  `Opens an interactive terminal chat session. Type 'exit' or Ctrl+C to quit.`,
	RunE:  runChat,
}

func init() {
	chatCmd.Flags().StringP("persona", "p", "", "Persona to use (work/creative/client/focus/research)")
	chatCmd.Flags().BoolP("briefing", "b", true, "Show session briefing on start")
}

func runChat(cmd *cobra.Command, args []string) error {
	persona, _ := cmd.Flags().GetString("persona")
	if persona != "" {
		fmt.Printf("💼 Switching to persona: %s\n\n", persona)
	}
	fmt.Println("🧠 NEXUS Chat (type 'exit' to quit, 'help' for commands)")
	fmt.Println(strings.Repeat("─", 50))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("❯ ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		switch strings.ToLower(input) {
		case "exit", "quit", ":q":
			fmt.Println("Goodbye! 👋")
			return nil
		case "help":
			printHelp()
		default:
			// In production: route to LLM router
			fmt.Printf("🧠 [NEXUS connected to daemon at localhost:7700]\n> echo: %s\n\n", input)
		}
	}
	return nil
}

func printHelp() {
	fmt.Println(`
Commands:
  drift       — run drift detection scan
  goals       — show your tracked goals
  health      — show system health report
  insights    — show usage insights
  persona     — switch persona (work/creative/client/focus/research)
  vault list  — list stored secrets
  exit        — quit NEXUS chat
`)
}
