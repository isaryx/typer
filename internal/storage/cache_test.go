package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

const quoteCacheJSONFile = "quotes_cache.json"

func typerCacheDirForTestHome(t *testing.T, home string) string {
	t.Helper()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Caches", "typer")
	case "windows":
		// os.UserCacheDir uses %LOCALAPPDATA% when set.
		localAppData := filepath.Join(home, "AppData", "Local")
		return filepath.Join(localAppData, "typer")
	default:
		// Typical Linux/XDG layout when XDG_* is not explicitly set.
		return filepath.Join(home, ".cache", "typer")
	}
}

func TestNewQuoteCacheStoreUsesCacheDir(t *testing.T) {
	home := t.TempDir()
	setTestUserDirs(t, home)

	wantCacheDir := typerCacheDirForTestHome(t, home)

	store, err := NewQuoteCacheStore()
	if err != nil {
		t.Fatalf("NewQuoteCacheStore: %v", err)
	}
	if got := filepath.Dir(store.path); got != wantCacheDir {
		t.Fatalf("cache dir = %q, want %q", got, wantCacheDir)
	}
	if filepath.Base(store.path) != quotesCacheFilename {
		t.Fatalf("unexpected cache filename: %q", store.path)
	}
}

func TestQuoteCacheStoreSaveLoadRoundTrip(t *testing.T) {
	store := NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), quoteCacheJSONFile))
	input := []CachedQuote{
		{Content: "Quote A", Author: "Author A", Source: "remote"},
		{Content: "Quote B", Author: "Author B", Source: "seed"},
	}
	if err := store.Save(input); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Quotes) != len(input) {
		t.Fatalf("len(Quotes) = %d, want %d", len(got.Quotes), len(input))
	}
	for i := range input {
		if got.Quotes[i] != input[i] {
			t.Fatalf("quotes[%d] = %#v, want %#v", i, got.Quotes[i], input[i])
		}
	}
	if got.FetchedAt.IsZero() {
		t.Fatal("expected non-zero FetchedAt")
	}
	if got.FetchedAt.After(time.Now().UTC().Add(2 * time.Second)) {
		t.Fatalf("FetchedAt seems invalid: %v", got.FetchedAt)
	}
}

func TestQuoteCacheStoreLoadMissingFileReturnsEmpty(t *testing.T) {
	store := NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), quoteCacheJSONFile))
	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got.Quotes) != 0 {
		t.Fatalf("expected empty quotes, got %d", len(got.Quotes))
	}
}

func TestQuoteCacheStoreLoadMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), quoteCacheJSONFile)
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatalf("write malformed cache: %v", err)
	}
	store := NewQuoteCacheStoreAt(path)
	if _, err := store.Load(); err == nil {
		t.Fatal("expected malformed JSON error")
	}
}
