// Package semantic provides vector-based semantic memory search.
// Uses Ollama embeddings API (free, local) to embed text into float32 vectors,
// stored in SQLite. No external vector DB required — pure stdlib + sqlite.
// Replaces Mem.ai ($15/mo) and Notion AI search.
package semantic

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Document is a piece of text stored with its embedding.
type Document struct {
	ID        int64
	Content   string
	Source    string    // e.g. "conversation", "notes", "email", "kb"
	CreatedAt time.Time
	Score     float64   // populated on search results
}

// Store manages the vector store.
type Store struct {
	db         *sql.DB
	ollamaURL  string
	model      string
	httpClient *http.Client
}

// New opens (or creates) the semantic store at dbPath.
// ollamaURL defaults to http://localhost:11434.
// model defaults to "nomic-embed-text" (best free local embedding model).
func New(dbPath, ollamaURL, model string) (*Store, error) {
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	if model == "" {
		model = "nomic-embed-text"
	}
	db, err := sql.Open("sqlite3", dbPath+"?_journal=WAL")
	if err != nil {
		return nil, fmt.Errorf("semantic: open db: %w", err)
	}
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("semantic: migrate: %w", err)
	}
	return &Store{
		db:         db,
		ollamaURL:  strings.TrimRight(ollamaURL, "/"),
		model:      model,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS documents (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			content    TEXT    NOT NULL,
			source     TEXT    NOT NULL DEFAULT '',
			created_at INTEGER NOT NULL,
			embedding  TEXT    NOT NULL  -- JSON array of float64
		);
		CREATE INDEX IF NOT EXISTS idx_documents_source ON documents(source);
	`)
	return err
}

// Embed fetches an embedding vector from Ollama for the given text.
func (s *Store) Embed(ctx context.Context, text string) ([]float64, error) {
	reqBody := map[string]string{"model": s.model, "prompt": text}
	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.ollamaURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("semantic: embed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("semantic: embed: status %d", resp.StatusCode)
	}
	var result struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Embedding, nil
}

// Add embeds and stores a document.
func (s *Store) Add(ctx context.Context, content, source string) (*Document, error) {
	vec, err := s.Embed(ctx, content)
	if err != nil {
		return nil, err
	}
	vecJSON, err := json.Marshal(vec)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO documents (content, source, created_at, embedding) VALUES (?, ?, ?, ?)`,
		content, source, now.Unix(), string(vecJSON),
	)
	if err != nil {
		return nil, fmt.Errorf("semantic: insert: %w", err)
	}
	id, _ := res.LastInsertId()
	return &Document{ID: id, Content: content, Source: source, CreatedAt: now}, nil
}

// Search returns the topK most semantically similar documents to query.
func (s *Store) Search(ctx context.Context, query string, topK int) ([]Document, error) {
	queryVec, err := s.Embed(ctx, query)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, content, source, created_at, embedding FROM documents`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type scored struct {
		Doc   Document
		score float64
	}
	var results []scored
	for rows.Next() {
		var d Document
		var createdUnix int64
		var embJSON string
		if err := rows.Scan(&d.ID, &d.Content, &d.Source, &createdUnix, &embJSON); err != nil {
			return nil, err
		}
		d.CreatedAt = time.Unix(createdUnix, 0).UTC()
		var vec []float64
		if err := json.Unmarshal([]byte(embJSON), &vec); err != nil {
			continue
		}
		score := cosineSimilarity(queryVec, vec)
		results = append(results, scored{Doc: d, score: score})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})
	if topK > len(results) {
		topK = len(results)
	}
	out := make([]Document, topK)
	for i := 0; i < topK; i++ {
		out[i] = results[i].Doc
		out[i].Score = results[i].score
	}
	return out, nil
}

// Delete removes a document by ID.
func (s *Store) Delete(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM documents WHERE id = ?`, id)
	return err
}

// Count returns the total number of stored documents.
func (s *Store) Count(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents`).Scan(&n)
	return n, err
}

// Close closes the underlying database.
func (s *Store) Close() error { return s.db.Close() }

// cosineSimilarity computes cosine similarity between two equal-length vectors.
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
