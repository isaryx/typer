package analytics

import (
	"testing"
	"time"

	"typer/internal/model"
	"typer/internal/scoring"
)

func TestBuildSummaryDeterministicAveragesAndCharBuckets(t *testing.T) {
	sessions := []model.SessionResult{
		{
			StartedAt: time.Now(),
			Prompt:    model.Prompt{Content: "a b"},
			TypedText: "a c",
			Metrics: model.SessionMetrics{
				GrossWPM: 50,
				NetWPM:   40,
				AdjustedWPM: 35,
				Accuracy: 80,
				Errors:   2,
			},
		},
		{
			StartedAt: time.Now(),
			Prompt:    model.Prompt{Content: "A\tz "},
			TypedText: "a\tyx",
			Metrics: model.SessionMetrics{
				GrossWPM: 70,
				NetWPM:   60,
				AdjustedWPM: 55,
				Accuracy: 90,
				Errors:   1,
			},
		},
		{
			StartedAt: time.Now(),
			Prompt:    model.Prompt{Content: "go\n"},
			TypedText: "go",
			Metrics: model.SessionMetrics{
				GrossWPM: 30,
				NetWPM:   25,
				AdjustedWPM: 20,
				Accuracy: 85,
				Errors:   1,
			},
		},
		{
			StartedAt: time.Now(),
			Prompt:    model.Prompt{Content: "ok"},
			TypedText: "ok!",
			Metrics: model.SessionMetrics{
				GrossWPM: 90,
				NetWPM:   80,
				AdjustedWPM: 75,
				Accuracy: 95,
				Errors:   0,
			},
		},
	}
	s := BuildSummary(sessions, 3)
	if s.Sessions != 4 {
		t.Fatalf("expected 4 sessions, got %d", s.Sessions)
	}
	if s.AvgGrossWPM != 60 {
		t.Fatalf("AvgGrossWPM = %.2f, want 60.00", s.AvgGrossWPM)
	}
	if s.AvgNetWPM != 51.25 {
		t.Fatalf("AvgNetWPM = %.2f, want 51.25", s.AvgNetWPM)
	}
	if s.AvgAdjustedWPM != 46.25 {
		t.Fatalf("AvgAdjustedWPM = %.2f, want 46.25", s.AvgAdjustedWPM)
	}
	if s.AvgAccuracy != 87.5 {
		t.Fatalf("AvgAccuracy = %.2f, want 87.50", s.AvgAccuracy)
	}
	if s.AvgErrors != 1 {
		t.Fatalf("AvgErrors = %.2f, want 1.00", s.AvgErrors)
	}
	wantConsistency := scoring.ConsistencyFromSamples([]float64{40, 60, 25, 80})
	if s.ConsistencyTrend != wantConsistency {
		t.Fatalf("ConsistencyTrend = %.2f, want %.2f", s.ConsistencyTrend, wantConsistency)
	}

	if len(s.TopErrorChars) != 3 {
		t.Fatalf("TopErrorChars length = %d, want 3", len(s.TopErrorChars))
	}
	wantTop := []CharCount{
		{Key: "<space>", Count: 1},
		{Key: "a", Count: 1},
		{Key: "b", Count: 1},
	}
	for i := range wantTop {
		if s.TopErrorChars[i] != wantTop[i] {
			t.Fatalf("TopErrorChars[%d] = %#v, want %#v", i, s.TopErrorChars[i], wantTop[i])
		}
	}
	if len(s.TopMissingChars) != 1 || s.TopMissingChars[0] != (CharCount{Key: "<newline>", Count: 1}) {
		t.Fatalf("unexpected TopMissingChars: %#v", s.TopMissingChars)
	}
	if len(s.TopUnexpectedChar) != 1 || s.TopUnexpectedChar[0] != (CharCount{Key: "!", Count: 1}) {
		t.Fatalf("unexpected TopUnexpectedChar: %#v", s.TopUnexpectedChar)
	}
}

func TestBuildSummaryEmptySessions(t *testing.T) {
	got := BuildSummary(nil, 0)
	if got.Sessions != 0 {
		t.Fatalf("Sessions = %d, want 0", got.Sessions)
	}
	if got.AvgGrossWPM != 0 || got.AvgNetWPM != 0 || got.AvgAdjustedWPM != 0 {
		t.Fatalf("expected zero averages, got %#v", got)
	}
}

func TestNormalizeChar(t *testing.T) {
	tests := []struct {
		in   rune
		want string
	}{
		{' ', "<space>"},
		{'\t', "<tab>"},
		{'\n', "<newline>"},
		{'A', "a"},
		{'z', "z"},
	}
	for _, tt := range tests {
		if got := normalizeChar(tt.in); got != tt.want {
			t.Fatalf("normalizeChar(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTopCountsSortsByCountThenKey(t *testing.T) {
	got := topCounts(map[string]int{
		"b": 2,
		"a": 2,
		"c": 1,
	}, 2)
	want := []CharCount{
		{Key: "a", Count: 2},
		{Key: "b", Count: 2},
	}
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}
