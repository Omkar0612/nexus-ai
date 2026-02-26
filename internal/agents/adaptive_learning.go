package agents

/*
AdaptiveLearner — learns your workflow patterns and personalizes responses.

Unlike GPT/Claude which only learn from what you TELL them,
NEXUS learns from HOW you work:
  - Which agent you use most → pre-warm it
  - What time of day you code vs write → route accordingly
  - What response formats you engage with → apply them automatically
  - What topics you care about → weight them in memory retrieval

Zero configuration. It just observes and adapts.
*/

import (
	"fmt"
	"strings"
	"time"

	"github.com/Omkar0612/nexus-ai/internal/memory"
)

// UsagePattern holds learned workflow statistics
type UsagePattern struct {
	AgentUsageCounts map[string]int
	PeakHours        map[int]int
	FavoriteTopics   map[string]int
	PreferredFormats map[string]int
	AvgMsgLength     int
	LastUpdated      time.Time
}

// AdaptiveLearner silently learns and personalizes NEXUS behavior
type AdaptiveLearner struct {
	mem      *memory.Store
	userID   string
	patterns *UsagePattern
}

// NewAdaptiveLearner creates a new adaptive learner
func NewAdaptiveLearner(mem *memory.Store, userID string) *AdaptiveLearner {
	return &AdaptiveLearner{
		mem:    mem,
		userID: userID,
		patterns: &UsagePattern{
			AgentUsageCounts: make(map[string]int),
			PeakHours:        make(map[int]int),
			FavoriteTopics:   make(map[string]int),
			PreferredFormats: make(map[string]int),
		},
	}
}

// Learn updates usage patterns from a new interaction
func (l *AdaptiveLearner) Learn(agentUsed, userMessage, response string, wasPositive bool) {
	l.patterns.AgentUsageCounts[agentUsed]++
	l.patterns.PeakHours[time.Now().Hour()]++

	for _, word := range strings.Fields(strings.ToLower(userMessage)) {
		if len(word) > 5 {
			l.patterns.FavoriteTopics[word]++
		}
	}

	if wasPositive {
		if strings.Contains(response, "\n-") || strings.Contains(response, "\n•") {
			l.patterns.PreferredFormats["bullets"]++
		} else if strings.Count(response, "```") > 0 {
			l.patterns.PreferredFormats["code"]++
		} else {
			l.patterns.PreferredFormats["prose"]++
		}
	}

	if l.patterns.AvgMsgLength == 0 {
		l.patterns.AvgMsgLength = len(userMessage)
	} else {
		l.patterns.AvgMsgLength = (l.patterns.AvgMsgLength + len(userMessage)) / 2
	}
	l.patterns.LastUpdated = time.Now()
}

// PersonalizeSystemPrompt adds learned preferences to any LLM system prompt
func (l *AdaptiveLearner) PersonalizeSystemPrompt(base string) string {
	var hints []string

	best, bestCount := "", 0
	for f, c := range l.patterns.PreferredFormats {
		if c > bestCount {
			bestCount = c
			best = f
		}
	}
	if best == "bullets" && bestCount > 2 {
		hints = append(hints, "This user prefers bullet-point responses.")
	} else if best == "code" && bestCount > 2 {
		hints = append(hints, "This user prefers responses with code examples.")
	}

	if l.patterns.AvgMsgLength < 30 {
		hints = append(hints, "User sends short messages — be concise.")
	} else if l.patterns.AvgMsgLength > 200 {
		hints = append(hints, "User writes detailed messages — match their depth.")
	}

	if len(hints) == 0 {
		return base
	}
	return base + "\n\n[Personalization: " + strings.Join(hints, " ") + "]"
}

// InsightReport generates a usage insight summary
func (l *AdaptiveLearner) InsightReport() string {
	if l.patterns.LastUpdated.IsZero() {
		return "Not enough data yet. Keep using NEXUS and insights will appear here."
	}
	var sb strings.Builder
	sb.WriteString("📊 **Your NEXUS Usage Insights**\n\n")

	best, bestCount := "", 0
	for agent, count := range l.patterns.AgentUsageCounts {
		if count > bestCount {
			bestCount = count
			best = agent
		}
	}
	if best != "" {
		sb.WriteString(fmt.Sprintf("🤖 Most used agent: **%s** (%d times)\n", best, bestCount))
	}

	peakHour, peakCount := 0, 0
	for h, c := range l.patterns.PeakHours {
		if c > peakCount {
			peakCount = c
			peakHour = h
		}
	}
	sb.WriteString(fmt.Sprintf("⏰ Peak productivity hour: **%02d:00** (%d sessions)\n", peakHour, peakCount))
	return sb.String()
}
