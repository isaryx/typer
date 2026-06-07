package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"typer/internal/model"
	"typer/internal/storage"
)

const (
	testFlagMode            = "--mode"
	testFlagStrict          = "--strict"
	testFlagLast            = "--last"
	testFlagReset           = "--reset-progress"
	testFlagWordsFile       = "--words-file"
	testHistoryFile         = "history.json"
	unexpectedErrFmt        = "unexpected error: %v"
	expectedPositionalError = "expected error for positional args"
)

type presenceFlagsCase struct {
	name           string
	args           []string
	wantRest       []string
	wantStrict     bool
	wantIndef      bool
	wantFingerHint bool
	wantErr        bool
}

func setupTestHistoryStoreFactory(t *testing.T, seed ...model.SessionResult) string {
	t.Helper()

	originalFactory := newHistoryStore
	t.Cleanup(func() { newHistoryStore = originalFactory })

	path := filepath.Join(t.TempDir(), testHistoryFile)
	store := storage.NewHistoryStoreAt(path)
	for i, result := range seed {
		if err := store.Append(result); err != nil {
			t.Fatalf("Append %d: %v", i+1, err)
		}
	}
	newHistoryStore = func() (*storage.HistoryStore, error) {
		return storage.NewHistoryStoreAt(path), nil
	}
	return path
}

func setTestUserDirs(t *testing.T, home string) {
	t.Helper()

	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
		t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
		t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	}
}

func assertExtractPresenceFlagsCase(t *testing.T, tc presenceFlagsCase) {
	t.Helper()

	gotRest, gotStrict, gotIndef, gotFingerHint, err := extractPresenceFlags(tc.args)
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
	if gotFingerHint != tc.wantFingerHint {
		t.Fatalf("fingerHint = %v, want %v", gotFingerHint, tc.wantFingerHint)
	}
	if !reflect.DeepEqual(gotRest, tc.wantRest) {
		t.Fatalf("rest = %#v, want %#v", gotRest, tc.wantRest)
	}
}

func TestExtractPresenceFlags(t *testing.T) {
	tests := []presenceFlagsCase{
		{
			name:           "empty",
			args:           nil,
			wantRest:       nil,
			wantStrict:     false,
			wantIndef:      false,
			wantFingerHint: false,
			wantErr:        false,
		},
		{
			name:           "no flags",
			args:           []string{testFlagMode, "words"},
			wantRest:       []string{testFlagMode, "words"},
			wantStrict:     false,
			wantIndef:      false,
			wantFingerHint: false,
			wantErr:        false,
		},
		{
			name:           "bare strict",
			args:           []string{testFlagStrict},
			wantRest:       nil,
			wantStrict:     true,
			wantIndef:      false,
			wantFingerHint: false,
			wantErr:        false,
		},
		{
			name:           "strict then flags",
			args:           []string{testFlagStrict, testFlagMode, "words"},
			wantRest:       []string{testFlagMode, "words"},
			wantStrict:     true,
			wantIndef:      false,
			wantFingerHint: false,
			wantErr:        false,
		},
		{
			name:           "strict shorthand",
			args:           []string{"-s", "-m", "words"},
			wantRest:       []string{"-m", "words"},
			wantStrict:     true,
			wantIndef:      false,
			wantFingerHint: false,
			wantErr:        false,
		},
		{
			name:           "dash long strict not accepted use double dash or short",
			args:           []string{"-strict", "-m", "words"},
			wantRest:       []string{"-strict", "-m", "words"},
			wantStrict:     false,
			wantIndef:      false,
			wantFingerHint: false,
			wantErr:        false,
		},
		{
			name:           "bare indefinite",
			args:           []string{"--indefinite"},
			wantRest:       nil,
			wantStrict:     false,
			wantIndef:      true,
			wantFingerHint: false,
			wantErr:        false,
		},
		{
			name:           "indef shorthand",
			args:           []string{"-i", "-m", "w"},
			wantRest:       []string{"-m", "w"},
			wantStrict:     false,
			wantIndef:      true,
			wantFingerHint: false,
			wantErr:        false,
		},
		{
			name:           "dash long indefinite not accepted use double dash or short",
			args:           []string{"-indefinite"},
			wantRest:       []string{"-indefinite"},
			wantStrict:     false,
			wantIndef:      false,
			wantFingerHint: false,
			wantErr:        false,
		},
		{
			name:           "strict and indefinite",
			args:           []string{"-s", "-i", "-w", "3"},
			wantRest:       []string{"-w", "3"},
			wantStrict:     true,
			wantIndef:      true,
			wantFingerHint: false,
			wantErr:        false,
		},
		{
			name:           "bare finger-hint",
			args:           []string{"--finger-hint"},
			wantRest:       nil,
			wantStrict:     false,
			wantIndef:      false,
			wantFingerHint: true,
			wantErr:        false,
		},
		{
			name:           "finger-hint shorthand",
			args:           []string{"-f", "-m", "words"},
			wantRest:       []string{"-m", "words"},
			wantStrict:     false,
			wantIndef:      false,
			wantFingerHint: true,
			wantErr:        false,
		},
		{
			name:           "eq strict",
			args:           []string{testFlagStrict + "=false"},
			wantRest:       nil,
			wantStrict:     false,
			wantIndef:      false,
			wantFingerHint: false,
			wantErr:        true,
		},
		{
			name:           "eq indefinite",
			args:           []string{"-i=true"},
			wantRest:       nil,
			wantStrict:     false,
			wantIndef:      false,
			wantFingerHint: false,
			wantErr:        true,
		},
		{
			name:           "eq finger-hint",
			args:           []string{"--finger-hint=false"},
			wantRest:       nil,
			wantStrict:     false,
			wantIndef:      false,
			wantFingerHint: false,
			wantErr:        true,
		},
		{
			name:           "space true after strict",
			args:           []string{testFlagStrict, "true"},
			wantRest:       nil,
			wantStrict:     false,
			wantIndef:      false,
			wantFingerHint: false,
			wantErr:        true,
		},
		{
			name:           "word after strict ok",
			args:           []string{testFlagStrict, "hello"},
			wantRest:       []string{"hello"},
			wantStrict:     true,
			wantIndef:      false,
			wantFingerHint: false,
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertExtractPresenceFlagsCase(t, tt)
		})
	}
}

func TestExecuteNoArgsShowsMenu(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := Execute(context.Background(), nil, strings.NewReader("\n"), &stdout, &stderr); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	got := stdout.String()
	for _, want := range []string{"Choose a command", "start", "train", "play", "history", "stats"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected menu to include %q, got %q", want, got)
		}
	}
}

func TestExecuteHelpAliases(t *testing.T) {
	for _, alias := range []string{"--help", "-h"} {
		var stdout, stderr bytes.Buffer
		if err := Execute(context.Background(), []string{alias}, strings.NewReader(""), &stdout, &stderr); err != nil {
			t.Fatalf("Execute %s: %v", alias, err)
		}
		if !strings.Contains(stdout.String(), "typer") {
			t.Fatalf("%s: expected help output, got %q", alias, stdout.String())
		}
	}
}

func TestExecuteVersionAliases(t *testing.T) {
	for _, alias := range []string{"--version", "-v"} {
		var stdout, stderr bytes.Buffer
		if err := Execute(context.Background(), []string{alias}, strings.NewReader(""), &stdout, &stderr); err != nil {
			t.Fatalf("Execute %s: %v", alias, err)
		}
		if !strings.HasPrefix(stdout.String(), "typer ") {
			t.Fatalf("%s: expected version line, got %q", alias, stdout.String())
		}
	}
}

func TestExecuteRejectsFlagFirstWithoutStart(t *testing.T) {
	for _, first := range []string{"-m", "--mode", "-w", "--strict"} {
		var stdout, stderr bytes.Buffer
		err := Execute(context.Background(), []string{first, "words"}, strings.NewReader(""), &stdout, &stderr)
		if err == nil {
			t.Fatalf("Execute %s: expected error", first)
		}
		if !strings.Contains(err.Error(), "unknown command") {
			t.Fatalf("%s: expected unknown command, got %v", first, err)
		}
		if !errors.Is(err, ErrUsage) {
			t.Fatalf("%s: expected ErrUsage, got %v", first, err)
		}
	}
}

func TestExecuteRejectsHelpAndVersionAsCommands(t *testing.T) {
	for _, cmd := range []string{"help", "version"} {
		var stdout, stderr bytes.Buffer
		err := Execute(context.Background(), []string{cmd}, strings.NewReader(""), &stdout, &stderr)
		if err == nil {
			t.Fatalf("Execute %s: expected error", cmd)
		}
		if !strings.Contains(err.Error(), "unknown command") {
			t.Fatalf("%s: expected unknown command, got %v", cmd, err)
		}
		if !errors.Is(err, ErrUsage) {
			t.Fatalf("%s: expected ErrUsage, got %v", cmd, err)
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
	if !strings.Contains(stdout.String(), "zenquotes") {
		t.Fatalf("expected zenquotes credit in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "type.fit") {
		t.Fatalf("expected type.fit credit in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "oklog/ulid") {
		t.Fatalf("expected ULID library credit in output, got %q", stdout.String())
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
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("expected ErrUsage, got %v", err)
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
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("expected ErrUsage, got %v", err)
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
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("expected ErrUsage, got %v", err)
	}
}

func TestRunStartRejectsWordsAboveMax(t *testing.T) {
	var stdout bytes.Buffer
	err := runStart(context.Background(), []string{"-m", "words", "-w", strconv.Itoa(model.MaxWordsPerPrompt + 1)}, strings.NewReader(""), &stdout)
	if err == nil {
		t.Fatal("expected error for words above max")
	}
	if !strings.Contains(err.Error(), "cannot exceed") {
		t.Fatalf(unexpectedErrFmt, err)
	}
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("expected ErrUsage, got %v", err)
	}
}

func TestRunHistoryRejectsLastOutOfRange(t *testing.T) {
	var stdout bytes.Buffer
	if err := runHistory([]string{testFlagLast, "0"}, &stdout); err == nil {
		t.Fatal("expected error for --last 0")
	} else if !errors.Is(err, ErrUsage) {
		t.Fatalf("expected ErrUsage, got %v", err)
	}
	stdout.Reset()
	if err := runHistory([]string{testFlagLast, strconv.Itoa(model.MaxRetainedHistorySessions + 1)}, &stdout); err == nil {
		t.Fatal("expected error for --last above max")
	} else if !errors.Is(err, ErrUsage) {
		t.Fatalf("expected ErrUsage, got %v", err)
	}
}

func TestRunStartHelpShortCircuits(t *testing.T) {
	var stdout bytes.Buffer
	err := runStart(context.Background(), []string{"-h"}, strings.NewReader(""), &stdout)
	if err != nil {
		t.Fatalf("runStart -h: %v", err)
	}
	if !strings.Contains(stdout.String(), "Free practice session") {
		t.Fatalf("expected start help, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "(default quotes)") {
		t.Fatalf("expected quotes default in start help, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "(default remote)") {
		t.Fatalf("expected remote source default in start help, got %q", stdout.String())
	}
}

func TestRunStartRejectsPositionalArgs(t *testing.T) {
	var stdout bytes.Buffer
	err := runStart(context.Background(), []string{"--strict", "hello"}, strings.NewReader(""), &stdout)
	if err == nil {
		t.Fatal(expectedPositionalError)
	}
	if !strings.Contains(err.Error(), "start does not take positional arguments") {
		t.Fatalf(unexpectedErrFmt, err)
	}
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("expected ErrUsage, got %v", err)
	}
}

func TestExecuteResetProgressConfirm(t *testing.T) {
	path := setupTestHistoryStoreFactory(t, model.SessionResult{ID: "seed", StartedAt: time.Unix(1, 0)})

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
	path := setupTestHistoryStoreFactory(t, model.SessionResult{ID: "seed", StartedAt: time.Unix(1, 0)})

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
	setupTestHistoryStoreFactory(t)

	var stdout, stderr bytes.Buffer
	err := Execute(context.Background(), []string{testFlagReset, "extra"}, strings.NewReader(""), &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for extra args")
	}
	if !strings.Contains(err.Error(), "does not take additional arguments") {
		t.Fatalf(unexpectedErrFmt, err)
	}
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("expected ErrUsage, got %v", err)
	}
}

func TestPrintMetricsTableIncludesAdjustedWPM(t *testing.T) {
	var out bytes.Buffer
	printMetricsTable(&out, "Session Stats", 72.0, 68.0, 64.0, 88.9, 91.2, 3, 27_000, false)
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

func TestPrintMetricsTableIncludesConsistencyWhenZero(t *testing.T) {
	var out bytes.Buffer
	printMetricsTable(&out, "Session Stats", 50.0, 45.0, 40.0, 90.0, 0, 2, 8_000, false)
	got := out.String()
	if !strings.Contains(got, "Pace stability") {
		t.Fatalf("expected Pace stability row, got %q", got)
	}
	if !strings.Contains(got, "0.00") {
		t.Fatalf("expected zero pace stability value, got %q", got)
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
	if !strings.Contains(err.Error(), "set requires --words-file, --passages-file, --show-hint, --input-position, and/or --quote-source") {
		t.Fatalf(unexpectedErrFmt, err)
	}
}

func TestRunSetRejectsPositionalArgs(t *testing.T) {
	var out bytes.Buffer
	err := runSet([]string{testFlagWordsFile, "words.txt", "extra"}, &out)
	if err == nil {
		t.Fatal(expectedPositionalError)
	}
	if !strings.Contains(err.Error(), "set does not take positional arguments") {
		t.Fatalf(unexpectedErrFmt, err)
	}
}

func TestRunSetRejectsMissingOrDirectoryPath(t *testing.T) {
	home := t.TempDir()
	setTestUserDirs(t, home)

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
	setTestUserDirs(t, home)

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

func TestRunSetShowHintOnlyPersists(t *testing.T) {
	home := t.TempDir()
	setTestUserDirs(t, home)

	var out bytes.Buffer
	if err := runSet([]string{"--show-hint", "off"}, &out); err != nil {
		t.Fatalf("runSet: %v", err)
	}
	if !strings.Contains(out.String(), "Typing hint: off") {
		t.Fatalf("expected confirmation, got %q", out.String())
	}

	store, err := storage.NewSettingsStore()
	if err != nil {
		t.Fatalf("NewSettingsStore: %v", err)
	}
	settings, err := store.Load()
	if err != nil {
		t.Fatalf("Load settings: %v", err)
	}
	if settings.ShowHint == nil || *settings.ShowHint {
		t.Fatalf("expected show_hint false, got %#v", settings.ShowHint)
	}

	out.Reset()
	if err := runSet([]string{"--show-hint", "on"}, &out); err != nil {
		t.Fatalf("runSet on: %v", err)
	}
	settings, err = store.Load()
	if err != nil {
		t.Fatalf("Load settings: %v", err)
	}
	if settings.ShowHint == nil || !*settings.ShowHint {
		t.Fatalf("expected show_hint true, got %#v", settings.ShowHint)
	}
}

func TestRunSetShowHintRejectsInvalid(t *testing.T) {
	var out bytes.Buffer
	err := runSet([]string{"--show-hint", "maybe"}, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "set: --show-hint must be one of on, off, yes, no, true, false, 1, 0") {
		t.Fatalf(unexpectedErrFmt, err)
	}
}

func TestRunSetInputPositionOnlyPersists(t *testing.T) {
	home := t.TempDir()
	setTestUserDirs(t, home)

	var out bytes.Buffer
	if err := runSet([]string{"--input-position", "tc"}, &out); err != nil {
		t.Fatalf("runSet: %v", err)
	}
	if !strings.Contains(out.String(), "Input position: top-center") {
		t.Fatalf("expected confirmation, got %q", out.String())
	}
	store, err := storage.NewSettingsStore()
	if err != nil {
		t.Fatalf("NewSettingsStore: %v", err)
	}
	settings, err := store.Load()
	if err != nil {
		t.Fatalf("Load settings: %v", err)
	}
	if settings.InputPosition != "top-center" {
		t.Fatalf("InputPosition = %q, want top-center", settings.InputPosition)
	}
}

func TestRunSetInputPositionRejectsInvalid(t *testing.T) {
	var out bytes.Buffer
	err := runSet([]string{"--input-position", "sideways"}, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "input position") {
		t.Fatalf(unexpectedErrFmt, err)
	}
}

func TestRunHistoryEmptyAndTruncatedListing(t *testing.T) {
	home := t.TempDir()
	setTestUserDirs(t, home)

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

func TestRunHistoryRejectsPositionalArgs(t *testing.T) {
	var out bytes.Buffer
	err := runHistory([]string{"extra"}, &out)
	if err == nil {
		t.Fatal(expectedPositionalError)
	}
	if !strings.Contains(err.Error(), "history does not take positional arguments") {
		t.Fatalf(unexpectedErrFmt, err)
	}
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("expected ErrUsage, got %v", err)
	}
}

func TestRunReplayRequiresSelector(t *testing.T) {
	var out bytes.Buffer
	err := runReplay(context.Background(), nil, strings.NewReader(""), &out)
	if err == nil {
		t.Fatal("expected error when replay has no --last, --nth, or --id")
	}
	if !strings.Contains(err.Error(), "replay requires") {
		t.Fatalf(unexpectedErrFmt, err)
	}
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("expected ErrUsage, got %v", err)
	}
}

func TestTypingTraceUserNote(t *testing.T) {
	if got := typingTraceUserNote(model.SessionResult{}); got == "" {
		t.Fatal("expected note when trace empty")
	}
	if got := typingTraceUserNote(model.SessionResult{TypingTrace: []model.ReplayEvent{{AtMS: 0, Kind: model.ReplayEventKey, Rune: "a"}}}); got != "" {
		t.Fatalf("expected no note when trace present, got %q", got)
	}
}

func TestPrintReplayComparisonIncludesAccuracyAndErrors(t *testing.T) {
	var out bytes.Buffer
	printReplayComparison(&out,
		model.SessionResult{
			Metrics:   model.SessionMetrics{NetWPM: 40, Accuracy: 98.5, Errors: 2},
			ElapsedMS: 10000,
		},
		model.SessionResult{
			Metrics:   model.SessionMetrics{NetWPM: 42, Accuracy: 99.2, Errors: 1},
			ElapsedMS: 9500,
		},
	)
	got := out.String()
	for _, want := range []string{
		"Net WPM:",
		"+2.00",
		"Accuracy:",
		"+0.70%",
		"Errors:",
		"-1",
		"Time:",
		"faster",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in:\n%s", want, got)
		}
	}
}

func TestRunStatsEmptyAndSummaryOutput(t *testing.T) {
	home := t.TempDir()
	setTestUserDirs(t, home)

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

func TestRunStatsRejectsPositionalArgs(t *testing.T) {
	var out bytes.Buffer
	err := runStats([]string{"extra"}, &out)
	if err == nil {
		t.Fatal(expectedPositionalError)
	}
	if !strings.Contains(err.Error(), "stats does not take positional arguments") {
		t.Fatalf(unexpectedErrFmt, err)
	}
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("expected ErrUsage, got %v", err)
	}
}

func TestHelpPrintersIncludeUsage(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		var out bytes.Buffer
		printSetHelp(&out)
		got := out.String()
		if !strings.Contains(got, "typer set") || !strings.Contains(got, "--words-file") || !strings.Contains(got, "--input-position") {
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

	t.Run("replay", func(t *testing.T) {
		var out bytes.Buffer
		printReplayHelp(&out)
		got := out.String()
		if !strings.Contains(got, "typer replay") || !strings.Contains(got, "--nth") ||
			!strings.Contains(got, "-l, --last") {
			t.Fatalf("unexpected replay help: %q", got)
		}
	})
}
