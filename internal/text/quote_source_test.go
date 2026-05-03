package text

import (
	"reflect"
	"testing"
)

func TestResolveEnabledQuoteRemotes_sessionAllowlist(t *testing.T) {
	got := ResolveEnabledQuoteRemotes(map[string]bool{QuoteRemoteIDZenquotes: false}, []string{QuoteRemoteIDZenquotes})
	want := []string{QuoteRemoteIDZenquotes}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestResolveEnabledQuoteRemotes_settingsOff(t *testing.T) {
	got := ResolveEnabledQuoteRemotes(map[string]bool{QuoteRemoteIDZenquotes: false, QuoteRemoteIDTypefit: true}, nil)
	want := []string{QuoteRemoteIDTypefit}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestResolveEnabledQuoteRemotes_defaultsAllOn(t *testing.T) {
	got := ResolveEnabledQuoteRemotes(nil, nil)
	want := KnownQuoteRemoteIDs()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestFrameCaptionForQuoteSource(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{quoteSourceZenQuotes, "ZenQuotes"},
		{quoteSourceTypeFit, "type.fit"},
		{"ZENQUOTES", "ZenQuotes"},
		{" seed ", ""},
		{"cache", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := FrameCaptionForQuoteSource(c.in); got != c.want {
			t.Fatalf("FrameCaptionForQuoteSource(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestQuoteFrameSourceCaption(t *testing.T) {
	if got := QuoteFrameSourceCaption(quoteSourceZenQuotes); got != "@ZenQuotes" {
		t.Fatalf("got %q", got)
	}
	if got := QuoteFrameSourceCaption(quoteSourceTypeFit); got != "@type.fit" {
		t.Fatalf("got %q", got)
	}
	if QuoteFrameSourceCaption("seed") != "" {
		t.Fatal("want empty for seed")
	}
}
