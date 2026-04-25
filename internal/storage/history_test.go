package storage

import (
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
	return &HistoryStore{path: filepath.Join(t.TempDir(), historyJSONFile)}
}

func TestNewHistoryStoreUsesConfigDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
		t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
		t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	}
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
	seed := make([]model.SessionResult, maxHistorySessions)
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
	if len(all) != maxHistorySessions {
		t.Fatalf("expected retained count %d, got %d", maxHistorySessions, len(all))
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
	store := &HistoryStore{path: path}
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
