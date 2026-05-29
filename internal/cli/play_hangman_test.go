package cli

import (
	"context"
	"strings"
	"testing"
)

func TestRunPlayHangmanRejectsInvalidStrikes(t *testing.T) {
	var stdout strings.Builder
	err := runPlayHangman(context.Background(), []string{"--strikes", "5"}, strings.NewReader(""), &stdout)
	if err == nil {
		t.Fatal("expected error for --strikes 5")
	}
	if !strings.Contains(err.Error(), "strikes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPlayHangmanHelp(t *testing.T) {
	var stdout strings.Builder
	if err := runPlayHangman(context.Background(), []string{"--help"}, strings.NewReader(""), &stdout); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(stdout.String(), "play hangman") {
		t.Fatalf("help output missing usage: %q", stdout.String())
	}
}

func TestRunPlayUnknownSubcommand(t *testing.T) {
	var stdout strings.Builder
	err := runPlay(context.Background(), []string{"nope"}, strings.NewReader(""), &stdout)
	if err == nil {
		t.Fatal("expected error")
	}
}
