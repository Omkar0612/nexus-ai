package agents

/*
LoopDetector — detects and breaks infinite agent loops before they burn tokens.

One of the top 4 production failure modes for autonomous agents (arXiv 2025).
Common example: agent calls web-search → summarise → web-search → summarise
literally forever, silently burning $40+ before anyone notices.

NEXUS LoopDetector:
  1. Hashes every (tool, input) call as it happens
  2. Detects identical call repeated >= threshold times (default: 3)
  3. Breaks the loop immediately
  4. Generates a plain-language explanation of WHAT looped and WHY
  5. Suggests a fix (different tool, reformulated query, human review)
  6. Estimates tokens/cost that would have been wasted if not caught

No other open-source agent catches infinite loops proactively.
*/

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"
)

// LoopEvent records a detected loop
type LoopEvent struct {
	Tool        string
	Input       string
	RepeatCount int
	DetectedAt  time.Time
	EstTokens   int    // tokens wasted if loop had continued
	EstCostUSD  float64
	Suggestion  string
}

// CallRecord tracks a single tool invocation
type CallRecord struct {
	Hash      string
	Tool      string
	Input     string
	Count     int
	FirstSeen time.Time
	LastSeen  time.Time
}

// LoopDetector monitors tool calls and fires on detected loops
type LoopDetector struct {
	mu           sync.Mutex
	calls        map[string]*CallRecord // hash -> record
	callOrder    []string               // ordered hashes for recency
	threshold    int                    // repeat count to trigger
	windowSize   int                    // only look at last N calls
	tokenPerCall int                    // estimated tokens per call
	onLoop       func(LoopEvent)
}

// NewLoopDetector creates a loop detector with sensible defaults
func NewLoopDetector(threshold, windowSize int) *LoopDetector {
	if threshold <= 0 {
		threshold = 3
	}
	if windowSize <= 0 {
		windowSize = 20
	}
	return &LoopDetector{
		calls:        make(map[string]*CallRecord),
		threshold:    threshold,
		windowSize:   windowSize,
		tokenPerCall: 500, // conservative estimate
	}
}

// SetLoopCallback sets a function called when a loop is detected
func (d *LoopDetector) SetLoopCallback(fn func(LoopEvent)) {
	d.onLoop = fn
}

// Record registers a tool call and returns (isLoop, event)
func (d *LoopDetector) Record(tool, input string) (bool, *LoopEvent) {
	d.mu.Lock()
	defer d.mu.Unlock()

	h := callHash(tool, input)
	now := time.Now()

	rec, exists := d.calls[h]
	if !exists {
		rec = &CallRecord{
			Hash:      h,
			Tool:      tool,
			Input:     input,
			FirstSeen: now,
		}
		d.calls[h] = rec
	}
	rec.Count++
	rec.LastSeen = now

	// Track call order, keep window
	d.callOrder = append(d.callOrder, h)
	if len(d.callOrder) > d.windowSize {
		oldHash := d.callOrder[0]
		d.callOrder = d.callOrder[1:]
		// Decrement count for evicted call
		if old, ok := d.calls[oldHash]; ok {
			old.Count--
			if old.Count <= 0 {
				delete(d.calls, oldHash)
			}
		}
	}

	if rec.Count >= d.threshold {
		wasted := d.threshold * d.tokenPerCall
		event := LoopEvent{
			Tool:        tool,
			Input:       truncate(input, 100),
			RepeatCount: rec.Count,
			DetectedAt:  now,
			EstTokens:   wasted,
			EstCostUSD:  float64(wasted) / 1_000_000 * 0.59, // Groq pricing
			Suggestion:  d.suggest(tool, input),
		}
		// Reset to prevent firing on every subsequent call
		rec.Count = 0
		if d.onLoop != nil {
			d.onLoop(event)
		}
		return true, &event
	}
	return false, nil
}

// Reset clears all call history (e.g. on new session)
func (d *LoopDetector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.calls = make(map[string]*CallRecord)
	d.callOrder = nil
}

// Stats returns a summary of current call counts
func (d *LoopDetector) Stats() map[string]int {
	d.mu.Lock()
	defer d.mu.Unlock()
	result := make(map[string]int, len(d.calls))
	for _, rec := range d.calls {
		result[rec.Tool+":"+truncate(rec.Input, 30)] = rec.Count
	}
	return result
}

// Format renders a loop event as a user-facing alert message
func (e *LoopEvent) Format() string {
	var sb strings.Builder
	sb.WriteString("🔁 **NEXUS Loop Detected — Stopped**\n\n")
	sb.WriteString(fmt.Sprintf("Tool `%s` was called **%d times** with the same input:\n", e.Tool, e.RepeatCount))
	sb.WriteString(fmt.Sprintf("  `%s`\n\n", e.Input))
	sb.WriteString(fmt.Sprintf("💰 Estimated waste if uncaught: **~%d tokens** (~$%.6f)\n\n", e.EstTokens, e.EstCostUSD))
	if e.Suggestion != "" {
		sb.WriteString(fmt.Sprintf("💡 Suggestion: %s\n", e.Suggestion))
	}
	return sb.String()
}

func (d *LoopDetector) suggest(tool, input string) string {
	toolLower := strings.ToLower(tool)
	switch {
	case strings.Contains(toolLower, "search") || strings.Contains(toolLower, "web"):
		return "Try rephrasing the search query, use a different source, or ask the user to clarify."
	case strings.Contains(toolLower, "file") || strings.Contains(toolLower, "read"):
		return "The file may not exist or be unreadable. Check the path and permissions."
	case strings.Contains(toolLower, "api") || strings.Contains(toolLower, "http"):
		return "The API may be returning an error. Check authentication and rate limits."
	case strings.Contains(toolLower, "memory"):
		return "Memory query returning empty results. Try broader search terms."
	default:
		return "Consider breaking the task into smaller steps or asking the user for guidance."
	}
}

func callHash(tool, input string) string {
	h := sha256.Sum256([]byte(tool + "|" + strings.TrimSpace(input)))
	return fmt.Sprintf("%x", h[:8])
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
