package text

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"typer/internal/storage"
)

const defaultTypeFitURL = "https://type.fit/api/quotes"

type typeFitHandler struct{}

func (typeFitHandler) registryID() string { return QuoteRemoteIDTypefit }

func (typeFitHandler) promptSource() string { return quoteSourceTypeFit }

func (typeFitHandler) defaultURL() string { return defaultTypeFitURL }

func (h typeFitHandler) fetch(ctx context.Context, client *http.Client, url string) ([]storage.CachedQuote, error) {
	if url == "" {
		url = h.defaultURL()
	}
	body, err := readRemoteGET(ctx, client, url)
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
	out := make([]storage.CachedQuote, 0, len(payload))
	for _, item := range payload {
		out = append(out, storage.CachedQuote{
			Content: item.Text,
			Author:  item.Author,
			Source:  h.promptSource(),
		})
	}
	if len(out) == 0 {
		return nil, errors.New("remote quote API returned empty payload")
	}
	return out, nil
}
