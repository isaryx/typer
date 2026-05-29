package cli

import (
	"context"
	"strings"
	"testing"
)

func TestParsePlayChoice(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"1", "hangman", true},
		{"2", "defense", true},
		{"hangman", "hangman", true},
		{"Defense", "defense", true},
		{"", "", false},
		{"3", "", false},
		{"nope", "", false},
	}
	for _, tc := range cases {
		got, ok := parsePlayChoice(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("parsePlayChoice(%q) = %q, %v; want %q, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestSelectPlayGame(t *testing.T) {
	t.Run("hangman by number", func(t *testing.T) {
		game, err := selectPlayGame(strings.NewReader("1\n"), &strings.Builder{})
		if err != nil || game != "hangman" {
			t.Fatalf("game=%q err=%v", game, err)
		}
	})

	t.Run("defense by name", func(t *testing.T) {
		game, err := selectPlayGame(strings.NewReader("defense\n"), &strings.Builder{})
		if err != nil || game != "defense" {
			t.Fatalf("game=%q err=%v", game, err)
		}
	})

	t.Run("empty cancels", func(t *testing.T) {
		game, err := selectPlayGame(strings.NewReader("\n"), &strings.Builder{})
		if err != nil || game != "" {
			t.Fatalf("game=%q err=%v", game, err)
		}
	})

	t.Run("invalid choice", func(t *testing.T) {
		_, err := selectPlayGame(strings.NewReader("xyz\n"), &strings.Builder{})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestRunPlayShowsMenu(t *testing.T) {
	var stdout strings.Builder
	err := runPlay(context.Background(), nil, strings.NewReader("\n"), &stdout)
	if err != nil {
		t.Fatalf("runPlay: %v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, "Choose a game") || !strings.Contains(got, "hangman") || !strings.Contains(got, "defense") {
		t.Fatalf("expected menu, got %q", got)
	}
}

func TestRunPlayHelp(t *testing.T) {
	var stdout strings.Builder
	if err := runPlay(context.Background(), []string{"--help"}, strings.NewReader(""), &stdout); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(stdout.String(), "typer play") || !strings.Contains(stdout.String(), "hangman") {
		t.Fatalf("help output missing usage: %q", stdout.String())
	}
}
