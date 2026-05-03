// Fetch policy helpers for quote mode (e.g. when loading may hit the network).
package text

import (
	"strings"

	"typer/internal/model"
)

// QuoteModeMayBlockOnNetwork reports whether loading the next quote may hit remote APIs
// (--source remote or empty default in quote mode). Local-only modes (cache, seed) return false.
func QuoteModeMayBlockOnNetwork(mode, source string) bool {
	if mode != model.ModeQuote {
		return false
	}
	s := strings.ToLower(strings.TrimSpace(source))
	if s == "" || s == "auto" {
		s = quoteSourceRemote
	}
	return s == quoteSourceRemote
}
