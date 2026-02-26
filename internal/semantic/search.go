// Package semantic provides cosine-similarity semantic search over text corpora.
// Zero external deps — uses in-process TF-IDF vectors with RWMutex concurrency safety.
// For production: drop in a real embedding model (Ollama nomic-embed-text is free).
package semantic

import (
	"math"
	"sort"
	"strings"
	"sync"
)

// Document is a piece of text with an ID and optional metadata.
type Document struct {
	ID       string
	Text     string
	Metadata map[string]string
}

// Result is a ranked search hit.
type Result struct {
	Document Document
	Score    float64
}

// Index holds TF-IDF vectors for all indexed documents.
// All methods are safe for concurrent use.
type Index struct {
	mu   sync.RWMutex
	docs []Document
	vecs []map[string]float64
	idf  map[string]float64
}

// NewIndex creates an empty semantic index.
func NewIndex() *Index {
	return &Index{idf: make(map[string]float64)}
}

// Add indexes a document. Call Rebuild() after adding all documents.
func (idx *Index) Add(doc Document) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.docs = append(idx.docs, doc)
}

// Size returns the number of indexed documents.
func (idx *Index) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.docs)
}

// Rebuild recomputes TF-IDF vectors for all documents.
// Safe to call concurrently — acquires a write lock for the swap.
func (idx *Index) Rebuild() {
	idx.mu.RLock()
	docs := make([]Document, len(idx.docs))
	copy(docs, idx.docs)
	idx.mu.RUnlock()

	N := float64(len(docs))
	if N == 0 {
		return
	}
	// Pre-allocate with a reasonable capacity to reduce rehashing.
	df := make(map[string]float64, len(docs)*10)
	tfs := make([]map[string]float64, len(docs))

	for i, doc := range docs {
		tokens := tokenise(doc.Text)
		tf := make(map[string]float64, len(tokens))
		for _, t := range tokens {
			tf[t]++
		}
		total := float64(len(tokens))
		if total == 0 {
			tfs[i] = tf
			continue
		}
		for t := range tf {
			tf[t] /= total
			df[t]++
		}
		tfs[i] = tf
	}

	idf := make(map[string]float64, len(df))
	for t, d := range df {
		idf[t] = math.Log(N / d)
	}

	vecs := make([]map[string]float64, len(docs))
	for i, tf := range tfs {
		vec := make(map[string]float64, len(tf))
		for t, v := range tf {
			vec[t] = v * idf[t]
		}
		vecs[i] = vec
	}

	// Swap atomically under write lock.
	idx.mu.Lock()
	idx.idf = idf
	idx.vecs = vecs
	idx.mu.Unlock()
}

// Search returns the top-k most similar documents to the query.
func (idx *Index) Search(query string, topK int) []Result {
	idx.mu.RLock()
	docs := idx.docs
	vecs := idx.vecs
	idf := idx.idf
	idx.mu.RUnlock()

	qVec := buildQueryVec(query, idf)
	if len(qVec) == 0 {
		return nil // no known tokens in query — early exit
	}

	results := make([]Result, 0, len(docs))
	for i, vec := range vecs {
		if score := cosine(qVec, vec); score > 0 {
			results = append(results, Result{Document: docs[i], Score: score})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if topK > 0 && len(results) > topK {
		return results[:topK]
	}
	return results
}

// buildQueryVec computes a TF-IDF vector for the query using the index IDF table.
func buildQueryVec(query string, idf map[string]float64) map[string]float64 {
	tokens := tokenise(query)
	if len(tokens) == 0 {
		return nil
	}
	tf := make(map[string]float64, len(tokens))
	for _, t := range tokens {
		tf[t]++
	}
	total := float64(len(tokens))
	vec := make(map[string]float64, len(tf))
	for t, v := range tf {
		if w, ok := idf[t]; ok {
			vec[t] = (v / total) * w
		}
	}
	return vec
}

// cosine computes the cosine similarity between two sparse TF-IDF vectors.
// Iterates over the smaller map (a = query) for efficiency.
func cosine(a, b map[string]float64) float64 {
	var dot, normA float64
	for t, v := range a {
		dot += v * b[t]
		normA += v * v
	}
	if normA == 0 {
		return 0 // early exit — query vec is zero
	}
	var normB float64
	for _, v := range b {
		normB += v * v
	}
	if normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// tokenise lowercases, splits on non-alphanumeric chars, and removes stop words.
// Filters in-place (words[:0]) to avoid a second allocation.
func tokenise(text string) []string {
	text = strings.ToLower(text)
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !('a' <= r && r <= 'z') && !('0' <= r && r <= '9')
	})
	out := words[:0]
	for _, w := range words {
		if len(w) > 2 && !stopWords[w] {
			out = append(out, w)
		}
	}
	return out
}

// stopWords is a compile-time map — zero allocation at runtime.
var stopWords = map[string]bool{
	"the": true, "and": true, "for": true, "are": true, "but": true,
	"not": true, "you": true, "all": true, "can": true, "has": true,
	"was": true, "had": true, "its": true, "our": true, "this": true,
	"that": true, "with": true, "from": true, "they": true, "will": true,
	"have": true, "been": true, "more": true, "also": true, "into": true,
	"use": true, "used": true, "using": true, "via": true, "per": true,
}
