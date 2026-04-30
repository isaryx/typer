package model

import (
	"strings"
	"testing"
)

func TestCanonicalPromptContent_CRLFAndSpaces(t *testing.T) {
	a := CanonicalPromptContent("hello  world")
	b := CanonicalPromptContent("hello\t\tworld")
	if a != b {
		t.Fatalf("want equal canonical, got %q vs %q", a, b)
	}
	c := CanonicalPromptContent("hello\r\nworld")
	d := CanonicalPromptContent("hello\nworld")
	if c != d {
		t.Fatalf("CRLF vs LF: %q vs %q", c, d)
	}
}

func TestPromptContentHash_Stability(t *testing.T) {
	h := PromptContentHash("alpha beta")
	if len(h) != 64 {
		t.Fatalf("want 64 hex chars, got %d", len(h))
	}
	if PromptContentHash("alpha beta") != h {
		t.Fatal("hash not stable")
	}
	// different canonical -> different hash
	if PromptContentHash("alpha  beta") != h {
		// actually canonical collapses spaces so should equal
		if PromptContentHash("alpha  beta") != h {
			t.Fatal("canonical should collapse inner spaces")
		}
	}
}

func TestSessionContentHashKey_PrefersStored(t *testing.T) {
	r := SessionResult{
		ContentHash: "abc",
		Prompt:      Prompt{Content: "x"},
	}
	if SessionContentHashKey(r) != "abc" {
		t.Fatal("stored hash ignored")
	}
}

func TestSessionContentHashKey_LegacyFallback(t *testing.T) {
	want := PromptContentHash("hello")
	r := SessionResult{Prompt: Prompt{Content: "hello"}}
	if SessionContentHashKey(r) != want {
		t.Fatalf("fallback: got %q want %q", SessionContentHashKey(r), want)
	}
}

func TestCanonicalPromptContent_lineTrim(t *testing.T) {
	a := CanonicalPromptContent("  foo  \n  bar  ")
	b := CanonicalPromptContent("foo\nbar")
	if a != b {
		t.Fatalf("%q vs %q", a, b)
	}
}

func TestCanonicalPromptContent_multilineBlankLines(t *testing.T) {
	b := CanonicalPromptContent("a\n\n\nb")
	want := strings.Join([]string{"a", "", "", "b"}, "\n")
	if b != want {
		t.Fatalf("unexpected: %q want %q", b, want)
	}
}
