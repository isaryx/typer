package storage

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"typer/internal/model"
)

const (
	historyJSONFile = "history.json"
	errListFmt      = "List: %v"
	errAppendFmt    = "Append: %v"
)

func newHistoryStoreAt(t *testing.T) *HistoryStore {
	t.Helper()
	return NewHistoryStoreAt(filepath.Join(t.TempDir(), historyJSONFile))
}

func TestNewHistoryStoreUsesConfigDir(t *testing.T) {
	home := t.TempDir()
	setTestUserDirs(t, home)
	want := filepath.Join(typerConfigDirForTestHome(home), historyJSONFile)

	store, err := NewHistoryStore()
	if err != nil {
		t.Fatalf("NewHistoryStore: %v", err)
	}
	if store.path != want {
		t.Fatalf("history path = %q, want %q", store.path, want)
	}
}

func TestHistoryStoreEmptyList(t *testing.T) {
	store := newHistoryStoreAt(t)

	got, err := store.List(10)
	if err != nil {
		t.Fatalf(errListFmt, err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty list, got %d", len(got))
	}
}

func TestHistoryStoreAppendAndListReverseOrder(t *testing.T) {
	store := newHistoryStoreAt(t)

	first := model.SessionResult{ID: "1", StartedAt: time.Unix(100, 0)}
	second := model.SessionResult{ID: "2", StartedAt: time.Unix(200, 0)}
	third := model.SessionResult{ID: "3", StartedAt: time.Unix(300, 0)}
	for _, r := range []model.SessionResult{first, second, third} {
		if err := store.Append(r); err != nil {
			t.Fatalf("Append %s: %v", r.ID, err)
		}
	}

	got, err := store.List(0) // 0 = unlimited
	if err != nil {
		t.Fatalf(errListFmt, err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(got))
	}
	// Reverse chronological (newest first).
	wantIDs := []string{"3", "2", "1"}
	for i, w := range wantIDs {
		if got[i].ID != w {
			t.Fatalf("idx %d: want ID %s, got %s", i, w, got[i].ID)
		}
	}
}

func TestHistoryStoreAppendEnforcesMaxSessions(t *testing.T) {
	// Seed the file directly to avoid O(n^2) full-file rewrites in a loop.
	store := newHistoryStoreAt(t)
	seed := make([]model.SessionResult, model.MaxRetainedHistorySessions)
	for i := range seed {
		seed[i] = model.SessionResult{ID: "seed"}
	}
	if err := writeJSONFileAtomic(store.path, model.HistoryFile{Version: historyVersion, Sessions: seed}); err != nil {
		t.Fatalf("seed history: %v", err)
	}

	// Appending past the cap must evict the oldest entries, not grow the log.
	for i := 0; i < 3; i++ {
		if err := store.Append(model.SessionResult{ID: "new"}); err != nil {
			t.Fatalf(errAppendFmt, err)
		}
	}

	all, err := store.List(0)
	if err != nil {
		t.Fatalf(errListFmt, err)
	}
	if len(all) != model.MaxRetainedHistorySessions {
		t.Fatalf("expected retained count %d, got %d", model.MaxRetainedHistorySessions, len(all))
	}
	// The three newest must be the appended ones (List returns newest first).
	for i := 0; i < 3; i++ {
		if all[i].ID != "new" {
			t.Fatalf("newest[%d] ID = %q, want %q", i, all[i].ID, "new")
		}
	}
}

func TestHistoryStoreWritesTightPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-style file modes are not meaningful on Windows")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, historyJSONFile)
	store := NewHistoryStoreAt(path)
	if err := store.Append(model.SessionResult{ID: "1"}); err != nil {
		t.Fatalf(errAppendFmt, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if mode := info.Mode().Perm(); mode != configFilePerm {
		t.Fatalf("history.json perm = %o, want %o", mode, configFilePerm)
	}
}

func TestHistoryStoreListTruncatesToLast(t *testing.T) {
	store := newHistoryStoreAt(t)
	for i := 0; i < 5; i++ {
		if err := store.Append(model.SessionResult{ID: string(rune('A' + i))}); err != nil {
			t.Fatalf(errAppendFmt, err)
		}
	}
	got, err := store.List(2)
	if err != nil {
		t.Fatalf(errListFmt, err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(got))
	}
	if got[0].ID != "E" || got[1].ID != "D" {
		t.Fatalf("expected [E,D] newest first, got [%s,%s]", got[0].ID, got[1].ID)
	}
}

func TestHistoryStoreResetClearsSessions(t *testing.T) {
	store := newHistoryStoreAt(t)
	if err := store.Append(model.SessionResult{ID: "1"}); err != nil {
		t.Fatalf(errAppendFmt, err)
	}
	if err := store.Append(model.SessionResult{ID: "2"}); err != nil {
		t.Fatalf(errAppendFmt, err)
	}

	if err := store.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	got, err := store.List(0)
	if err != nil {
		t.Fatalf(errListFmt, err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty list after reset, got %d", len(got))
	}
}

func TestHistoryStoreReadMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), historyJSONFile)
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatalf("write malformed history: %v", err)
	}
	store := NewHistoryStoreAt(path)
	if _, err := store.read(); err == nil {
		t.Fatal("expected malformed history error")
	}
}

func TestHistoryStoreReadNormalizesLegacyVersion(t *testing.T) {
	store := newHistoryStoreAt(t)
	if err := writeJSONFileAtomic(store.path, model.HistoryFile{
		Version:  0,
		Sessions: []model.SessionResult{{ID: "legacy"}},
	}); err != nil {
		t.Fatalf("seed legacy history: %v", err)
	}

	got, err := store.read()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got.Version != historyVersion {
		t.Fatalf("Version = %d, want %d", got.Version, historyVersion)
	}
	if len(got.Sessions) != 1 || got.Sessions[0].ID != "legacy" {
		t.Fatalf("unexpected sessions: %#v", got.Sessions)
	}
}

func TestHistoryStoreGetByID(t *testing.T) {
	store := newHistoryStoreAt(t)
	if err := store.Append(model.SessionResult{ID: "a"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(model.SessionResult{ID: "b"}); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetByID("a")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "a" {
		t.Fatalf("got %q", got.ID)
	}
	if _, err := store.GetByID("missing"); err == nil {
		t.Fatal("expected error")
	} else if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound: %v", err)
	}
}

func TestHistoryStoreNthNewest(t *testing.T) {
	store := newHistoryStoreAt(t)
	for _, id := range []string{"1", "2", "3"} {
		if err := store.Append(model.SessionResult{ID: id}); err != nil {
			t.Fatal(err)
		}
	}
	first, err := store.NthNewest(1)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != "3" {
		t.Fatalf("newest = %q, want 3", first.ID)
	}
	second, err := store.NthNewest(2)
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != "2" {
		t.Fatalf("2nd = %q, want 2", second.ID)
	}
	if _, err := store.NthNewest(10); err == nil {
		t.Fatal("expected error for out-of-range nth")
	} else if !errors.Is(err, ErrInsufficientHistory) {
		t.Fatalf("expected ErrInsufficientHistory: %v", err)
	}
	if _, err := store.NthNewest(0); err == nil {
		t.Fatal("expected error for nth < 1")
	} else if !errors.Is(err, ErrNthOutOfRange) {
		t.Fatalf("expected ErrNthOutOfRange: %v", err)
	}
}

func TestHistoryStoreSessionsWithContentHash(t *testing.T) {
	store := newHistoryStoreAt(t)
	h := model.PromptContentHash("same text")
	if err := store.Append(model.SessionResult{
		ID:          "a",
		ContentHash: h,
		Prompt:      model.Prompt{Content: "same text"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(model.SessionResult{
		ID:          "b",
		ContentHash: "",
		Prompt:      model.Prompt{Content: "same text"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(model.SessionResult{
		ID:          "c",
		ContentHash: model.PromptContentHash("other"),
		Prompt:      model.Prompt{Content: "other"},
	}); err != nil {
		t.Fatal(err)
	}
	list, err := store.SessionsWithContentHash(h)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 sessions, got %d", len(list))
	}
}

func TestHistoryStoreBestSessionForGhost(t *testing.T) {
	store := newHistoryStoreAt(t)
	text := "ghost pick"
	h := model.PromptContentHash(text)
	trace := []model.ReplayEvent{{AtMS: 0, Kind: model.ReplayEventKey, Rune: "a"}}

	if err := store.Append(model.SessionResult{
		ID:          "weak",
		ContentHash: h,
		Prompt:      model.Prompt{Content: text},
		TypingTrace: trace,
		Metrics:     model.SessionMetrics{NetWPM: 30},
		StartedAt:   time.Unix(100, 0),
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(model.SessionResult{
		ID:          "best",
		ContentHash: h,
		Prompt:      model.Prompt{Content: text},
		TypingTrace: trace,
		Metrics:     model.SessionMetrics{NetWPM: 80},
		StartedAt:   time.Unix(200, 0),
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(model.SessionResult{
		ID:          "no_trace",
		ContentHash: h,
		Prompt:      model.Prompt{Content: text},
		TypingTrace: nil,
		Metrics:     model.SessionMetrics{NetWPM: 999},
	}); err != nil {
		t.Fatal(err)
	}

	got, err := store.BestSessionForGhost(h)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "best" {
		t.Fatalf("want best id, got %q", got.ID)
	}

	if _, err := store.BestSessionForGhost(model.PromptContentHash("unknown")); !errors.Is(err, ErrNoGhostCandidate) {
		t.Fatalf("want ErrNoGhostCandidate, got %v", err)
	}
}

func TestHistoryStoreCacheReusesMemoryAcrossCalls(t *testing.T) {
	store := newHistoryStoreAt(t)
	if err := store.Append(model.SessionResult{ID: "1"}); err != nil {
		t.Fatal(err)
	}
	if store.cache == nil {
		t.Fatal("expected cache after append")
	}
	cachePtr := store.cache
	if _, err := store.List(10); err != nil {
		t.Fatal(err)
	}
	if store.cache != cachePtr {
		t.Fatal("List should not reload from disk")
	}
	if _, err := store.GetByID("1"); err != nil {
		t.Fatal(err)
	}
	if store.cache != cachePtr {
		t.Fatal("GetByID should not reload from disk")
	}
}

func TestHistoryStoreSidecarTraceRoundTrip(t *testing.T) {
	store := newHistoryStoreAt(t)
	trace := []model.ReplayEvent{{AtMS: 0, Kind: model.ReplayEventKey, Rune: "x"}}
	if err := store.Append(model.SessionResult{
		ID:          "s1",
		TypingTrace: trace,
		Prompt:      model.Prompt{Content: "hello"},
	}); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetByID("s1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.TypingTrace) != 1 || got.TypingTrace[0].Rune != "x" {
		t.Fatalf("trace = %#v, want one key event", got.TypingTrace)
	}
	reloaded := NewHistoryStoreAt(store.path)
	sess, err := reloaded.GetByID("s1")
	if err != nil {
		t.Fatal(err)
	}
	if len(sess.TypingTrace) != 1 {
		t.Fatalf("reloaded trace = %#v", sess.TypingTrace)
	}
}

func TestHistoryStoreListOmitsTraces(t *testing.T) {
	store := newHistoryStoreAt(t)
	trace := []model.ReplayEvent{{AtMS: 0, Kind: model.ReplayEventKey, Rune: "a"}}
	if err := store.Append(model.SessionResult{ID: "s1", TypingTrace: trace}); err != nil {
		t.Fatal(err)
	}
	list, err := store.List(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || len(list[0].TypingTrace) != 0 {
		t.Fatalf("List should omit traces, got %#v", list[0].TypingTrace)
	}
}

func TestHistoryStoreMigratesInlineTracesToSidecar(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, historyJSONFile)
	trace := []model.ReplayEvent{{AtMS: 1, Kind: model.ReplayEventKey, Rune: "m"}}
	if err := writeJSONFileAtomic(path, model.HistoryFile{
		Version: historyVersion,
		Sessions: []model.SessionResult{{
			ID:          "legacy",
			TypingTrace: trace,
			Prompt:      model.Prompt{Content: "text"},
		}},
	}); err != nil {
		t.Fatal(err)
	}
	store := NewHistoryStoreAt(path)
	if err := store.Append(model.SessionResult{ID: "new"}); err != nil {
		t.Fatal(err)
	}
	file, err := store.read()
	if err != nil {
		t.Fatal(err)
	}
	for _, sess := range file.Sessions {
		if len(sess.TypingTrace) != 0 {
			t.Fatalf("session %q still has inline trace", sess.ID)
		}
	}
	got, err := store.GetByID("legacy")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.TypingTrace) != 1 || got.TypingTrace[0].Rune != "m" {
		t.Fatalf("legacy trace = %#v", got.TypingTrace)
	}
}

func TestHistoryStoreGhostIndexAfterRetentionEviction(t *testing.T) {
	store := newHistoryStoreAt(t)
	text := "evict ghost"
	h := model.PromptContentHash(text)
	trace := []model.ReplayEvent{{AtMS: 0, Kind: model.ReplayEventKey, Rune: "a"}}
	seed := make([]model.SessionResult, model.MaxRetainedHistorySessions)
	for i := range seed {
		seed[i] = model.SessionResult{ID: fmt.Sprintf("fill-%d", i)}
	}
	if err := writeJSONFileAtomic(store.path, model.HistoryFile{Version: historyVersion, Sessions: seed}); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(model.SessionResult{
		ID:          "ghost-old",
		ContentHash: h,
		Prompt:      model.Prompt{Content: text},
		TypingTrace: trace,
		Metrics:     model.SessionMetrics{NetWPM: 50},
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(model.SessionResult{
		ID:          "ghost-best",
		ContentHash: h,
		Prompt:      model.Prompt{Content: text},
		TypingTrace: trace,
		Metrics:     model.SessionMetrics{NetWPM: 90},
	}); err != nil {
		t.Fatal(err)
	}
	got, err := store.BestSessionForGhost(h)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "ghost-best" {
		t.Fatalf("ghost = %q, want ghost-best", got.ID)
	}
}

func TestHistoryWritesCompactJSON(t *testing.T) {
	store := newHistoryStoreAt(t)
	if err := store.Append(model.SessionResult{ID: "1", Prompt: model.Prompt{Content: "hi"}}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(store.path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte("\n  ")) {
		t.Fatalf("expected compact history.json, got indented output")
	}
}
