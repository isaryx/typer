package cli

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"
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
	if err := Execute(context.Background(), nil, &stdout, &stderr); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout.String(), "typer") {
		t.Fatalf("expected help output, got %q", stdout.String())
	}
}

func TestExecute_HelpAliases(t *testing.T) {
	for _, alias := range []string{"help", "--help", "-h"} {
		var stdout, stderr bytes.Buffer
		if err := Execute(context.Background(), []string{alias}, &stdout, &stderr); err != nil {
			t.Fatalf("Execute %s: %v", alias, err)
		}
		if !strings.Contains(stdout.String(), "typer") {
			t.Fatalf("%s: expected help output, got %q", alias, stdout.String())
		}
	}
}

func TestExecute_UnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Execute(context.Background(), []string{"nope"}, &stdout, &stderr)
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
