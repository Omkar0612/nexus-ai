package memory

import (
	"context"
	"testing"
	"time"
)

type mockDB struct {
	tokens   int
	episodes []Episode
	concepts []*Concept
	deleted  []string
}

func (m *mockDB) GetRecentEpisodes(limit int) ([]Episode, error) { return nil, nil }
func (m *mockDB) GetOldEpisodes(olderThan time.Time) ([]Episode, error) {
	return m.episodes, nil
}
func (m *mockDB) DeleteEpisodes(ids []string) error {
	m.deleted = append(m.deleted, ids...)
	m.tokens -= 4000 // simulate freeing tokens
	return nil
}
func (m *mockDB) StoreConcept(c *Concept) error {
	m.concepts = append(m.concepts, c)
	m.tokens += c.Tokens
	return nil
}
func (m *mockDB) GetContextWindow() (int, error) {
	return m.tokens, nil
}

type mockLLM struct{}

func (m *mockLLM) Generate(ctx context.Context, sys, prompt string) (string, error) {
	return "User needs n8n deployments for ERP. Prefers Go over Python.", nil
}

func TestLiquidContext_Consolidation(t *testing.T) {
	db := &mockDB{
		tokens: 9000, // Above the 8000 threshold
		episodes: []Episode{
			{ID: "ep1", Role: "user", Content: "Hey, can you help me build an n8n deployment?", Tokens: 2000},
			{ID: "ep2", Role: "agent", Content: "Sure, I can help. Do you prefer Go or Python?", Tokens: 2000},
			{ID: "ep3", Role: "user", Content: "I prefer Go.", Tokens: 1000},
		},
	}

	llm := &mockLLM{}
	lc := NewLiquidContext(db, llm, 8000, 10*time.Millisecond)

	// Manually trigger consolidation for the test
	lc.consolidate(context.Background())

	// 1. Ensure old episodes were deleted
	if len(db.deleted) != 3 {
		t.Errorf("expected 3 episodes to be deleted, got %d", len(db.deleted))
	}

	// 2. Ensure the new dense concept was stored
	if len(db.concepts) != 1 {
		t.Fatalf("expected 1 concept to be stored, got %d", len(db.concepts))
	}

	if db.concepts[0].DenseSummary != "User needs n8n deployments for ERP. Prefers Go over Python." {
		t.Errorf("unexpected concept summary: %s", db.concepts[0].DenseSummary)
	}

	// 3. Ensure tokens were reduced
	if db.tokens >= 8000 {
		t.Errorf("expected tokens to drop below threshold, currently %d", db.tokens)
	}
}
