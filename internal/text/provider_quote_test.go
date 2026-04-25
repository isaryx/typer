package text

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"typer/internal/storage"
)

const testCacheFilename = "quotes_cache.json"

func TestQuoteProviderRemoteThenCache(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"text":"Quote one","author":"Author A"},{"text":"Quote one","author":"Author A"},{"text":"Quote two","author":"Author B"}]`))
	}))
	defer srv.Close()

	cache := storage.NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), testCacheFilename))
	p := NewQuoteProviderForTesting(cache, srv.URL, srv.Client())

	prompt, err := p.Next(context.Background(), Constraints{Source: "remote"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prompt.Content == "" {
		t.Fatalf("expected quote content")
	}

	stored, err := cache.Load()
	if err != nil {
		t.Fatalf("expected cache read success: %v", err)
	}
	if len(stored.Quotes) != 2 {
		t.Fatalf("expected deduped quotes in cache, got %d", len(stored.Quotes))
	}
}

func TestQuoteProviderRemoteFallbackToCache(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cache := storage.NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), testCacheFilename))
	if err := cache.Save([]storage.CachedQuote{{
		Content: "Cached quote",
		Author:  "Cached Author",
		Source:  "cache",
	}}); err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	p := NewQuoteProviderForTesting(cache, srv.URL, srv.Client())
	prompt, err := p.Next(context.Background(), Constraints{Source: "remote"})
	if err != nil {
		t.Fatalf("expected fallback to cache, got error: %v", err)
	}
	if prompt.Source != "cache" {
		t.Fatalf("expected cache source, got %s", prompt.Source)
	}
}

func TestQuoteProviderRemoteFallbackToSeed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cache := storage.NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), testCacheFilename))
	p := NewQuoteProviderForTesting(cache, srv.URL, srv.Client())

	prompt, err := p.Next(context.Background(), Constraints{Source: "remote"})
	if err != nil {
		t.Fatalf("expected seed fallback, got error: %v", err)
	}
	if prompt.Source != "seed" {
		t.Fatalf("expected seed source, got %s", prompt.Source)
	}
}
