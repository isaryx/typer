package cli

import (
	"context"
	"strings"
	"testing"
)

func TestParseRootChoice(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"1", "start", true},
		{"2", "train", true},
		{"3", "play", true},
		{"4", "history", true},
		{"5", "stats", true},
		{"start", "start", true},
		{"Train", "train", true},
		{"", "", false},
		{"6", "", false},
		{"nope", "", false},
	}
	for _, tc := range cases {
		got, ok := parseRootChoice(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("parseRootChoice(%q) = %q, %v; want %q, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestSelectRootCommand(t *testing.T) {
	t.Run("start by number", func(t *testing.T) {
		cmd, err := selectRootCommand(strings.NewReader("1\n"), &strings.Builder{})
		if err != nil || cmd != "start" {
			t.Fatalf("cmd=%q err=%v", cmd, err)
		}
	})

	t.Run("play by name", func(t *testing.T) {
		cmd, err := selectRootCommand(strings.NewReader("play\n"), &strings.Builder{})
		if err != nil || cmd != "play" {
			t.Fatalf("cmd=%q err=%v", cmd, err)
		}
	})

	t.Run("empty cancels", func(t *testing.T) {
		cmd, err := selectRootCommand(strings.NewReader("\n"), &strings.Builder{})
		if err != nil || cmd != "" {
			t.Fatalf("cmd=%q err=%v", cmd, err)
		}
	})

	t.Run("invalid choice", func(t *testing.T) {
		_, err := selectRootCommand(strings.NewReader("xyz\n"), &strings.Builder{})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestRunRootInteractiveShowsMenu(t *testing.T) {
	var stdout strings.Builder
	err := runRootInteractive(context.Background(), strings.NewReader("\n"), &stdout)
	if err != nil {
		t.Fatalf("runRootInteractive: %v", err)
	}
	got := stdout.String()
	for _, want := range []string{"Choose a command", "start", "train", "play"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected menu to include %q, got %q", want, got)
		}
	}
}
