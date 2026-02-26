package memory

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// Memory is a single stored memory entry
type Memory struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`      // episodic, semantic, working
	Tags      string    `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	Importance float64  `json:"importance"`
}

// Store is the NEXUS memory database
type Store struct {
	db   *sql.DB
	path string
}

// New opens (or creates) the memory store at the given path
func New(dataDir string) (*Store, error) {
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".nexus")
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dataDir, "memory.db")
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open memory db: %w", err)
	}
	s := &Store{db: db, path: dbPath}
	return s, s.migrate()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS memories (
			id          TEXT PRIMARY KEY,
			user_id     TEXT NOT NULL,
			role        TEXT NOT NULL DEFAULT 'assistant',
			content     TEXT NOT NULL,
			type        TEXT NOT NULL DEFAULT 'episodic',
			tags        TEXT DEFAULT '',
			importance  REAL DEFAULT 0.5,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_memories_user_type ON memories(user_id, type);
		CREATE INDEX IF NOT EXISTS idx_memories_created ON memories(created_at);
	`)
	return err
}

// Add stores a new memory entry
func (s *Store) Add(userID, role, content, memType string, importance float64) error {
	id := fmt.Sprintf("mem-%d", time.Now().UnixNano())
	_, err := s.db.Exec(
		`INSERT INTO memories (id, user_id, role, content, type, importance) VALUES (?, ?, ?, ?, ?, ?)`,
		id, userID, role, content, memType, importance,
	)
	if err != nil {
		log.Error().Err(err).Str("user", userID).Msg("failed to store memory")
	}
	return err
}

// GetEpisodicHistory returns the N most recent episodic memories for a user
func (s *Store) GetEpisodicHistory(userID string, limit int) ([]Memory, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, role, content, type, tags, importance, created_at
		 FROM memories WHERE user_id = ? AND type = 'episodic'
		 ORDER BY created_at DESC LIMIT ?`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var mems []Memory
	for rows.Next() {
		var m Memory
		if err := rows.Scan(&m.ID, &m.UserID, &m.Role, &m.Content, &m.Type, &m.Tags, &m.Importance, &m.CreatedAt); err != nil {
			return nil, err
		}
		mems = append(mems, m)
	}
	return mems, rows.Err()
}

// Search returns memories matching a keyword
func (s *Store) Search(userID, query string, limit int) ([]Memory, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, role, content, type, tags, importance, created_at
		 FROM memories WHERE user_id = ? AND content LIKE ?
		 ORDER BY importance DESC, created_at DESC LIMIT ?`,
		userID, "%"+query+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var mems []Memory
	for rows.Next() {
		var m Memory
		if err := rows.Scan(&m.ID, &m.UserID, &m.Role, &m.Content, &m.Type, &m.Tags, &m.Importance, &m.CreatedAt); err != nil {
			return nil, err
		}
		mems = append(mems, m)
	}
	return mems, rows.Err()
}

// Delete removes memories older than the given duration for a user
func (s *Store) Delete(userID string, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	res, err := s.db.Exec(
		`DELETE FROM memories WHERE user_id = ? AND created_at < ?`,
		userID, cutoff,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Count returns the number of stored memories for a user
func (s *Store) Count(userID string) (int, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM memories WHERE user_id = ?`, userID).Scan(&n)
	return n, err
}

// Close shuts down the memory store
func (s *Store) Close() error { return s.db.Close() }
