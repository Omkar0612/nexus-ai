// Package semantic provides cosine-similarity semantic search over text corpora.
// Zero external deps — uses in-process TF-IDF vectors.
// For production: drop in a real embedding model (Ollama nomic-embed-text is free).
package semantic

import (
	"math"
	"sort"
	"strings"
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
type Index struct {
	docs   []Document
	vecs   []map[string]float64
	idf    map[string]float64
}

// NewIndex creates an empty semantic index.
func NewIndex() *Index {
	return &Index{idf: make(map[string]float64)}
}

// Add indexes a document. Call Rebuild() after adding all documents.
func (idx *Index) Add(doc Document) {
	idx.docs = append(idx.docs, doc)
}

// Rebuild recomputes TF-IDF vectors for all documents.
func (idx *Index) Rebuild() {
	N := float64(len(idx.docs))
	df := make(map[string]float64)
	tfs := make([]map[string]float64, len(idx.docs))
	for i, doc := range idx.docs {
		tokens := tokenise(doc.Text)
		tf := make(map[string]float64)
		for _, t := range tokens {
			tf[t]++
		}
		total := float64(len(tokens))
		for t := range tf {
			tf[t] /= total
			df[t]++
		}
		tfs[i] = tf
	}
	idf := make(map[string]float64)
	for t, d := range df {
		idf[t] = math.Log(N / d)
	}
	idx.idf = idf
	idx.vecs = make([]map[string]float64, len(idx.docs))
	for i, tf := range tfs {
		vec := make(map[string]float64, len(tf))
		for t, v := range tf {
			vec[t] = v * idf[t]
		}
		idx.vecs[i] = vec
	}
}

// Search returns the top-k most similar documents to the query.
func (idx *Index) Search(query string, topK int) []Result {
	qVec := idx.queryVec(query)
	results := make([]Result, 0, len(idx.docs))
	for i, vec := range idx.vecs {
		score := cosine(qVec, vec)
		if score > 0 {
			results = append(results, Result{Document: idx.docs[i], Score: score})
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}
	return results
}

func (idx *Index) queryVec(query string) map[string]float64 {
	tokens := tokenise(query)
	tf := make(map[string]float64)
	for _, t := range tokens {
		tf[t]++
	}
	total := float64(len(tokens))
	vec := make(map[string]float64, len(tf))
	for t, v := range tf {
		vec[t] = (v / total) * idx.idf[t]
	}
	return vec
}

func cosine(a, b map[string]float64) float64 {
	var dot, normA, normB float64
	for t, v := range a {
		dot += v * b[t]
		normA += v * v
	}
	for _, v := range b {
		normB += v * v
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func tokenise(text string) []string {
	text = strings.ToLower(text)
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !('a' <= r && r <= 'z') && !('0' <= r && r <= '9')
	})
	out := words[:0]
	for _, w := range words {
		if len(w) > 2 && !isStopWord(w) {
			out = append(out, w)
		}
	}
	return out
}

var stopWords = map[string]bool{
	"the": true, "and": true, "for": true, "are": true, "but": true,
	"not": true, "you": true, "all": true, "can": true, "has": true,
	"was": true, "had": true, "its": true, "our": true, "this": true,
	"that": true, "with": true, "from": true, "they": true, "will": true,
	"have": true, "been": true, "more": true, "also": true, "into": true,
}

func isStopWord(w string) bool { return stopWords[w] }
