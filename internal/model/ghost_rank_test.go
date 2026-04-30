package model

import (
	"testing"
	"time"
)

func TestBetterGhostCandidate_traceAndAbort(t *testing.T) {
	withTrace := SessionResult{TypingTrace: []ReplayEvent{{Kind: ReplayEventKey, Rune: "a"}}}
	noTrace := SessionResult{}
	if !BetterGhostCandidate(withTrace, noTrace) {
		t.Fatal("trace should win")
	}
	ok := SessionResult{Aborted: false}
	ab := SessionResult{Aborted: true}
	if !BetterGhostCandidate(ok, ab) {
		t.Fatal("non-aborted should win")
	}
}

func TestBetterGhostCandidate_metrics(t *testing.T) {
	base := func(net float64, acc float64, err int, ms int64) SessionResult {
		return SessionResult{
			TypingTrace: []ReplayEvent{{Kind: ReplayEventKey, Rune: "x"}},
			Metrics: SessionMetrics{
				NetWPM:   net,
				Accuracy: acc,
				Errors:   err,
			},
			ElapsedMS: ms,
			StartedAt: time.Unix(100, 0),
		}
	}
	a := base(50, 99, 1, 5000)
	b := base(40, 99, 1, 5000)
	if !BetterGhostCandidate(a, b) {
		t.Fatal("higher net WPM should win")
	}
	c := base(50, 98, 1, 5000)
	if !BetterGhostCandidate(a, c) {
		t.Fatal("higher accuracy should win")
	}
	d := base(50, 99, 2, 5000)
	if !BetterGhostCandidate(a, d) {
		t.Fatal("fewer errors should win")
	}
	e := base(50, 99, 1, 6000)
	if !BetterGhostCandidate(a, e) {
		t.Fatal("faster (lower ms) should win")
	}
}

func TestBetterGhostCandidate_newerTieBreak(t *testing.T) {
	a := SessionResult{
		TypingTrace: []ReplayEvent{{Kind: ReplayEventKey, Rune: "x"}},
		Metrics:     SessionMetrics{NetWPM: 50, Accuracy: 99, Errors: 0},
		ElapsedMS:   5000,
		StartedAt:   time.Unix(200, 0),
	}
	b := SessionResult{
		TypingTrace: []ReplayEvent{{Kind: ReplayEventKey, Rune: "x"}},
		Metrics:     SessionMetrics{NetWPM: 50, Accuracy: 99, Errors: 0},
		ElapsedMS:   5000,
		StartedAt:   time.Unix(100, 0),
	}
	if !BetterGhostCandidate(a, b) {
		t.Fatal("newer StartedAt should win tie-break")
	}
}
