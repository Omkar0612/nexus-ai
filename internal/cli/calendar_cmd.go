package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Omkar0612/nexus-ai/internal/calendar"
	"github.com/spf13/cobra"
)

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Calendar agent — schedule, conflicts, free slots",
	Long: `Interact with your calendar using natural language.

Subcommands:
  today      List today's events
  tomorrow   List tomorrow's events
  week       List this week's events
  conflicts  Detect scheduling conflicts
  free       Find the next free time slot
  digest     Morning digest summary`,
}

var calTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "List today's events",
	RunE:  runCalToday,
}

var calTomorrowCmd = &cobra.Command{
	Use:   "tomorrow",
	Short: "List tomorrow's events",
	RunE:  runCalTomorrow,
}

var calWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "List this week's events",
	RunE:  runCalWeek,
}

var calConflictsCmd = &cobra.Command{
	Use:   "conflicts",
	Short: "Detect scheduling conflicts",
	RunE:  runCalConflicts,
}

var calFreeCmd = &cobra.Command{
	Use:   "free",
	Short: "Find the next available free slot",
	RunE:  runCalFree,
}

var calDigestCmd = &cobra.Command{
	Use:   "digest",
	Short: "Morning digest — summary of today's events",
	RunE:  runCalDigest,
}

func init() {
	calFreeCmd.Flags().Duration("duration", 1*time.Hour, "Required slot duration (e.g. 30m, 1h)")
	calFreeCmd.Flags().Duration("lookahead", 7*24*time.Hour, "Look-ahead window (default: 7 days)")

	calendarCmd.AddCommand(calTodayCmd)
	calendarCmd.AddCommand(calTomorrowCmd)
	calendarCmd.AddCommand(calWeekCmd)
	calendarCmd.AddCommand(calConflictsCmd)
	calendarCmd.AddCommand(calFreeCmd)
	calendarCmd.AddCommand(calDigestCmd)
}

// newCalAgent builds a calendar agent from env vars.
// Set NEXUS_GCAL_TOKEN and optionally NEXUS_GCAL_ID.
func newCalAgent() *calendar.Agent {
	tz := time.Local
	token := os.Getenv("NEXUS_GCAL_TOKEN")
	if token == "" {
		// No provider configured — return agent with no backends
		// (will return empty lists, not an error)
		return calendar.New(tz)
	}
	calID := os.Getenv("NEXUS_GCAL_ID")
	if calID == "" {
		calID = "primary"
	}
	provider := calendar.NewGoogle(token, calID)
	return calendar.New(tz, provider)
}

func runCalToday(_ *cobra.Command, _ []string) error {
	a := newCalAgent()
	events, err := a.Today(context.Background())
	if err != nil {
		return fmt.Errorf("calendar today: %w", err)
	}
	printEvents(events, "Today")
	return nil
}

func runCalTomorrow(_ *cobra.Command, _ []string) error {
	a := newCalAgent()
	events, err := a.Tomorrow(context.Background())
	if err != nil {
		return fmt.Errorf("calendar tomorrow: %w", err)
	}
	printEvents(events, "Tomorrow")
	return nil
}

func runCalWeek(_ *cobra.Command, _ []string) error {
	a := newCalAgent()
	events, err := a.Week(context.Background())
	if err != nil {
		return fmt.Errorf("calendar week: %w", err)
	}
	printEvents(events, "This Week")
	return nil
}

func runCalConflicts(cmd *cobra.Command, _ []string) error {
	a := newCalAgent()
	events, err := a.Week(context.Background())
	if err != nil {
		return fmt.Errorf("calendar conflicts: %w", err)
	}
	conflicts := a.DetectConflicts(events)
	if len(conflicts) == 0 {
		fmt.Println("\033[32m✅ No scheduling conflicts this week.\033[0m")
		return nil
	}
	fmt.Printf("\n\033[33m⚠️  %d conflict(s) detected:\033[0m\n\n", len(conflicts))
	for i, c := range conflicts {
		fmt.Printf("  %d. \033[31m%s\033[0m\n", i+1, c.EventA.Title)
		fmt.Printf("     overlaps with \033[31m%s\033[0m\n", c.EventB.Title)
		fmt.Printf("     Overlap: %s\n\n", c.Overlap)
	}
	return nil
}

func runCalFree(cmd *cobra.Command, _ []string) error {
	duration, _ := cmd.Flags().GetDuration("duration")
	lookahead, _ := cmd.Flags().GetDuration("lookahead")
	a := newCalAgent()
	slot, err := a.FindFreeSlot(context.Background(), duration, lookahead, time.Now())
	if err != nil {
		return fmt.Errorf("calendar free: %w", err)
	}
	fmt.Printf("\n\033[32m✅ Next free slot (%s):\033[0m %s\n",
		duration, slot.Format("Mon 02 Jan 15:04"))
	return nil
}

func runCalDigest(_ *cobra.Command, _ []string) error {
	a := newCalAgent()
	events, err := a.Today(context.Background())
	if err != nil {
		return fmt.Errorf("calendar digest: %w", err)
	}
	lines := calendar.DigestLines(events, time.Local)
	fmt.Println("\n\033[35m📅 Today's Digest\033[0m")
	if len(lines) == 0 {
		fmt.Println("  No events today.")
		return nil
	}
	for _, l := range lines {
		fmt.Println(" ", l)
	}
	return nil
}

func printEvents(events []calendar.Event, label string) {
	fmt.Printf("\n\033[35m📅 %s\033[0m\n", label)
	if len(events) == 0 {
		fmt.Println("  No events.")
		return
	}
	for _, e := range events {
		if e.AllDay {
			fmt.Printf("  📅  %s (all day)\n", e.Title)
		} else {
			fmt.Printf("  🕐  %s — %s  %s\n",
				e.Start.Format("15:04"),
				e.End.Format("15:04"),
				e.Title,
			)
		}
		if e.Location != "" {
			fmt.Printf("       📍 %s\n", e.Location)
		}
		if len(e.Attendees) > 0 {
			fmt.Printf("       👥 %s\n", joinStrings(e.Attendees, ", "))
		}
	}
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
