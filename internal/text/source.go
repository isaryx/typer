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
// (internal names used by providers and persisted sessions).
func CanonicalMode(mode string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "passage", "passages", "p":
		return model.ModePassage, nil
	case "words", "w":
		return model.ModeWords, nil
	case "quote", "quotes", "q":
		return model.ModeQuote, nil
	default:
		return "", fmt.Errorf("unsupported mode %q (valid: passages|p, words|w, quotes|q)", mode)
	}
}

// NewProvider expects a canonical mode (as returned by CanonicalMode).
// quoteCfg is used only for quotes mode (remote API selection and URL overrides).
func NewProvider(mode string, cache *storage.QuoteCacheStore, wordsFile, passagesFile string, quoteCfg QuoteProviderConfig) (Provider, error) {
	switch mode {
	case model.ModePassage:
		return NewLocalProvider(passagesFile)
	case model.ModeWords:
		return NewWordsProvider(wordsFile)
	case model.ModeQuote:
		if cache == nil {
			return nil, errors.New("quotes mode requires cache store")
		}
		return NewQuoteProvider(cache, quoteCfg), nil
	default:
		return nil, fmt.Errorf("unsupported mode %q (valid: passage, words, quote)", mode)
	}
}
