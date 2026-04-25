package storage

import (
	"path/filepath"
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
