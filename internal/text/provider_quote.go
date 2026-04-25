package text

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"typer/assets"
	"typer/internal/model"
	"typer/internal/storage"
)

const typeFitURL = "https://type.fit/api/quotes"
const defaultRemoteQuoteLimit = 250
const (
	quoteSourceAuto   = "auto"
	quoteSourceRemote = "remote"
	quoteSourceCache  = "cache"
	quoteSourceSeed   = "seed"
	quoteSourceTypeFit = "type.fit"
)

// maxRemoteBodyBytes caps the response body from the third-party quote API so a
// hostile or compromised upstream cannot exhaust memory before we parse.
const maxRemoteBodyBytes = 2 << 20 // 2 MiB

// maxQuoteRuneLen bounds each cached quote/author field. The typing UI is not
// designed for arbitrary-length strings, and this also limits blast radius of a
// poisoned upstream.
const maxQuoteRuneLen = 1024

type QuoteProvider struct {
	cache    *storage.QuoteCacheStore
	http     *http.Client
	endpoint string
}

func NewQuoteProvider(cache *storage.QuoteCacheStore) *QuoteProvider {
	return &QuoteProvider{
		cache:    cache,
		endpoint: typeFitURL,
		http: &http.Client{
			Timeout: 4 * time.Second,
		},
	}
}

func NewQuoteProviderForTesting(cache *storage.QuoteCacheStore, endpoint string, client *http.Client) *QuoteProvider {
	p := NewQuoteProvider(cache)
	if strings.TrimSpace(endpoint) != "" {
		p.endpoint = endpoint
	}
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
	switch source {
	case quoteSourceSeed:
		return p.seedQuotes()
	case quoteSourceCache:
		return p.cacheThenSeed()
	case quoteSourceRemote, quoteSourceAuto:
		return p.remoteThenCacheThenSeed(ctx)
	default:
		return nil, fmt.Errorf(
			"unsupported quote source %q (valid: %s, %s, %s, %s)",
			source,
			quoteSourceAuto,
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

func (p *QuoteProvider) remoteThenCache(ctx context.Context) ([]storage.CachedQuote, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("remote quote API returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxRemoteBodyBytes))
	if err != nil {
		return nil, err
	}
	var payload []struct {
		Text   string `json:"text"`
		Author string `json:"author"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	quotes := make([]storage.CachedQuote, 0, len(payload))
	seen := map[string]struct{}{}
	for _, item := range payload {
		content := sanitizeQuoteField(item.Text, maxQuoteRuneLen)
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
			Source:  quoteSourceTypeFit,
		})
		if len(quotes) >= defaultRemoteQuoteLimit {
			break
		}
	}
	if len(quotes) == 0 {
		return nil, errors.New("remote quote API returned empty payload")
	}
	if err := p.cache.Save(quotes); err != nil {
		log.Printf("typer: failed to save quote cache: %v", err)
	}
	return quotes, nil
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
