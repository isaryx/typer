package storage

import (
	"path/filepath"
	"testing"
)

func TestZenQuotesBatchPoolPopRemovesQuote(t *testing.T) {
	pool := NewZenQuotesBatchPoolAt(filepath.Join(t.TempDir(), zenQuotesBatchPoolFilename))
	quotes := []CachedQuote{
		{Content: "A", Author: "One", Source: "zenquotes"},
		{Content: "B", Author: "Two", Source: "zenquotes"},
		{Content: "C", Author: "Three", Source: "zenquotes"},
	}
	if err := pool.Refill(quotes); err != nil {
		t.Fatalf("Refill: %v", err)
	}

	q, ok, err := pool.PopRandom()
	if err != nil || !ok {
		t.Fatalf("PopRandom: ok=%v err=%v", ok, err)
	}
	if q.Content == "" {
		t.Fatal("expected content")
	}

	n, err := pool.Len()
	if err != nil {
		t.Fatalf("Len: %v", err)
	}
	if n != 2 {
		t.Fatalf("len = %d, want 2", n)
	}
}

func TestZenQuotesBatchPoolEmptyPop(t *testing.T) {
	pool := NewZenQuotesBatchPoolAt(filepath.Join(t.TempDir(), zenQuotesBatchPoolFilename))
	_, ok, err := pool.PopRandom()
	if err != nil {
		t.Fatalf("PopRandom: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false on empty pool")
	}
}

func TestZenQuotesBatchPoolRefillAndPop(t *testing.T) {
	pool := NewZenQuotesBatchPoolAt(filepath.Join(t.TempDir(), zenQuotesBatchPoolFilename))
	if err := pool.Refill([]CachedQuote{{Content: "Only", Author: "X", Source: "zenquotes"}}); err != nil {
		t.Fatalf("Refill: %v", err)
	}
	q, ok, err := pool.PopRandom()
	if err != nil || !ok || q.Content != "Only" {
		t.Fatalf("PopRandom = %#v ok=%v err=%v", q, ok, err)
	}
	_, ok, err = pool.PopRandom()
	if err != nil || ok {
		t.Fatalf("second pop: ok=%v err=%v", ok, err)
	}
}

func TestZenQuotesBatchPoolLoadMissingFileReturnsEmpty(t *testing.T) {
	pool := NewZenQuotesBatchPoolAt(filepath.Join(t.TempDir(), zenQuotesBatchPoolFilename))
	file, err := pool.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(file.Quotes) != 0 {
		t.Fatalf("expected empty quotes, got %d", len(file.Quotes))
	}
}
