package memory

import (
	"time"
)

// Episode represents a single turn in a conversation or a single agent action.
type Episode struct {
	ID        string
	Timestamp time.Time
	Role      string // e.g. "user", "agent_research", "system"
	Content   string
	Tokens    int
}

// Concept is a highly compressed, dense summary of multiple past episodes.
type Concept struct {
	ID           string
	Timestamp    time.Time
	Topic        string
	DenseSummary string
	Tokens       int
}

// Storage represents the SQLite Vector DB backend.
type Storage interface {
	GetRecentEpisodes(limit int) ([]Episode, error)
	GetOldEpisodes(olderThan time.Time) ([]Episode, error)
	DeleteEpisodes(ids []string) error
	StoreConcept(c *Concept) error
	GetContextWindow() (int, error) // Returns total tokens currently in context
}
