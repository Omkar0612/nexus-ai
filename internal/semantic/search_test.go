package semantic

import (
	"testing"
)

func TestSemanticSearch(t *testing.T) {
	idx := NewIndex()
	idx.Add(Document{ID: "1", Text: "NEXUS AI agent self-healing drift detection"})
	idx.Add(Document{ID: "2", Text: "calendar scheduling meetings time management"})
	idx.Add(Document{ID: "3", Text: "vision image recognition object detection"})
	idx.Add(Document{ID: "4", Text: "voice speech recognition transcription"})
	idx.Rebuild()

	results := idx.Search("detect drift agent", 2)
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if results[0].Document.ID != "1" {
		t.Errorf("expected doc 1 first, got %s (score %.4f)", results[0].Document.ID, results[0].Score)
	}
}

func TestSearchEmpty(t *testing.T) {
	idx := NewIndex()
	idx.Rebuild()
	results := idx.Search("anything", 5)
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
