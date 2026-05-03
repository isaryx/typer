package text

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"typer/internal/storage"
)

const defaultZenQuotesURL = "https://zenquotes.io/api/quotes"

type zenQuotesHandler struct{}

func (zenQuotesHandler) registryID() string { return QuoteRemoteIDZenquotes }

func (zenQuotesHandler) promptSource() string { return quoteSourceZenQuotes }

func (zenQuotesHandler) defaultURL() string { return defaultZenQuotesURL }

func (h zenQuotesHandler) fetch(ctx context.Context, client *http.Client, url string) ([]storage.CachedQuote, error) {
	if url == "" {
		url = h.defaultURL()
	}
	body, err := readRemoteGET(ctx, client, url)
	if err != nil {
		return nil, err
	}
	var payload []struct {
		Q string `json:"q"`
		A string `json:"a"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	out := make([]storage.CachedQuote, 0, len(payload))
	for _, item := range payload {
		out = append(out, storage.CachedQuote{
			Content: item.Q,
			Author:    item.A,
			Source:    h.promptSource(),
		})
	}
	if len(out) == 0 {
		return nil, errors.New("remote quote API returned empty payload")
	}
	return out, nil
}
