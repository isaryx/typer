package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"typer/internal/model"
)

func newHistoryStoreAt(t *testing.T) *HistoryStore {
	t.Helper()
	return &HistoryStore{path: filepath.Join(t.TempDir(), "history.json")}
}

func TestHistoryStore_EmptyList(t *testing.T) {
	store := newHistoryStoreAt(t)

	got, err := store.List(10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty list, got %d", len(got))
	}
}

func TestHistoryStore_AppendAndListReverseOrder(t *testing.T) {
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
		t.Fatalf("List: %v", err)
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

func TestHistoryStore_AppendEnforcesMaxSessions(t *testing.T) {
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
			t.Fatalf("Append: %v", err)
		}
	}

	all, err := store.List(0)
	if err != nil {
		t.Fatalf("List: %v", err)
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

func TestHistoryStore_WritesTightPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX-style file modes are not meaningful on Windows")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")
	store := &HistoryStore{path: path}
	if err := store.Append(model.SessionResult{ID: "1"}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if mode := info.Mode().Perm(); mode != configFilePerm {
		t.Fatalf("history.json perm = %o, want %o", mode, configFilePerm)
	}
}

func TestHistoryStore_ListTruncatesToLast(t *testing.T) {
	store := newHistoryStoreAt(t)
	for i := 0; i < 5; i++ {
		if err := store.Append(model.SessionResult{ID: string(rune('A' + i))}); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
	got, err := store.List(2)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(got))
	}
	if got[0].ID != "E" || got[1].ID != "D" {
		t.Fatalf("expected [E,D] newest first, got [%s,%s]", got[0].ID, got[1].ID)
	}
}

func TestHistoryStore_ResetClearsSessions(t *testing.T) {
	store := newHistoryStoreAt(t)
	if err := store.Append(model.SessionResult{ID: "1"}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := store.Append(model.SessionResult{ID: "2"}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	if err := store.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	got, err := store.List(0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty list after reset, got %d", len(got))
	}
}
