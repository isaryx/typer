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
	want := []string{QuoteRemoteIDZenquotes, QuoteRemoteIDTypefit}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestResolveEnabledQuoteRemotes_zenquotesRandomOptIn(t *testing.T) {
	got := ResolveEnabledQuoteRemotes(map[string]bool{QuoteRemoteIDZenquotesRandom: true}, nil)
	want := []string{QuoteRemoteIDZenquotes, QuoteRemoteIDZenquotesRandom, QuoteRemoteIDTypefit}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestQuoteRemoteEffectiveEnabled(t *testing.T) {
	if !QuoteRemoteEffectiveEnabled(nil, QuoteRemoteIDZenquotes) {
		t.Fatal("zenquotes should default on")
	}
	if QuoteRemoteEffectiveEnabled(nil, QuoteRemoteIDZenquotesRandom) {
		t.Fatal("zenquotes-random should default off")
	}
	if !QuoteRemoteEffectiveEnabled(map[string]bool{QuoteRemoteIDZenquotesRandom: true}, QuoteRemoteIDZenquotesRandom) {
		t.Fatal("explicit on should win")
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
