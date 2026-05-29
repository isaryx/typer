package cli

import (
	"context"
	"strings"
	"testing"
)

func TestRunPlayDefenseDefaults(t *testing.T) {
	fs := strings.Builder{}
	// Invalid lives should fail fast without TUI.
	err := runPlayDefense(context.Background(), []string{"--lives", "0"}, strings.NewReader(""), &fs)
	if err == nil {
		t.Fatal("expected error for --lives 0")
	}
	if !strings.Contains(err.Error(), "lives") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunPlayDefenseHelp(t *testing.T) {
	var stdout strings.Builder
	if err := runPlayDefense(context.Background(), []string{"--help"}, strings.NewReader(""), &stdout); err != nil {
		t.Fatalf("help: %v", err)
	}
	if !strings.Contains(stdout.String(), "play defense") {
		t.Fatalf("help output missing usage: %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "--lives") {
		t.Fatalf("help should mention --lives")
	}
	if !strings.Contains(stdout.String(), "--seed") {
		t.Fatalf("help should mention --seed")
	}
}

func TestRunPlayDefenseSubcommand(t *testing.T) {
	var stdout strings.Builder
	err := runPlay(context.Background(), []string{"defense", "--help"}, strings.NewReader(""), &stdout)
	if err != nil {
		t.Fatalf("defense help via play: %v", err)
	}
	if !strings.Contains(stdout.String(), "play defense") {
		t.Fatalf("expected defense help")
	}
}
