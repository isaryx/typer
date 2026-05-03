package text

import (
	"testing"

	"typer/internal/model"
)

func TestQuoteModeMayBlockOnNetwork(t *testing.T) {
	cases := []struct {
		mode, src string
		want      bool
	}{
		{model.ModeQuote, "remote", true},
		{model.ModeQuote, "REMOTE", true},
		{model.ModeQuote, "", true},
		{model.ModeQuote, "auto", true},
		{model.ModeQuote, "cache", false},
		{model.ModeQuote, "seed", false},
		{model.ModeQuote, "zenquotes", false},
		{model.ModeWords, "remote", false},
	}
	for _, tc := range cases {
		if got := QuoteModeMayBlockOnNetwork(tc.mode, tc.src); got != tc.want {
			t.Fatalf("QuoteModeMayBlockOnNetwork(%q, %q) = %v, want %v", tc.mode, tc.src, got, tc.want)
		}
	}
}
