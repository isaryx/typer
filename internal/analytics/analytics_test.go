package analytics

import (
	"testing"
	"time"

	"typer/internal/model"
)

func TestBuildSummary(t *testing.T) {
	sessions := []model.SessionResult{
		{
			StartedAt: time.Now(),
			Prompt:    model.Prompt{Content: "hello world"},
			TypedText: "hello worxd",
			Metrics: model.SessionMetrics{
				GrossWPM: 50,
				NetWPM:   45,
				Accuracy: 90,
				Errors:   1,
			},
		},
		{
			StartedAt: time.Now(),
			Prompt:    model.Prompt{Content: "go is fun"},
			TypedText: "go os fun",
			Metrics: model.SessionMetrics{
				GrossWPM: 60,
				NetWPM:   55,
				Accuracy: 92,
				Errors:   1,
			},
		},
	}
	s := BuildSummary(sessions, 5)
	if s.Sessions != 2 {
		t.Fatalf("expected 2 sessions, got %d", s.Sessions)
	}
	if s.AvgNetWPM <= 0 || s.AvgAccuracy <= 0 {
		t.Fatalf("expected positive averages")
	}
	if len(s.TopErrorChars) == 0 {
		t.Fatalf("expected non-empty error chars")
	}
}
