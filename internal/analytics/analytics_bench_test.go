package analytics

import (
	"testing"
	"time"

	"typer/internal/model"
)

func BenchmarkBuildSummary(b *testing.B) {
	sessions := make([]model.SessionResult, 500)
	for i := range sessions {
		sessions[i] = model.SessionResult{
			StartedAt: time.Unix(int64(i), 0),
			Prompt:    model.Prompt{Content: "the quick brown fox"},
			TypedText: "the quik brown fox",
			Metrics: model.SessionMetrics{
				GrossWPM:    60,
				NetWPM:      55,
				AdjustedWPM: 50,
				Accuracy:    95,
				Errors:      1,
			},
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildSummary(sessions, 5)
	}
}
