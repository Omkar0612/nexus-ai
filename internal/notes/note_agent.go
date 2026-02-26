package notes

/*
NoteAgent — auto-capture and organise meeting notes.

NEXUS NoteAgent:
  1. Create notes from text, voice transcripts, or clipboard
  2. Auto-extract action items (lines starting with TODO/ACTION/☐)
  3. Auto-tag by topic using keyword detection
  4. Full-text search across all notes
  5. Export to Markdown
  6. Link notes to goals and audit log
  7. Daily note auto-created on first access
*/

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Note is a single captured note
type Note struct {
	ID          string
	Title       string
	Content     string
	Tags        []string
	ActionItems []string
	LinkedGoal  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Pinned      bool
	Archived    bool
}

// NoteAgent manages note capture and organisation
type NoteAgent struct {
	mu    sync.RWMutex
	notes map[string]*Note
}

// New creates a NoteAgent
func New() *NoteAgent {
	return &NoteAgent{notes: make(map[string]*Note)}
}

// Create adds a new note and returns it
func (n *NoteAgent) Create(title, content string, tags []string) *Note {
	now := time.Now()
	note := &Note{
		ID:          fmt.Sprintf("note-%d", now.UnixNano()),
		Title:       title,
		Content:     content,
		Tags:        autoTag(content, tags),
		ActionItems: extractActionItems(content),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	n.mu.Lock()
	n.notes[note.ID] = note
	n.mu.Unlock()
	return note
}

// Update modifies an existing note
func (n *NoteAgent) Update(id, content string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	note, ok := n.notes[id]
	if !ok {
		return fmt.Errorf("note %s not found", id)
	}
	note.Content = content
	note.UpdatedAt = time.Now()
	note.ActionItems = extractActionItems(content)
	note.Tags = autoTag(content, note.Tags)
	return nil
}

// Search returns notes matching a query string
func (n *NoteAgent) Search(query string) []*Note {
	query = strings.ToLower(query)
	n.mu.RLock()
	defer n.mu.RUnlock()
	var results []*Note
	for _, note := range n.notes {
		if note.Archived {
			continue
		}
		if strings.Contains(strings.ToLower(note.Title), query) ||
			strings.Contains(strings.ToLower(note.Content), query) {
			results = append(results, note)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].UpdatedAt.After(results[j].UpdatedAt)
	})
	return results
}

// GetOrCreateDaily returns today's daily note, creating it if needed
func (n *NoteAgent) GetOrCreateDaily() *Note {
	title := "Daily Note — " + time.Now().Format("Jan 2, 2006")
	results := n.Search(title)
	if len(results) > 0 {
		return results[0]
	}
	return n.Create(title, "", []string{"daily"})
}

// ExportMarkdown exports a note as a Markdown string
func ExportMarkdown(note *Note) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n", note.Title))
	sb.WriteString(fmt.Sprintf("*Created: %s | Updated: %s*\n",
		note.CreatedAt.Format("Jan 2 15:04"),
		note.UpdatedAt.Format("Jan 2 15:04")))
	if len(note.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("Tags: `%s`\n", strings.Join(note.Tags, "` `")))
	}
	sb.WriteString("\n" + note.Content + "\n")
	if len(note.ActionItems) > 0 {
		sb.WriteString("\n## Action Items\n")
		for _, item := range note.ActionItems {
			sb.WriteString(fmt.Sprintf("- [ ] %s\n", item))
		}
	}
	return sb.String()
}

// Stats returns a summary of the note collection
func (n *NoteAgent) Stats() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	total := len(n.notes)
	actions := 0
	for _, note := range n.notes {
		actions += len(note.ActionItems)
	}
	return fmt.Sprintf("📝 Notes: %d | Action items: %d", total, actions)
}

func extractActionItems(content string) []string {
	var items []string
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		prefixes := []string{"TODO:", "ACTION:", "- [ ]", "☐", "* [ ]"}
		for _, prefix := range prefixes {
			if strings.HasPrefix(strings.ToUpper(trimmed), strings.ToUpper(prefix)) {
				item := strings.TrimSpace(trimmed[len(prefix):])
				if item != "" {
					items = append(items, item)
				}
				break
			}
		}
	}
	return items
}

func autoTag(content string, existingTags []string) []string {
	tagMap := map[string][]string{
		"meeting":  {"meeting", "standup", "sync", "call", "zoom", "teams"},
		"code":     {"function", "bug", "deploy", "commit", "PR", "pull request"},
		"finance":  {"invoice", "payment", "budget", "cost", "expense"},
		"research": {"research", "paper", "article", "study", "analysis"},
		"client":   {"client", "customer", "prospect", "proposal"},
	}
	tagSet := make(map[string]bool)
	for _, t := range existingTags {
		tagSet[t] = true
	}
	lower := strings.ToLower(content)
	for tag, keywords := range tagMap {
		for _, kw := range keywords {
			if strings.Contains(lower, strings.ToLower(kw)) {
				tagSet[tag] = true
				break
			}
		}
	}
	result := make([]string, 0, len(tagSet))
	for t := range tagSet {
		result = append(result, t)
	}
	sort.Strings(result)
	return result
}
