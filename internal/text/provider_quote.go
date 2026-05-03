package text

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"typer/assets"
	"typer/internal/model"
	"typer/internal/storage"
)

const defaultRemoteQuoteLimit = 250

// Session source modes (Constraints.Source / --source): where the provider looks first.
// "remote" tries the enabled remote chain, then cache, then seed; "cache" and "seed" are stricter.
const (
	quoteSourceRemote = "remote"
	quoteSourceCache  = "cache"
	quoteSourceSeed   = "seed"
)

// maxRemoteBodyBytes caps the response body from the third-party quote API so a
// hostile or compromised upstream cannot exhaust memory before we parse.
const maxRemoteBodyBytes = 2 << 20 // 2 MiB

// maxQuoteRuneLen bounds each cached quote/author field. The typing UI is not
// designed for arbitrary-length strings, and this also limits blast radius of a
// poisoned upstream.
const maxQuoteRuneLen = 1024

// QuoteProviderConfig configures which remote APIs to call and optional URL overrides (tests).
type QuoteProviderConfig struct {
	// EnabledRemoteIDs is the ordered list of registry IDs to try. Nil means all known remotes
	// in chain order; an empty non-nil slice means no remotes (e.g. all toggled off in settings).
	EnabledRemoteIDs []string
	// URLs maps registry ID (QuoteRemoteID*) to a full URL override; empty string uses provider default.
	URLs map[string]string
}

type QuoteProvider struct {
	cache *storage.QuoteCacheStore
	http  *http.Client
	cfg   QuoteProviderConfig
}

// NewQuoteProvider builds the default quote provider: all remotes enabled, production URLs.
func NewQuoteProvider(cache *storage.QuoteCacheStore, cfg QuoteProviderConfig) *QuoteProvider {
	if cfg.URLs == nil {
		cfg.URLs = map[string]string{}
	}
	return &QuoteProvider{
		cache: cache,
		cfg:   cfg,
		http: &http.Client{
			Timeout: 4 * time.Second,
		},
	}
}

// NewQuoteProviderForTesting uses cfg plus an optional HTTP client (defaults to http.DefaultClient if nil).
func NewQuoteProviderForTesting(cache *storage.QuoteCacheStore, cfg QuoteProviderConfig, client *http.Client) *QuoteProvider {
	p := NewQuoteProvider(cache, cfg)
	if client != nil {
		p.http = client
	}
	return p
}

func (p *QuoteProvider) Name() string {
	return "quotes"
}

func (p *QuoteProvider) Next(ctx context.Context, c Constraints) (model.Prompt, error) {
	source := strings.ToLower(strings.TrimSpace(c.Source))
	if source == "" {
		source = quoteSourceSeed
	}

	quotes, err := p.loadQuotes(ctx, source)
	if err != nil {
		return model.Prompt{}, err
	}
	if len(quotes) == 0 {
		return model.Prompt{}, errors.New("no quotes available")
	}
	q := quotes[rand.IntN(len(quotes))]
	content := sanitizeQuoteField(q.Content, maxQuoteRuneLen)
	if content == "" {
		return model.Prompt{}, errors.New("no quotes available")
	}
	return model.Prompt{
		ID:      time.Now().UTC().Format(time.RFC3339Nano),
		Content: content,
		Author:  sanitizeQuoteField(q.Author, maxQuoteRuneLen),
		Source:  q.Source,
		Mode:    model.ModeQuote,
	}, nil
}

func (p *QuoteProvider) loadQuotes(ctx context.Context, source string) ([]storage.CachedQuote, error) {
	if source == "auto" {
		source = quoteSourceRemote // legacy alias from older sessions / CLI
	}
	switch source {
	case quoteSourceSeed:
		return p.seedQuotes()
	case quoteSourceCache:
		return p.cacheThenSeed()
	case quoteSourceRemote:
		return p.remoteThenCacheThenSeed(ctx)
	default:
		return nil, fmt.Errorf(
			"unsupported quote source %q (valid: %s, %s, %s)",
			source,
			quoteSourceRemote,
			quoteSourceCache,
			quoteSourceSeed,
		)
	}
}

func (p *QuoteProvider) remoteThenCacheThenSeed(ctx context.Context) ([]storage.CachedQuote, error) {
	quotes, err := p.remoteThenCache(ctx)
	if err == nil && len(quotes) > 0 {
		return quotes, nil
	}
	return p.cacheThenSeed()
}

func (p *QuoteProvider) cacheThenSeed() ([]storage.CachedQuote, error) {
	cache, err := p.cache.Load()
	if err == nil && len(cache.Quotes) > 0 {
		return cache.Quotes, nil
	}
	return p.seedQuotes()
}

func (p *QuoteProvider) enabledRemoteOrder() []string {
	if p.cfg.EnabledRemoteIDs != nil {
		return p.cfg.EnabledRemoteIDs
	}
	return KnownQuoteRemoteIDs()
}

func (p *QuoteProvider) remoteThenCache(ctx context.Context) ([]storage.CachedQuote, error) {
	ids := p.enabledRemoteOrder()
	for _, id := range ids {
		h := handlerByRegistryID(id)
		if h == nil {
			continue
		}
		url := ""
		if p.cfg.URLs != nil {
			url = p.cfg.URLs[id]
		}
		raw, err := h.fetch(ctx, p.http, url)
		if err != nil {
			continue
		}
		quotes := normalizeRemoteQuotes(raw)
		if len(quotes) == 0 {
			continue
		}
		_ = p.cache.Save(quotes)
		return quotes, nil
	}
	return nil, errors.New("remote quote APIs unavailable or returned empty payload")
}

func normalizeRemoteQuotes(raw []storage.CachedQuote) []storage.CachedQuote {
	quotes := make([]storage.CachedQuote, 0, len(raw))
	seen := map[string]struct{}{}
	for _, item := range raw {
		content := sanitizeQuoteField(item.Content, maxQuoteRuneLen)
		if content == "" {
			continue
		}
		if _, ok := seen[content]; ok {
			continue
		}
		seen[content] = struct{}{}
		quotes = append(quotes, storage.CachedQuote{
			Content: content,
			Author:  sanitizeQuoteField(item.Author, maxQuoteRuneLen),
			Source:  item.Source,
		})
		if len(quotes) >= defaultRemoteQuoteLimit {
			break
		}
	}
	return quotes
}

func (p *QuoteProvider) seedQuotes() ([]storage.CachedQuote, error) {
	var quotes []storage.CachedQuote
	if err := json.Unmarshal(assets.QuotesSeed, &quotes); err != nil {
		return nil, err
	}
	for i := range quotes {
		quotes[i].Source = quoteSourceSeed
	}
	return quotes, nil
}

// sanitizeQuoteField strips ASCII/C1 control characters (keeping \t) from
// untrusted third-party strings before they are cached or rendered to the
// terminal, then trims surrounding whitespace and truncates to maxRunes runes.
// This prevents a compromised or poisoned upstream from injecting ANSI/OSC
// escape sequences into the typing UI.
func sanitizeQuoteField(s string, maxRunes int) string {
	cleaned := strings.Map(func(r rune) rune {
		if r == '\t' {
			return r
		}
		if r < 0x20 || r == 0x7F {
			return -1
		}
		if r >= 0x80 && r < 0xA0 {
			return -1
		}
		return r
	}, s)
	cleaned = strings.TrimSpace(cleaned)
	if maxRunes > 0 {
		runes := []rune(cleaned)
		if len(runes) > maxRunes {
			cleaned = string(runes[:maxRunes])
		}
	}
	return cleaned
}
