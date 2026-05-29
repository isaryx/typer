package storage

import (
	"fmt"
	"testing"
	"time"

	"typer/internal/model"
)

func BenchmarkHistoryBestSessionForGhost(b *testing.B) {
	dir := b.TempDir()
	path := dir + "/history.json"
	text := "benchmark ghost text"
	h := model.PromptContentHash(text)
	trace := []model.ReplayEvent{{AtMS: 0, Kind: model.ReplayEventKey, Rune: "a"}}
	sessions := make([]model.SessionResult, 5000)
	for i := range sessions {
		sessions[i] = model.SessionResult{
			ID:          "s",
			ContentHash: h,
			Prompt:      model.Prompt{Content: text},
			TypingTrace: trace,
			Metrics:     model.SessionMetrics{NetWPM: float64(i % 100)},
			StartedAt:   time.Unix(int64(i), 0),
		}
		sessions[i].ID = fmt.Sprintf("s%d", i)
	}
	if err := writeJSONFileAtomic(path, model.HistoryFile{Version: historyVersion, Sessions: sessions}); err != nil {
		b.Fatal(err)
	}
	benchStore := NewHistoryStoreAt(path)
	if err := benchStore.ensureLoaded(); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := benchStore.BestSessionForGhost(h); err != nil {
			b.Fatal(err)
		}
	}
}
