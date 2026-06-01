package text

import (
	"context"
	"net/http"

	"typer/internal/storage"
)

const defaultZenQuotesBatchURL = "https://zenquotes.io/api/quotes"

const zenQuotesBatchPopRetries = 5

type zenQuotesBatchHandler struct{}

func (zenQuotesBatchHandler) registryID() string { return QuoteRemoteIDZenquotes }

func (zenQuotesBatchHandler) promptSource() string { return quoteSourceZenQuotes }

func (zenQuotesBatchHandler) defaultURL() string { return defaultZenQuotesBatchURL }

func (h zenQuotesBatchHandler) fetch(ctx context.Context, client *http.Client, url string) ([]storage.CachedQuote, error) {
	if url == "" {
		url = h.defaultURL()
	}
	body, err := readRemoteGET(ctx, client, url)
	if err != nil {
		return nil, err
	}
	return parseZenQuotesPayload(body, h.promptSource())
}

// takeFromZenQuotesBatch pops from the pool, refilling from the batch API when empty.
// ok is false when the pool path cannot supply a quote (caller should fall through to other remotes).
func takeFromZenQuotesBatch(
	ctx context.Context,
	pool *storage.ZenQuotesBatchPool,
	client *http.Client,
	url string,
) (storage.CachedQuote, bool, error) {
	if pool == nil {
		return storage.CachedQuote{}, false, nil
	}
	h := zenQuotesBatchHandler{}
	if url == "" {
		url = h.defaultURL()
	}

	tryPop := func() (storage.CachedQuote, bool, error) {
		for range zenQuotesBatchPopRetries {
			q, ok, err := pool.PopRandom()
			if err != nil {
				return storage.CachedQuote{}, false, err
			}
			if !ok {
				return storage.CachedQuote{}, false, nil
			}
			content := sanitizeQuoteField(q.Content, maxQuoteRuneLen)
			if content == "" {
				continue
			}
			return storage.CachedQuote{
				Content: content,
				Author:  sanitizeQuoteField(q.Author, maxQuoteRuneLen),
				Source:  q.Source,
			}, true, nil
		}
		return storage.CachedQuote{}, false, nil
	}

	if q, ok, err := tryPop(); err != nil || ok {
		return q, ok, err
	}

	raw, err := h.fetch(ctx, client, url)
	if err != nil {
		return storage.CachedQuote{}, false, nil
	}
	quotes := normalizeRemoteQuotes(raw)
	if len(quotes) == 0 {
		return storage.CachedQuote{}, false, nil
	}
	if err := pool.Refill(quotes); err != nil {
		return storage.CachedQuote{}, false, err
	}
	return tryPop()
}
