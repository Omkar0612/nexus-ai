package notes

import (
	"testing"
)

func TestNoteCreateAndSearch(t *testing.T) {
	a := New()
	note := a.Create("Meeting Notes", "Discussed the NEXUS roadmap. TODO: write tests.", nil)
	if note.ID == "" {
		t.Error("expected non-empty ID")
	}
	results := a.Search("NEXUS roadmap")
	if len(results) == 0 {
		t.Error("expected search to find the note")
	}
}

func TestNoteActionItemExtraction(t *testing.T) {
	a := New()
	note := a.Create("Sprint Planning",
		"Reviewed backlog.\nTODO: update CI config\nACTION: ping designer\n- [ ] write docs",
		nil)
	if len(note.ActionItems) < 3 {
		t.Errorf("expected 3 action items, got %d: %v", len(note.ActionItems), note.ActionItems)
	}
}

func TestNoteAutoTag(t *testing.T) {
	a := New()
	note := a.Create("Client Call", "Had a zoom call with the client about the invoice.", nil)
	hasClient := false
	hasMeeting := false
	for _, tag := range note.Tags {
		if tag == "client" { hasClient = true }
		if tag == "meeting" { hasMeeting = true }
	}
	if !hasClient || !hasMeeting {
		t.Errorf("expected client+meeting tags, got: %v", note.Tags)
	}
}

func TestNoteDailyNote(t *testing.T) {
	a := New()
	daily1 := a.GetOrCreateDaily()
	daily2 := a.GetOrCreateDaily()
	if daily1.ID != daily2.ID {
		t.Error("expected same daily note on second call")
	}
}

func TestNoteExportMarkdown(t *testing.T) {
	a := New()
	note := a.Create("Test", "Content here.\nTODO: do something", nil)
	md := ExportMarkdown(note)
	if md == "" {
		t.Error("expected non-empty markdown")
	}
	if !containsStr(md, "# Test") {
		t.Error("expected markdown header")
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(s) > 0 && (s[:len(sub)] == sub || containsStr(s[1:], sub))))
}
