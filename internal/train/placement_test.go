package train

import (
	"testing"

	"typer/internal/model"
)

func TestPlacementStaticSegmentsCoverAlphabet(t *testing.T) {
	t.Parallel()

	c, err := LoadCurriculum()
	if err != nil {
		t.Fatal(err)
	}
	segs := c.PlacementSegments()
	if len(segs) != 3 {
		t.Fatalf("segments = %d, want 3", len(segs))
	}

	combined := ""
	for _, seg := range segs {
		if seg.Content != "" {
			combined += " " + seg.Content
		}
	}
	if !HasAllLetters(combined) {
		t.Fatalf("combined static placement content missing letters: %v", MissingPlacementLetters([]PlacementSegment{{Content: combined}}))
	}
}

func TestPlacementSegmentStrictFlags(t *testing.T) {
	t.Parallel()

	c, err := LoadCurriculum()
	if err != nil {
		t.Fatal(err)
	}
	segs := c.PlacementSegments()
	if len(segs) != 3 {
		t.Fatalf("segments = %d, want 3", len(segs))
	}
	if segs[0].PlacementStrict() {
		t.Fatal("segment A should be non-strict warm-up")
	}
	if !segs[1].PlacementStrict() || !segs[2].PlacementStrict() {
		t.Fatal("segments B and C should be strict")
	}
}

func TestFinalizePlacementProfileDetectsWeakKeys(t *testing.T) {
	t.Parallel()

	opts := model.SessionOptionsSnapshot{Strict: true, Mode: model.ModeTrain}
	p := NewProfile(PlacementResult{AssignedTier: TierFoundation, AssignedLesson: "1.1"})

	for range 12 {
		MergeSession(&p, model.SessionResult{
			Options:   opts,
			Prompt:    model.Prompt{Content: "qqqq zzzz"},
			TypedText: "qqqq zzzz",
			TypingTrace: strictTraceWithWrongBeforeEach("qqqq", 'w'),
		})
	}
	FinalizePlacementProfile(&p)
	if len(p.WeakKeys) == 0 {
		t.Fatal("expected weak keys after placement-style strict sessions")
	}
}

func TestPlacementSegmentAIsFJWarmup(t *testing.T) {
	t.Parallel()

	c, err := LoadCurriculum()
	if err != nil {
		t.Fatal(err)
	}
	segs := c.PlacementSegments()
	if len(segs) == 0 || segs[0].Content == "" {
		t.Fatal("missing segment A content")
	}
	coverage := LetterKeyCoverage(segs[0].Content)
	for _, key := range []string{"f", "j"} {
		if coverage[key] == 0 {
			t.Fatalf("segment A missing home key %q", key)
		}
	}
	for _, key := range []string{"q", "z", "x"} {
		if coverage[key] > 0 {
			t.Fatalf("segment A should not drill %q yet, got %d", key, coverage[key])
		}
	}
}

func TestPlacementSegmentBKeySpread(t *testing.T) {
	t.Parallel()

	c, err := LoadCurriculum()
	if err != nil {
		t.Fatal(err)
	}
	segs := c.PlacementSegments()
	if len(segs) < 2 || segs[1].Content == "" {
		t.Fatal("missing segment B content")
	}
	coverage := LetterKeyCoverage(segs[1].Content)
	if len(coverage) < 20 {
		t.Fatalf("segment B covers %d keys, want at least 20", len(coverage))
	}
	for _, key := range []string{"f", "j", "a", "s", "d", "k", "l", "q", "z", "x"} {
		if coverage[key] == 0 {
			t.Fatalf("segment B missing key %q", key)
		}
	}
}

func TestSegmentAccuracyStrictUsesKeystrokes(t *testing.T) {
	t.Parallel()

	textAcc := SegmentAccuracy(model.SessionResult{
		Options:           model.SessionOptionsSnapshot{Strict: true},
		Metrics:           model.SessionMetrics{Accuracy: 100},
		TotalKeystrokes:   10,
		CorrectKeystrokes: 8,
	})
	if textAcc != 80 {
		t.Fatalf("strict segment accuracy = %v, want 80 keystroke acc", textAcc)
	}
	nonStrict := SegmentAccuracy(model.SessionResult{
		Options:           model.SessionOptionsSnapshot{Strict: false},
		Metrics:           model.SessionMetrics{Accuracy: 95},
		TotalKeystrokes:   10,
		CorrectKeystrokes: 8,
	})
	if nonStrict != 95 {
		t.Fatalf("non-strict segment accuracy = %v, want 95 text acc", nonStrict)
	}
}
