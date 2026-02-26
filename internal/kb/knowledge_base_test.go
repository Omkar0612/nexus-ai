package kb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKBAddAndSearch(t *testing.T) {
	dir := t.TempDir()
	kbase, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	kbase.AddText("doc1", "NEXUS Features",
		"NEXUS has a drift detector that finds stalled tasks. It also has a self-healing agent.",
		[]string{"features"},
	)
	kbase.AddText("doc2", "Vault Security",
		"The privacy vault uses AES-256-GCM encryption to store API keys and secrets.",
		[]string{"security"},
	)

	results := kbase.Search("drift detector stalled tasks", 3)
	if len(results) == 0 {
		t.Fatal("expected at least 1 search result")
	}
	if results[0].DocTitle != "NEXUS Features" {
		t.Errorf("expected NEXUS Features as top result, got %s", results[0].DocTitle)
	}
}

func TestKBBuildContext(t *testing.T) {
	dir := t.TempDir()
	kbase, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	kbase.AddText("doc1", "Architecture",
		"NEXUS uses a modular architecture with a router, memory store, and agent bus.",
		nil,
	)

	ctx := kbase.BuildContext("router architecture", 3, 2000)
	if ctx == "" {
		t.Error("expected non-empty context")
	}
	if len(ctx) < 10 {
		t.Error("context too short")
	}
}

func TestKBIndexFile(t *testing.T) {
	dir := t.TempDir()
	kbase, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	testFile := filepath.Join(dir, "notes.md")
	if err := os.WriteFile(testFile, []byte("# NEXUS Notes\n\nThe goal tracker helps you stay on target."), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := kbase.IndexFile(testFile); err != nil {
		t.Fatalf("IndexFile: %v", err)
	}

	results := kbase.Search("goal tracker", 3)
	if len(results) == 0 {
		t.Fatal("expected result after indexing file")
	}
}

func TestKBStats(t *testing.T) {
	dir := t.TempDir()
	kbase, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	stats := kbase.Stats()
	if stats == "" {
		t.Error("expected non-empty stats")
	}

	kbase.AddText("d1", "Test", "some content here for testing", nil)
	stats = kbase.Stats()
	if stats == "" {
		t.Error("expected stats after adding document")
	}
}
