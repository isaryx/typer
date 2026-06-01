package text

import (
	"context"
	"net/http"

	"typer/internal/storage"
)

const defaultZenQuotesRandomURL = "https://zenquotes.io/api/random"

type zenQuotesRandomHandler struct{}

func (zenQuotesRandomHandler) registryID() string { return QuoteRemoteIDZenquotesRandom }

func (zenQuotesRandomHandler) promptSource() string { return quoteSourceZenQuotes }

func (zenQuotesRandomHandler) defaultURL() string { return defaultZenQuotesRandomURL }

func (h zenQuotesRandomHandler) fetch(ctx context.Context, client *http.Client, url string) ([]storage.CachedQuote, error) {
	if url == "" {
		url = h.defaultURL()
	}
	body, err := readRemoteGET(ctx, client, url)
	if err != nil {
		return nil, err
	}
	return parseZenQuotesPayload(body, h.promptSource())
}
