package text

import (
	"encoding/json"
	"errors"

	"typer/internal/storage"
)

func parseZenQuotesPayload(body []byte, source string) ([]storage.CachedQuote, error) {
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
			Author:  item.A,
			Source:  source,
		})
	}
	if len(out) == 0 {
		return nil, errors.New("remote quote API returned empty payload")
	}
	return out, nil
}
