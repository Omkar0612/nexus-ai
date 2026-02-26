package memory

import (
	"os"
	"testing"
)

func TestMemoryStoreAddAndRetrieve(t *testing.T) {
	dir := t.TempDir()
	s, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer s.Close()

	if err := s.Add("user1", "user", "working on nexus AI features", "episodic", 0.8); err != nil {
		t.Fatalf("Add: %v", err)
	}

	mems, err := s.GetEpisodicHistory("user1", 10)
	if err != nil {
		t.Fatalf("GetEpisodicHistory: %v", err)
	}
	if len(mems) != 1 {
		t.Fatalf("expected 1 memory, got %d", len(mems))
	}
	if mems[0].Content != "working on nexus AI features" {
		t.Errorf("unexpected content: %s", mems[0].Content)
	}
}

func TestMemorySearch(t *testing.T) {
	dir := t.TempDir()
	s, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer s.Close()

	_ = s.Add("user1", "user", "building a drift detector", "episodic", 0.9)
	_ = s.Add("user1", "user", "fixing the vault encryption", "episodic", 0.7)

	results, err := s.Search("user1", "drift", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 search result, got %d", len(results))
	}
}

func TestMemoryCount(t *testing.T) {
	dir := t.TempDir()
	s, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer s.Close()

	for i := 0; i < 5; i++ {
		_ = s.Add("user1", "user", "test entry", "episodic", 0.5)
	}

	n, err := s.Count("user1")
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5, got %d", n)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
