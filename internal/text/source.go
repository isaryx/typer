package text

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"typer/internal/model"
	"typer/internal/storage"
)

type Constraints struct {
	Words  int
	Source string
}

type Provider interface {
	Next(ctx context.Context, c Constraints) (model.Prompt, error)
	Name() string
}

// CanonicalMode maps user input (including shorthands) to passage, words, or quote
// (internal names used by providers and persisted sessions). It also accepts those
// internal names so it is safe to call twice (CLI canonicalizes then NewProvider does).
func CanonicalMode(mode string) (string, error) {
	s := strings.ToLower(strings.TrimSpace(mode))
	switch s {
	case "passage", "passages", "p":
		return "passage", nil
	case "words", "w":
		return "words", nil
	case "quote", "quotes", "q":
		return "quote", nil
	default:
		return "", fmt.Errorf("unsupported mode %q (valid: passages|p, words|w, quotes|q)", mode)
	}
}

func NewProvider(mode string, cache *storage.QuoteCacheStore, wordsFile, passagesFile string) (Provider, error) {
	m, err := CanonicalMode(mode)
	if err != nil {
		return nil, err
	}
	switch m {
	case "passage":
		return NewLocalProvider(passagesFile)
	case "words":
		return NewWordsProvider(wordsFile)
	case "quote":
		if cache == nil {
			return nil, errors.New("quotes mode requires cache store")
		}
		return NewQuoteProvider(cache), nil
	default:
		return nil, fmt.Errorf("unsupported mode %q (valid: passages|p, words|w, quotes|q)", m)
	}
}
