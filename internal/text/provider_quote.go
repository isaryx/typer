package text

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"typer/assets"
	"typer/internal/model"
	"typer/internal/storage"
)

const typeFitURL = "https://type.fit/api/quotes"
const defaultRemoteQuoteLimit = 250

type QuoteProvider struct {
	cache    *storage.QuoteCacheStore
	rng      *rand.Rand
	http     *http.Client
	endpoint string
}

func NewQuoteProvider(cache *storage.QuoteCacheStore) *QuoteProvider {
	return &QuoteProvider{
		cache:    cache,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
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
		source = "remote"
	}

	quotes, err := p.loadQuotes(ctx, source)
	if err != nil {
		return model.Prompt{}, err
	}
	if len(quotes) == 0 {
		return model.Prompt{}, errors.New("no quotes available")
	}
	q := quotes[p.rng.Intn(len(quotes))]
	return model.Prompt{
		ID:      time.Now().UTC().Format(time.RFC3339Nano),
		Content: q.Content,
		Author:  q.Author,
		Source:  q.Source,
		Mode:    "quote",
	}, nil
}

func (p *QuoteProvider) loadQuotes(ctx context.Context, source string) ([]storage.CachedQuote, error) {
	switch source {
	case "seed":
		return p.seedQuotes()
	case "cache":
		return p.cacheThenSeed()
	case "remote", "auto":
		return p.remoteThenCacheThenSeed(ctx)
	default:
		return nil, fmt.Errorf("unsupported quote source %q (valid: auto, remote, cache, seed)", source)
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
	body, err := io.ReadAll(resp.Body)
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
		content := strings.TrimSpace(item.Text)
		if content == "" {
			continue
		}
		if _, ok := seen[content]; ok {
			continue
		}
		seen[content] = struct{}{}
		quotes = append(quotes, storage.CachedQuote{
			Content: content,
			Author:  strings.TrimSpace(item.Author),
			Source:  "type.fit",
		})
		if len(quotes) >= defaultRemoteQuoteLimit {
			break
		}
	}
	if len(quotes) == 0 {
		return nil, errors.New("remote quote API returned empty payload")
	}
	_ = p.cache.Save(quotes)
	return quotes, nil
}

func (p *QuoteProvider) seedQuotes() ([]storage.CachedQuote, error) {
	var quotes []storage.CachedQuote
	if err := json.Unmarshal(assets.QuotesSeed, &quotes); err != nil {
		return nil, err
	}
	return quotes, nil
}
