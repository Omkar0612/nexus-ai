package semantic

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockEmbedServer returns deterministic embeddings based on text.
func mockEmbedServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]string
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad", 400)
			return
		}
		text := req["prompt"]
		vec := make([]float64, 4)
		for i, ch := range text {
			if i >= 4 {
				break
			}
			vec[i] = float64(ch) / 1000.0
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"embedding": vec})
	}))
}

func TestCosineSimilarityIdentical(t *testing.T) {
	v := []float64{1, 2, 3, 4}
	got := cosineSimilarity(v, v)
	if math.Abs(got-1.0) > 1e-9 {
		t.Errorf("identical vectors should have similarity 1.0, got %f", got)
	}
}

func TestCosineSimilarityOrthogonal(t *testing.T) {
	a := []float64{1, 0, 0, 0}
	b := []float64{0, 1, 0, 0}
	got := cosineSimilarity(a, b)
	if math.Abs(got) > 1e-9 {
		t.Errorf("orthogonal vectors should have similarity 0, got %f", got)
	}
}

func TestCosineSimilarityMismatch(t *testing.T) {
	a := []float64{1, 2, 3}
	b := []float64{1, 2}
	if cosineSimilarity(a, b) != 0 {
		t.Error("mismatched lengths should return 0")
	}
}

func TestStoreAddAndSearch(t *testing.T) {
	ts := mockEmbedServer(t)
	defer ts.Close()

	store, err := New(":memory:", ts.URL, "nomic-embed-text")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	if _, err = store.Add(ctx, "hello world", "test"); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if _, err = store.Add(ctx, "goodbye world", "test"); err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	count, err := store.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected 2 documents, got %d", count)
	}

	results, err := store.Search(ctx, "hello", 1)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestStoreDelete(t *testing.T) {
	ts := mockEmbedServer(t)
	defer ts.Close()

	store, err := New(":memory:", ts.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()
	doc, err := store.Add(ctx, "delete me", "test")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(ctx, doc.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	count, _ := store.Count(ctx)
	if count != 0 {
		t.Errorf("expected 0 after delete, got %d", count)
	}
}
