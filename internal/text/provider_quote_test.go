package text

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"typer/internal/storage"
)

const testCacheFilename = "quotes_cache.json"
const testZenPoolFilename = "zenquotes_batch_pool.json"

func newTestQuoteProvider(t *testing.T, cache *storage.QuoteCacheStore, zenPool *storage.ZenQuotesBatchPool, cfg QuoteProviderConfig, client *http.Client) *QuoteProvider {
	t.Helper()
	if zenPool == nil {
		zenPool = storage.NewZenQuotesBatchPoolAt(filepath.Join(t.TempDir(), testZenPoolFilename))
	}
	return NewQuoteProviderForTesting(cache, zenPool, cfg, client)
}

func quoteCfgTypeFitOnly(url string) QuoteProviderConfig {
	return QuoteProviderConfig{
		EnabledRemoteIDs: []string{QuoteRemoteIDTypefit},
		URLs: map[string]string{
			QuoteRemoteIDTypefit: url,
		},
	}
}

func quoteCfgZenThenFit(zenURL, fitURL string) QuoteProviderConfig {
	return QuoteProviderConfig{
		URLs: map[string]string{
			QuoteRemoteIDZenquotes: zenURL,
			QuoteRemoteIDTypefit:   fitURL,
		},
	}
}

func TestQuoteProviderRemoteThenCache(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"text":"Quote one","author":"Author A"},{"text":"Quote one","author":"Author A"},{"text":"Quote two","author":"Author B"}]`))
	}))
	defer srv.Close()

	cache := storage.NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), testCacheFilename))
	p := newTestQuoteProvider(t, cache, nil, quoteCfgTypeFitOnly(srv.URL), srv.Client())

	prompt, err := p.Next(context.Background(), Constraints{Source: "remote"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prompt.Content == "" {
		t.Fatalf("expected quote content")
	}
	if prompt.Source != quoteSourceTypeFit {
		t.Fatalf("source = %q, want type.fit", prompt.Source)
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

	p := newTestQuoteProvider(t, cache, nil, quoteCfgTypeFitOnly(srv.URL), srv.Client())
	prompt, err := p.Next(context.Background(), Constraints{Source: "remote"})
	if err != nil {
		t.Fatalf("expected fallback to cache, got error: %v", err)
	}
	if prompt.Source != "cache" {
		t.Fatalf("expected cache source, got %s", prompt.Source)
	}
}

func TestSanitizeQuoteField(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"strips escape sequences", "hello\x1b[31mRED\x1b[0mworld", "hello[31mRED[0mworld"},
		{"strips OSC bell", "safe\x07text", "safetext"},
		{"strips NUL and DEL", "a\x00b\x7fc", "abc"},
		{"strips C1 controls", "a\u009bfoo", "afoo"},
		{"keeps tab", "a\tb", "a\tb"},
		{"trims whitespace", "  padded \n", "padded"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sanitizeQuoteField(tc.in, maxQuoteRuneLen); got != tc.want {
				t.Fatalf("sanitize(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}

}

func TestSanitizeQuoteField_Truncates(t *testing.T) {
	in := strings.Repeat("a", maxQuoteRuneLen+50)
	got := sanitizeQuoteField(in, maxQuoteRuneLen)
	if len([]rune(got)) != maxQuoteRuneLen {
		t.Fatalf("expected truncation to %d runes, got %d", maxQuoteRuneLen, len([]rune(got)))
	}
}

func TestQuoteProviderRemote_SanitizesContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"text":"evil\u001b[2Jquote","author":"bad\u0007guy"}]`))
	}))
	defer srv.Close()

	cache := storage.NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), testCacheFilename))
	p := newTestQuoteProvider(t, cache, nil, quoteCfgTypeFitOnly(srv.URL), srv.Client())

	prompt, err := p.Next(context.Background(), Constraints{Source: "remote"})
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if strings.ContainsRune(prompt.Content, 0x1b) {
		t.Fatalf("ESC leaked into content: %q", prompt.Content)
	}
	if strings.ContainsRune(prompt.Author, 0x07) {
		t.Fatalf("BEL leaked into author: %q", prompt.Author)
	}

	cached, err := cache.Load()
	if err != nil {
		t.Fatalf("Load cache: %v", err)
	}
	if len(cached.Quotes) == 0 {
		t.Fatalf("expected cached quote")
	}
	if strings.ContainsRune(cached.Quotes[0].Content, 0x1b) {
		t.Fatalf("ESC persisted to cache: %q", cached.Quotes[0].Content)
	}
}

func TestQuoteProviderRemote_LimitsBodySize(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"text":"`))
		huge := make([]byte, maxRemoteBodyBytes+1024)
		for i := range huge {
			huge[i] = 'a'
		}
		w.Write(huge)
		w.Write([]byte(`","author":"a"}]`))
	}))
	defer srv.Close()

	cache := storage.NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), testCacheFilename))
	p := newTestQuoteProvider(t, cache, nil, quoteCfgTypeFitOnly(srv.URL), srv.Client())
	prompt, err := p.Next(context.Background(), Constraints{Source: "remote"})
	if err != nil {
		t.Fatalf("expected fallback to seed, got error: %v", err)
	}
	if prompt.Source != "seed" {
		t.Fatalf("expected seed source after oversized body, got %q", prompt.Source)
	}
}

func TestQuoteProviderRemoteFallbackToSeed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cache := storage.NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), testCacheFilename))
	p := newTestQuoteProvider(t, cache, nil, quoteCfgTypeFitOnly(srv.URL), srv.Client())

	prompt, err := p.Next(context.Background(), Constraints{Source: "remote"})
	if err != nil {
		t.Fatalf("expected seed fallback, got error: %v", err)
	}
	if prompt.Source != "seed" {
		t.Fatalf("expected seed source, got %s", prompt.Source)
	}
}

func TestQuoteProviderEmptySourceDefaultsToSeed(t *testing.T) {
	cache := storage.NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), testCacheFilename))
	p := newTestQuoteProvider(t, cache, nil, QuoteProviderConfig{}, nil)

	prompt, err := p.Next(context.Background(), Constraints{})
	if err != nil {
		t.Fatalf("expected seed default, got error: %v", err)
	}
	if prompt.Source != quoteSourceSeed {
		t.Fatalf("expected seed source by default, got %q", prompt.Source)
	}
}

func TestQuoteProviderZenQuotesBatchPool(t *testing.T) {
	dir := t.TempDir()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"q":"Hello","a":"World"},{"q":"Second","a":"Author"}]`))
	}))
	defer srv.Close()

	cache := storage.NewQuoteCacheStoreAt(filepath.Join(dir, testCacheFilename))
	pool := storage.NewZenQuotesBatchPoolAt(filepath.Join(dir, testZenPoolFilename))
	cfg := QuoteProviderConfig{
		EnabledRemoteIDs: []string{QuoteRemoteIDZenquotes},
		URLs: map[string]string{
			QuoteRemoteIDZenquotes: srv.URL,
		},
	}
	p := newTestQuoteProvider(t, cache, pool, cfg, srv.Client())

	prompt, err := p.Next(context.Background(), Constraints{Source: "remote"})
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if prompt.Content != "Hello" && prompt.Content != "Second" {
		t.Fatalf("prompt = %#v", prompt)
	}
	if prompt.Source != quoteSourceZenQuotes {
		t.Fatalf("source = %q, want zenquotes", prompt.Source)
	}

	n, err := pool.Len()
	if err != nil {
		t.Fatalf("Len: %v", err)
	}
	if n != 1 {
		t.Fatalf("pool len = %d, want 1 after first pop", n)
	}

	stored, err := cache.Load()
	if err != nil {
		t.Fatalf("cache Load: %v", err)
	}
	if len(stored.Quotes) != 0 {
		t.Fatalf("zenquotes batch should not write quotes_cache.json, got %d", len(stored.Quotes))
	}
}

func TestQuoteProviderZenQuotesBatchConsumesPool(t *testing.T) {
	dir := t.TempDir()
	if err := storage.NewZenQuotesBatchPoolAt(filepath.Join(dir, testZenPoolFilename)).Refill([]storage.CachedQuote{
		{Content: "A", Author: "1", Source: quoteSourceZenQuotes},
		{Content: "B", Author: "2", Source: quoteSourceZenQuotes},
	}); err != nil {
		t.Fatalf("Refill: %v", err)
	}

	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		_, _ = w.Write([]byte(`[{"q":"Nope","a":"X"}]`))
	}))
	defer srv.Close()

	cache := storage.NewQuoteCacheStoreAt(filepath.Join(dir, testCacheFilename))
	pool := storage.NewZenQuotesBatchPoolAt(filepath.Join(dir, testZenPoolFilename))
	cfg := QuoteProviderConfig{
		EnabledRemoteIDs: []string{QuoteRemoteIDZenquotes},
		URLs:           map[string]string{QuoteRemoteIDZenquotes: srv.URL},
	}
	p := newTestQuoteProvider(t, cache, pool, cfg, srv.Client())

	seen := map[string]bool{}
	for range 2 {
		prompt, err := p.Next(context.Background(), Constraints{Source: "remote"})
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		seen[prompt.Content] = true
	}
	if !seen["A"] || !seen["B"] {
		t.Fatalf("expected both pool quotes, got %v", seen)
	}
	if hit {
		t.Fatal("HTTP should not run while pool has quotes")
	}
}

func TestQuoteProviderZenQuotesRandomJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"q":"Random one","a":"Author R"}]`))
	}))
	defer srv.Close()

	cache := storage.NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), testCacheFilename))
	cfg := QuoteProviderConfig{
		EnabledRemoteIDs: []string{QuoteRemoteIDZenquotesRandom},
		URLs: map[string]string{
			QuoteRemoteIDZenquotesRandom: srv.URL,
		},
	}
	p := newTestQuoteProvider(t, cache, nil, cfg, srv.Client())
	prompt, err := p.Next(context.Background(), Constraints{Source: "remote"})
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if prompt.Content != "Random one" || prompt.Author != "Author R" {
		t.Fatalf("prompt = %#v", prompt)
	}
	if prompt.Source != quoteSourceZenQuotes {
		t.Fatalf("source = %q, want zenquotes", prompt.Source)
	}
}

func TestQuoteProviderSkipsZenQuotesWhenDisabled(t *testing.T) {
	fitSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"text":"From fit","author":"A"}]`))
	}))
	defer fitSrv.Close()

	badZen := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer badZen.Close()

	cache := storage.NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), testCacheFilename))
	cfg := QuoteProviderConfig{
		EnabledRemoteIDs: []string{QuoteRemoteIDTypefit},
		URLs: map[string]string{
			QuoteRemoteIDZenquotes: badZen.URL,
			QuoteRemoteIDTypefit:   fitSrv.URL,
		},
	}
	p := newTestQuoteProvider(t, cache, nil, cfg, fitSrv.Client())
	prompt, err := p.Next(context.Background(), Constraints{Source: "remote"})
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if prompt.Source != quoteSourceTypeFit {
		t.Fatalf("want type.fit, got %q", prompt.Source)
	}
	if prompt.Content != "From fit" {
		t.Fatalf("content = %q", prompt.Content)
	}
}

func TestQuoteProviderLegacySourceAutoAlias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"text":"Legacy","author":"X"}]`))
	}))
	defer srv.Close()

	cache := storage.NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), testCacheFilename))
	p := newTestQuoteProvider(t, cache, nil, quoteCfgTypeFitOnly(srv.URL), srv.Client())
	prompt, err := p.Next(context.Background(), Constraints{Source: "auto"})
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if prompt.Content != "Legacy" {
		t.Fatalf("legacy auto alias should behave like remote, got %#v", prompt)
	}
}

func TestQuoteProviderFallbackZenToFit(t *testing.T) {
	badZen := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer badZen.Close()

	fitSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"text":"Second","author":"B"}]`))
	}))
	defer fitSrv.Close()

	cache := storage.NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), testCacheFilename))
	cfg := quoteCfgZenThenFit(badZen.URL, fitSrv.URL)
	p := newTestQuoteProvider(t, cache, nil, cfg, fitSrv.Client())
	prompt, err := p.Next(context.Background(), Constraints{Source: "remote"})
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if prompt.Source != quoteSourceTypeFit {
		t.Fatalf("want type.fit after zen fails, got %q", prompt.Source)
	}
	if prompt.Content != "Second" {
		t.Fatalf("content = %q", prompt.Content)
	}
}

func TestQuoteProviderEmptyEnabledRemoteIDsSkipsHTTP(t *testing.T) {
	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		_, _ = w.Write([]byte(`[{"q":"nope","a":"x"}]`))
	}))
	defer srv.Close()

	cache := storage.NewQuoteCacheStoreAt(filepath.Join(t.TempDir(), testCacheFilename))
	cfg := QuoteProviderConfig{
		EnabledRemoteIDs: []string{},
		URLs: map[string]string{
			QuoteRemoteIDZenquotes: srv.URL,
		},
	}
	p := newTestQuoteProvider(t, cache, nil, cfg, srv.Client())
	_, err := p.Next(context.Background(), Constraints{Source: "remote"})
	if err != nil {
		t.Fatalf("Next: %v", err)
	}
	if hit {
		t.Fatal("remote HTTP should not run when EnabledRemoteIDs is empty")
	}
}
