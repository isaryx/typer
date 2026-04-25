package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"typer/internal/model"
	"typer/internal/storage"
)

func TestExtractPresenceFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantRest   []string
		wantStrict bool
		wantIndef  bool
		wantErr    bool
	}{
		{"empty", nil, nil, false, false, false},
		{"no flags", []string{"--mode", "words"}, []string{"--mode", "words"}, false, false, false},
		{"bare strict", []string{"--strict"}, nil, true, false, false},
		{"strict then flags", []string{"--strict", "--mode", "words"}, []string{"--mode", "words"}, true, false, false},
		{"strict shorthand", []string{"-s", "-m", "words"}, []string{"-m", "words"}, true, false, false},
		{"bare indefinite", []string{"--indefinite"}, nil, false, true, false},
		{"indef shorthand", []string{"-i", "-m", "w"}, []string{"-m", "w"}, false, true, false},
		{"strict and indefinite", []string{"-s", "-i", "-w", "3"}, []string{"-w", "3"}, true, true, false},
		{"eq strict", []string{"--strict=false"}, nil, false, false, true},
		{"eq indefinite", []string{"-i=true"}, nil, false, false, true},
		{"space true after strict", []string{"--strict", "true"}, nil, false, false, true},
		{"word after strict ok", []string{"--strict", "hello"}, []string{"hello"}, true, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRest, gotStrict, gotIndef, err := extractPresenceFlags(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("extractPresenceFlags: err = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("extractPresenceFlags: %v", err)
			}
			if gotStrict != tt.wantStrict {
				t.Fatalf("strict = %v, want %v", gotStrict, tt.wantStrict)
			}
			if gotIndef != tt.wantIndef {
				t.Fatalf("indefinite = %v, want %v", gotIndef, tt.wantIndef)
			}
			if !reflect.DeepEqual(gotRest, tt.wantRest) {
				t.Fatalf("rest = %#v, want %#v", gotRest, tt.wantRest)
			}
		})
	}
}

func TestExecute_NoArgsPrintsHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Execute(context.Background(), nil, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout.String(), "typer") {
		t.Fatalf("expected help output, got %q", stdout.String())
	}
}

func TestExecute_HelpAliases(t *testing.T) {
	for _, alias := range []string{"help", "--help", "-h"} {
		var stdout, stderr bytes.Buffer
		if err := Execute(context.Background(), []string{alias}, strings.NewReader(""), &stdout, &stderr); err != nil {
			t.Fatalf("Execute %s: %v", alias, err)
		}
		if !strings.Contains(stdout.String(), "typer") {
			t.Fatalf("%s: expected help output, got %q", alias, stdout.String())
		}
	}
}

func TestExecute_UnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Execute(context.Background(), []string{"nope"}, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("expected unknown-command error, got %v", err)
	}
	if !strings.Contains(stderr.String(), "typer") {
		t.Fatalf("expected help printed to stderr, got %q", stderr.String())
	}
}

func TestRunStart_UnknownMode(t *testing.T) {
	var stdout bytes.Buffer
	err := runStart(context.Background(), []string{"--mode", "nonsense"}, strings.NewReader(""), &stdout)
	if err == nil {
		t.Fatal("expected error for unknown mode")
	}
	if !strings.Contains(err.Error(), "unsupported mode") {
		t.Fatalf("expected unsupported-mode error, got %v", err)
	}
}

func TestRunStart_SourceRejectedForNonQuoteMode(t *testing.T) {
	var stdout bytes.Buffer
	err := runStart(context.Background(), []string{"--mode", "words", "--source", "cache"}, strings.NewReader(""), &stdout)
	if err == nil {
		t.Fatal("expected error when --source combined with words mode")
	}
	if !strings.Contains(err.Error(), "--source is only valid") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunStart_HelpShortCircuits(t *testing.T) {
	var stdout bytes.Buffer
	err := runStart(context.Background(), []string{"-h"}, strings.NewReader(""), &stdout)
	if err != nil {
		t.Fatalf("runStart -h: %v", err)
	}
	if !strings.Contains(stdout.String(), "Run an interactive typing session") {
		t.Fatalf("expected start help, got %q", stdout.String())
	}
}

func TestExecute_ResetProgress(t *testing.T) {
	originalFactory := newHistoryStore
	t.Cleanup(func() { newHistoryStore = originalFactory })

	t.Run("confirm", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "history.json")
		if err := storage.NewHistoryStoreAt(path).Append(model.SessionResult{ID: "seed", StartedAt: time.Unix(1, 0)}); err != nil {
			t.Fatalf("Append: %v", err)
		}

		newHistoryStore = func() (*storage.HistoryStore, error) {
			return storage.NewHistoryStoreAt(path), nil
		}

		var stdout, stderr bytes.Buffer
		if err := Execute(context.Background(), []string{"--reset-progress"}, strings.NewReader("y\n"), &stdout, &stderr); err != nil {
			t.Fatalf("Execute --reset-progress: %v", err)
		}
		if !strings.Contains(stderr.String(), "Reset all local session history?") {
			t.Fatalf("expected prompt on stderr, got %q", stderr.String())
		}
		if !strings.Contains(stdout.String(), "Progress reset.") {
			t.Fatalf("expected reset confirmation, got %q", stdout.String())
		}
		after, err := storage.NewHistoryStoreAt(path).List(0)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(after) != 0 {
			t.Fatalf("expected history cleared, got %d session(s)", len(after))
		}
	})

	t.Run("cancel", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "history.json")
		if err := storage.NewHistoryStoreAt(path).Append(model.SessionResult{ID: "seed", StartedAt: time.Unix(1, 0)}); err != nil {
			t.Fatalf("Append: %v", err)
		}
		newHistoryStore = func() (*storage.HistoryStore, error) {
			return storage.NewHistoryStoreAt(path), nil
		}

		var stdout, stderr bytes.Buffer
		err := Execute(context.Background(), []string{"--reset-progress"}, strings.NewReader("n\n"), &stdout, &stderr)
		if err != nil {
			t.Fatalf("Execute --reset-progress: %v", err)
		}
		if !strings.Contains(stderr.String(), "Reset cancelled.") {
			t.Fatalf("expected cancel message on stderr, got %q", stderr.String())
		}
		if strings.Contains(stdout.String(), "Progress reset.") {
			t.Fatalf("expected no success output, got %q", stdout.String())
		}
		after, err := storage.NewHistoryStoreAt(path).List(0)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(after) != 1 || after[0].ID != "seed" {
			t.Fatalf("expected history unchanged, got %#v", after)
		}
	})
}

func TestExecute_ResetProgressRejectsExtraArgs(t *testing.T) {
	originalFactory := newHistoryStore
	t.Cleanup(func() { newHistoryStore = originalFactory })
	newHistoryStore = func() (*storage.HistoryStore, error) {
		return storage.NewHistoryStoreAt(filepath.Join(t.TempDir(), "history.json")), nil
	}

	var stdout, stderr bytes.Buffer
	err := Execute(context.Background(), []string{"--reset-progress", "extra"}, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for extra args")
	}
	if !strings.Contains(err.Error(), "does not take additional arguments") {
		t.Fatalf("unexpected error: %v", err)
	}
}
