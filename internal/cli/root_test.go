package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"typer/internal/model"
	"typer/internal/storage"
)

const (
	testFlagMode      = "--mode"
	testFlagStrict    = "--strict"
	testFlagLast      = "--last"
	testFlagReset     = "--reset-progress"
	testFlagWordsFile = "--words-file"
	testHistoryFile   = "history.json"
	unexpectedErrFmt  = "unexpected error: %v"
)

type presenceFlagsCase struct {
	name       string
	args       []string
	wantRest   []string
	wantStrict bool
	wantIndef  bool
	wantErr    bool
}

func assertExtractPresenceFlagsCase(t *testing.T, tc presenceFlagsCase) {
	t.Helper()

	gotRest, gotStrict, gotIndef, err := extractPresenceFlags(tc.args)
	if tc.wantErr {
		if err == nil {
			t.Fatalf("extractPresenceFlags: err = nil, want error")
		}
		return
	}
	if err != nil {
		t.Fatalf("extractPresenceFlags: %v", err)
	}
	if gotStrict != tc.wantStrict {
		t.Fatalf("strict = %v, want %v", gotStrict, tc.wantStrict)
	}
	if gotIndef != tc.wantIndef {
		t.Fatalf("indefinite = %v, want %v", gotIndef, tc.wantIndef)
	}
	if !reflect.DeepEqual(gotRest, tc.wantRest) {
		t.Fatalf("rest = %#v, want %#v", gotRest, tc.wantRest)
	}
}

func TestExtractPresenceFlags(t *testing.T) {
	tests := []presenceFlagsCase{
		{"empty", nil, nil, false, false, false},
		{"no flags", []string{testFlagMode, "words"}, []string{testFlagMode, "words"}, false, false, false},
		{"bare strict", []string{testFlagStrict}, nil, true, false, false},
		{"strict then flags", []string{testFlagStrict, testFlagMode, "words"}, []string{testFlagMode, "words"}, true, false, false},
		{"strict shorthand", []string{"-s", "-m", "words"}, []string{"-m", "words"}, true, false, false},
		{"bare indefinite", []string{"--indefinite"}, nil, false, true, false},
		{"indef shorthand", []string{"-i", "-m", "w"}, []string{"-m", "w"}, false, true, false},
		{"strict and indefinite", []string{"-s", "-i", "-w", "3"}, []string{"-w", "3"}, true, true, false},
		{"eq strict", []string{testFlagStrict + "=false"}, nil, false, false, true},
		{"eq indefinite", []string{"-i=true"}, nil, false, false, true},
		{"space true after strict", []string{testFlagStrict, "true"}, nil, false, false, true},
		{"word after strict ok", []string{testFlagStrict, "hello"}, []string{"hello"}, true, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertExtractPresenceFlagsCase(t, tt)
		})
	}
}

func TestExecuteNoArgsPrintsHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Execute(context.Background(), nil, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout.String(), "typer") {
		t.Fatalf("expected help output, got %q", stdout.String())
	}
}

func TestExecuteHelpAliases(t *testing.T) {
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

func TestExecuteCredits(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Execute(context.Background(), []string{"credits"}, strings.NewReader(""), &stdout, &stderr); err != nil {
		t.Fatalf("Execute credits: %v", err)
	}
	if !strings.Contains(stdout.String(), "Data Credits") {
		t.Fatalf("expected credits heading, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "type.fit") {
		t.Fatalf("expected quote API credit in output, got %q", stdout.String())
	}
}

func TestExecuteUnknownCommand(t *testing.T) {
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

func TestRunStartUnknownMode(t *testing.T) {
	var stdout bytes.Buffer
	err := runStart(context.Background(), []string{testFlagMode, "nonsense"}, strings.NewReader(""), &stdout)
	if err == nil {
		t.Fatal("expected error for unknown mode")
	}
	if !strings.Contains(err.Error(), "unsupported mode") {
		t.Fatalf("expected unsupported-mode error, got %v", err)
	}
}

func TestRunStartSourceRejectedForNonQuoteMode(t *testing.T) {
	var stdout bytes.Buffer
	err := runStart(context.Background(), []string{testFlagMode, "words", "--source", "cache"}, strings.NewReader(""), &stdout)
	if err == nil {
		t.Fatal("expected error when --source combined with words mode")
	}
	if !strings.Contains(err.Error(), "--source is only valid") {
		t.Fatalf(unexpectedErrFmt, err)
	}
}

func TestRunStartHelpShortCircuits(t *testing.T) {
	var stdout bytes.Buffer
	err := runStart(context.Background(), []string{"-h"}, strings.NewReader(""), &stdout)
	if err != nil {
		t.Fatalf("runStart -h: %v", err)
	}
	if !strings.Contains(stdout.String(), "Run an interactive typing session") {
		t.Fatalf("expected start help, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "(default quotes)") {
		t.Fatalf("expected quotes default in start help, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "(default seed)") {
		t.Fatalf("expected seed source default in start help, got %q", stdout.String())
	}
}

func TestExecuteResetProgressConfirm(t *testing.T) {
	originalFactory := newHistoryStore
	t.Cleanup(func() { newHistoryStore = originalFactory })

	path := filepath.Join(t.TempDir(), testHistoryFile)
	if err := storage.NewHistoryStoreAt(path).Append(model.SessionResult{ID: "seed", StartedAt: time.Unix(1, 0)}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	newHistoryStore = func() (*storage.HistoryStore, error) {
		return storage.NewHistoryStoreAt(path), nil
	}

	var stdout, stderr bytes.Buffer
	if err := Execute(context.Background(), []string{testFlagReset}, strings.NewReader("y\n"), &stdout, &stderr); err != nil {
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
}

func TestExecuteResetProgressCancel(t *testing.T) {
	originalFactory := newHistoryStore
	t.Cleanup(func() { newHistoryStore = originalFactory })

	dir := t.TempDir()
	path := filepath.Join(dir, testHistoryFile)
	if err := storage.NewHistoryStoreAt(path).Append(model.SessionResult{ID: "seed", StartedAt: time.Unix(1, 0)}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	newHistoryStore = func() (*storage.HistoryStore, error) {
		return storage.NewHistoryStoreAt(path), nil
	}

	var stdout, stderr bytes.Buffer
	err := Execute(context.Background(), []string{testFlagReset}, strings.NewReader("n\n"), &stdout, &stderr)
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
}

func TestExecuteResetProgressRejectsExtraArgs(t *testing.T) {
	originalFactory := newHistoryStore
	t.Cleanup(func() { newHistoryStore = originalFactory })
	newHistoryStore = func() (*storage.HistoryStore, error) {
		return storage.NewHistoryStoreAt(filepath.Join(t.TempDir(), testHistoryFile)), nil
	}

	var stdout, stderr bytes.Buffer
	err := Execute(context.Background(), []string{testFlagReset, "extra"}, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for extra args")
	}
	if !strings.Contains(err.Error(), "does not take additional arguments") {
		t.Fatalf(unexpectedErrFmt, err)
	}
}

func TestPrintMetricsTableIncludesAdjustedWPM(t *testing.T) {
	var out bytes.Buffer
	printMetricsTable(&out, "Session result", 72.0, 68.0, 64.0, 88.9, 91.2, 3, 27_000, false)
	got := out.String()
	if !strings.Contains(got, "Adjusted WPM") {
		t.Fatalf("expected Adjusted WPM row, got %q", got)
	}
	if !strings.Contains(got, "64.00") {
		t.Fatalf("expected adjusted value, got %q", got)
	}
}

func TestPrintMetricsTableSummaryIncludesAvgAdjustedWPM(t *testing.T) {
	var out bytes.Buffer
	printMetricsTable(&out, "Summary (2 completed sessions)", 72.0, 68.0, 64.0, 88.9, 91.2, 3, 27_000, true)
	got := out.String()
	if !strings.Contains(got, "Avg adjusted WPM") {
		t.Fatalf("expected Avg adjusted WPM row, got %q", got)
	}
	if !strings.Contains(got, "64.00") {
		t.Fatalf("expected avg adjusted value, got %q", got)
	}
}

func TestRunVersionPrintsVersion(t *testing.T) {
	var out bytes.Buffer
	if err := runVersion(&out); err != nil {
		t.Fatalf("runVersion: %v", err)
	}
	if !strings.HasPrefix(out.String(), "typer ") {
		t.Fatalf("expected version prefix, got %q", out.String())
	}
}

func TestRunSetRequiresAtLeastOneFlag(t *testing.T) {
	var out bytes.Buffer
	err := runSet(nil, &out)
	if err == nil {
		t.Fatal("expected runSet to fail without flags")
	}
	if !strings.Contains(err.Error(), "set requires --words-file and/or --passages-file") {
		t.Fatalf(unexpectedErrFmt, err)
	}
}

func TestRunSetRejectsMissingOrDirectoryPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	var out bytes.Buffer
	err := runSet([]string{testFlagWordsFile, filepath.Join(home, "missing.txt")}, &out)
	if err == nil {
		t.Fatal("expected missing file error")
	}
	if !strings.Contains(err.Error(), "invalid words file") {
		t.Fatalf(unexpectedErrFmt, err)
	}

	out.Reset()
	err = runSet([]string{"--passages-file", home}, &out)
	if err == nil {
		t.Fatal("expected directory rejection")
	}
	if !strings.Contains(err.Error(), "expected file, got directory") {
		t.Fatalf(unexpectedErrFmt, err)
	}
}

func TestRunSetSavesAbsolutePathsForBothFiles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	words := filepath.Join(home, "words.txt")
	passages := filepath.Join(home, "passages.txt")
	if err := os.WriteFile(words, []byte("alpha\nbeta\n"), 0o644); err != nil {
		t.Fatalf("write words: %v", err)
	}
	if err := os.WriteFile(passages, []byte("hello world\n"), 0o644); err != nil {
		t.Fatalf("write passages: %v", err)
	}

	var out bytes.Buffer
	if err := runSet([]string{testFlagWordsFile, words, "--passages-file", passages}, &out); err != nil {
		t.Fatalf("runSet: %v", err)
	}
	if !strings.Contains(out.String(), "Custom words file set to") || !strings.Contains(out.String(), "Custom passages file set to") {
		t.Fatalf("expected both confirmations, got %q", out.String())
	}

	store, err := storage.NewSettingsStore()
	if err != nil {
		t.Fatalf("NewSettingsStore: %v", err)
	}
	settings, err := store.Load()
	if err != nil {
		t.Fatalf("Load settings: %v", err)
	}
	wantWords, _ := filepath.Abs(words)
	wantPassages, _ := filepath.Abs(passages)
	if settings.WordsFile != wantWords {
		t.Fatalf("WordsFile = %q, want %q", settings.WordsFile, wantWords)
	}
	if settings.PassagesFile != wantPassages {
		t.Fatalf("PassagesFile = %q, want %q", settings.PassagesFile, wantPassages)
	}
}

func TestRunHistoryEmptyAndTruncatedListing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	store, err := storage.NewHistoryStore()
	if err != nil {
		t.Fatalf("NewHistoryStore: %v", err)
	}

	var out bytes.Buffer
	if err := runHistory(nil, &out); err != nil {
		t.Fatalf("runHistory empty: %v", err)
	}
	if !strings.Contains(out.String(), noHistoryMessage) {
		t.Fatalf("expected no-history message, got %q", out.String())
	}

	out.Reset()
	if err := store.Append(model.SessionResult{
		ID:        "1",
		StartedAt: time.Unix(100, 0),
		Prompt:    model.Prompt{Mode: model.ModeWords, Source: "seed"},
		Metrics:   model.SessionMetrics{NetWPM: 40, Accuracy: 95, Errors: 1},
	}); err != nil {
		t.Fatalf("Append 1: %v", err)
	}
	if err := store.Append(model.SessionResult{
		ID:        "2",
		StartedAt: time.Unix(200, 0),
		Prompt:    model.Prompt{Mode: model.ModeQuote, Source: "cache"},
		Metrics:   model.SessionMetrics{NetWPM: 55, Accuracy: 98, Errors: 0},
	}); err != nil {
		t.Fatalf("Append 2: %v", err)
	}

	if err := runHistory([]string{testFlagLast, "1"}, &out); err != nil {
		t.Fatalf("runHistory --last 1: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Showing 1 session(s).") {
		t.Fatalf("expected showing count, got %q", got)
	}
	if !strings.Contains(got, "mode=quote") || !strings.Contains(got, "source=cache") {
		t.Fatalf("expected newest session fields, got %q", got)
	}
	if strings.Contains(got, "mode=words") {
		t.Fatalf("expected truncated output to newest one, got %q", got)
	}
}

func TestRunStatsEmptyAndSummaryOutput(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	store, err := storage.NewHistoryStore()
	if err != nil {
		t.Fatalf("NewHistoryStore: %v", err)
	}

	var out bytes.Buffer
	if err := runStats(nil, &out); err != nil {
		t.Fatalf("runStats empty: %v", err)
	}
	if !strings.Contains(out.String(), noHistoryMessage) {
		t.Fatalf("expected no-history message, got %q", out.String())
	}

	out.Reset()
	if err := store.Append(model.SessionResult{
		ID:        "1",
		StartedAt: time.Unix(100, 0),
		Prompt:    model.Prompt{Content: "ab", Source: "seed"},
		TypedText: "ab",
		Metrics: model.SessionMetrics{
			GrossWPM:    50,
			NetWPM:      49,
			AdjustedWPM: 48,
			Accuracy:    99,
			Errors:      0,
		},
	}); err != nil {
		t.Fatalf("Append 1: %v", err)
	}

	if err := runStats([]string{testFlagLast, "1"}, &out); err != nil {
		t.Fatalf("runStats --last 1: %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Stats for last 1 session(s).",
		"Avg Gross WPM",
		"Avg Net WPM",
		"Avg Adjusted WPM",
		"Avg Accuracy",
		"Top Mistyped Target Characters:",
		"  - none",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in stats output: %q", want, got)
		}
	}
}

func TestHelpPrintersIncludeUsage(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		var out bytes.Buffer
		printSetHelp(&out)
		got := out.String()
		if !strings.Contains(got, "typer set") || !strings.Contains(got, "--words-file") {
			t.Fatalf("unexpected set help: %q", got)
		}
	})

	t.Run("history", func(t *testing.T) {
		var out bytes.Buffer
		printHistoryHelp(&out)
		got := out.String()
		if !strings.Contains(got, "typer history") || !strings.Contains(got, testFlagLast) {
			t.Fatalf("unexpected history help: %q", got)
		}
	})

	t.Run("stats", func(t *testing.T) {
		var out bytes.Buffer
		printStatsHelp(&out)
		got := out.String()
		if !strings.Contains(got, "typer stats") || !strings.Contains(got, testFlagLast) {
			t.Fatalf("unexpected stats help: %q", got)
		}
	})
}
