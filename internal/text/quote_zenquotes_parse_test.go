package text

import (
	"testing"
)

func TestParseZenQuotesPayload(t *testing.T) {
	raw := []byte(`[{"q":"Hello","a":"World"}]`)
	got, err := parseZenQuotesPayload(raw, quoteSourceZenQuotes)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(got) != 1 || got[0].Content != "Hello" || got[0].Author != "World" {
		t.Fatalf("got %#v", got)
	}
}

func TestParseZenQuotesPayload_empty(t *testing.T) {
	_, err := parseZenQuotesPayload([]byte(`[]`), quoteSourceZenQuotes)
	if err == nil {
		t.Fatal("expected error for empty payload")
	}
}
