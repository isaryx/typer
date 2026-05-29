package storage

import (
	"testing"
	"time"

	"typer/internal/model"
)

func TestHistoryStoreBestSessionForGhostAfterReload(t *testing.T) {
	store := newHistoryStoreAt(t)
	text := "ghost pick reload"
	h := model.PromptContentHash(text)
	trace := []model.ReplayEvent{{AtMS: 0, Kind: model.ReplayEventKey, Rune: "a"}}
	if err := store.Append(model.SessionResult{
		ID:          "best",
		ContentHash: h,
		Prompt:      model.Prompt{Content: text},
		TypingTrace: trace,
		Metrics:     model.SessionMetrics{NetWPM: 80},
		StartedAt:   time.Unix(1, 0),
	}); err != nil {
		t.Fatal(err)
	}
	reloaded := NewHistoryStoreAt(store.path)
	got, err := reloaded.BestSessionForGhost(h)
	if err != nil {
		t.Fatalf("BestSessionForGhost: %v", err)
	}
	if got.ID != "best" {
		t.Fatalf("got ID %q, want best", got.ID)
	}
	if len(got.TypingTrace) == 0 {
		t.Fatal("expected trace attached from sidecar")
	}
}
