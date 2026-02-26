package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/Omkar0612/nexus-ai/internal/plugin"
	"github.com/spf13/cobra"
)

// globalRegistry is the shared plugin registry for the CLI.
// In production this would be populated from loaded plugin binaries.
var globalRegistry = plugin.NewRegistry()

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage and run NEXUS plugins (skills)",
	Long: `List and run registered NEXUS skills.

Subcommands:
  list    Show all registered skills
  run     Run a named skill`,
}

var skillsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered skills",
	RunE:  runSkillsList,
}

var skillsRunCmd = &cobra.Command{
	Use:   "run <skill-name> [key=value ...]",
	Short: "Run a skill with optional key=value arguments",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSkillsRun,
}

func init() {
	skillsCmd.AddCommand(skillsListCmd)
	skillsCmd.AddCommand(skillsRunCmd)

	// Register built-in demo skills
	_ = globalRegistry.Register(plugin.NewSkill(
		"echo",
		"Echoes the command back (built-in demo)",
		func(in plugin.Input) plugin.Output {
			return plugin.Output{Text: "ECHO: " + in.Command}
		},
	))
	_ = globalRegistry.Register(plugin.NewSkill(
		"ping",
		"Responds with pong (built-in demo)",
		func(_ plugin.Input) plugin.Output {
			return plugin.Output{Text: "pong"}
		},
	))
}

func runSkillsList(_ *cobra.Command, _ []string) error {
	entries := globalRegistry.List()
	if len(entries) == 0 {
		fmt.Println("No skills registered. See CONTRIBUTING.md to build your own.")
		return nil
	}
	fmt.Println("\n\033[35mRegistered Skills\033[0m")
	fmt.Println(strings.Repeat("-", 50))
	for _, entry := range entries {
		fmt.Println(" ", entry)
	}
	return nil
}

func runSkillsRun(_ *cobra.Command, args []string) error {
	name := args[0]
	kvArgs := make(map[string]string)
	for _, kv := range args[1:] {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			kvArgs[parts[0]] = parts[1]
		}
	}

	out, err := globalRegistry.Execute(name, plugin.Input{
		Command: name,
		Args:    kvArgs,
		Context: context.Background(),
	})
	if err != nil {
		return fmt.Errorf("skills run: %w", err)
	}
	if out.Error != nil {
		return fmt.Errorf("skill %q error: %w", name, out.Error)
	}
	if out.Text != "" {
		fmt.Println(out.Text)
	}
	return nil
}
