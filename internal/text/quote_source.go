package text

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"typer/internal/storage"
)

// Stable registry IDs for settings.json and CLI (no dots in keys).
const (
	QuoteRemoteIDZenquotes = "zenquotes"
	QuoteRemoteIDTypefit   = "typefit"
)

// Quote provenance labels on Prompt.Source / CachedQuote.Source (which API or bundled list fed the quote).
// Not the same as session --source modes (remote|cache|seed) nor QuoteRemoteID* (settings toggles).
const (
	quoteSourceZenQuotes = "zenquotes"
	quoteSourceTypeFit   = "type.fit"
)

// FrameCaptionForQuoteSource returns the TUI top-right badge for remote quote provenance (Prompt.Source),
// or "" when no badge (seed, cache, unknown).
func FrameCaptionForQuoteSource(promptSource string) string {
	switch strings.ToLower(strings.TrimSpace(promptSource)) {
	case quoteSourceZenQuotes:
		return "ZenQuotes"
	case quoteSourceTypeFit:
		return "type.fit"
	default:
		return ""
	}
}

const quoteFrameSourceCaptionPrefix = "@"

// QuoteFrameSourceCaption is the full top-right label in the quote passage frame (@ + API name), or "" when no badge.
func QuoteFrameSourceCaption(promptSource string) string {
	base := FrameCaptionForQuoteSource(promptSource)
	if base == "" {
		return ""
	}
	return quoteFrameSourceCaptionPrefix + base
}

// quoteRemoteHandler fetches quotes from one third-party API with its own JSON shape.
type quoteRemoteHandler interface {
	registryID() string
	promptSource() string
	defaultURL() string
	fetch(ctx context.Context, client *http.Client, url string) ([]storage.CachedQuote, error)
}

// quoteRemoteChain is the global fetch order when multiple remotes are enabled.
var quoteRemoteChain = []quoteRemoteHandler{
	zenQuotesHandler{},
	typeFitHandler{},
}

// KnownQuoteRemoteIDs returns registry IDs in chain order for CLI validation and display.
func KnownQuoteRemoteIDs() []string {
	out := make([]string, 0, len(quoteRemoteChain))
	for _, h := range quoteRemoteChain {
		out = append(out, h.registryID())
	}
	return out
}

// IsKnownQuoteRemoteID reports whether id is a registered remote quote source.
func IsKnownQuoteRemoteID(id string) bool {
	for _, h := range quoteRemoteChain {
		if h.registryID() == id {
			return true
		}
	}
	return false
}

func handlerByRegistryID(id string) quoteRemoteHandler {
	for _, h := range quoteRemoteChain {
		if h.registryID() == id {
			return h
		}
	}
	return nil
}

func readRemoteGET(ctx context.Context, client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("remote quote API returned status %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxRemoteBodyBytes))
}

// ResolveEnabledQuoteRemotes returns registry IDs to try, in chain order.
// enabledMap: nil or missing key means true for that ID.
// When sessionAllowlist is non-empty, only those IDs are included (session wins over settings).
func ResolveEnabledQuoteRemotes(enabledMap map[string]bool, sessionAllowlist []string) []string {
	var out []string
	sessionMode := len(sessionAllowlist) > 0
	allow := map[string]struct{}{}
	if sessionMode {
		for _, id := range sessionAllowlist {
			id = strings.ToLower(strings.TrimSpace(id))
			if id != "" {
				allow[id] = struct{}{}
			}
		}
	}
	for _, h := range quoteRemoteChain {
		id := h.registryID()
		if sessionMode {
			if _, ok := allow[id]; !ok {
				continue
			}
		} else {
			if enabledMap != nil {
				if v, ok := enabledMap[id]; ok && !v {
					continue
				}
			}
		}
		out = append(out, id)
	}
	return out
}
