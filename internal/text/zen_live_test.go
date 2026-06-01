//go:build integration

package text

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"typer/internal/storage"
)

func TestLiveZenQuotesBatchRefill(t *testing.T) {
	dir := t.TempDir()
	pool := storage.NewZenQuotesBatchPoolAt(filepath.Join(dir, "pool.json"))
	client := &http.Client{Timeout: 60 * time.Second}
	q, ok, err := takeFromZenQuotesBatch(context.Background(), pool, client, "")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected ok from live zenquotes")
	}
	if q.Source != quoteSourceZenQuotes {
		t.Fatalf("source=%q", q.Source)
	}
	t.Logf("got: %q — %s", q.Content[:min(40, len(q.Content))], q.Author)
}
