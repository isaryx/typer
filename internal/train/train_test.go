package train_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"typer/internal/model"
	"typer/internal/train"
)

func TestAssignPlacement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		netWPM float64
		acc    float64
		tier   string
		lesson string
	}{
		{10, 80, train.TierFoundation, "1.1"},
		{30, 90, train.TierBuilding, "2.1"},
		{28, 90, train.TierFoundation, "1.5"},
		{25, 88, train.TierFoundation, "1.5"},
		{40, 93, train.TierBuilding, "2.4"},
		{55, 96, train.TierFluent, "3.1"},
	}
	for _, tc := range cases {
		got := train.AssignPlacement(train.PlacementInput{
			Segments: []model.SessionResult{{
				Metrics: model.SessionMetrics{NetWPM: tc.netWPM, Accuracy: tc.acc},
			}},
		})
		if got.AssignedTier != tc.tier || got.AssignedLesson != tc.lesson {
			t.Errorf("net=%.0f acc=%.0f: got tier=%s lesson=%s, want %s %s",
				tc.netWPM, tc.acc, got.AssignedTier, got.AssignedLesson, tc.tier, tc.lesson)
		}
	}
}

func TestProfileStoreSaveLoadReset(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "profile.json")
	store := train.NewProfileStoreAt(path)
	if _, err := store.Load(); err != train.ErrNoProfile {
		t.Fatalf("Load() = %v, want ErrNoProfile", err)
	}

	p := train.NewProfile(train.PlacementResult{
		NetWPM:         42,
		Accuracy:       94,
		AssignedTier:   train.TierBuilding,
		AssignedLesson: "2.1",
	})
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Placement.AssignedLesson != "2.1" {
		t.Fatalf("lesson = %q", got.Placement.AssignedLesson)
	}
	if err := store.Reset(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(); err != train.ErrNoProfile {
		t.Fatalf("after reset: %v", err)
	}
}

func TestMergeSessionWeakKeys(t *testing.T) {
	t.Parallel()

	p := train.NewProfile(train.PlacementResult{AssignedTier: train.TierFoundation, AssignedLesson: "1.1"})
	for i := 0; i < 25; i++ {
		train.MergeSession(&p, model.SessionResult{
			Prompt:    model.Prompt{Content: "pizza pizza pizza"},
			TypedText: "pizaa pizaa pizaa",
			TypingTrace: []model.ReplayEvent{
				{AtMS: 0, Kind: model.ReplayEventKey, Rune: "p"},
				{AtMS: 100, Kind: model.ReplayEventKey, Rune: "i"},
			},
		})
	}
	if len(p.WeakKeys) == 0 {
		t.Fatal("expected weak keys after repeated errors")
	}
}

func TestLoadCurriculum(t *testing.T) {
	t.Parallel()

	c, err := train.LoadCurriculum()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.AllLessons()) < 30 {
		t.Fatalf("lessons = %d, want at least 30", len(c.AllLessons()))
	}
	if _, ok := c.Lesson("1.1"); !ok {
		t.Fatal("missing lesson 1.1")
	}
}

func TestEvaluateLessonPassFail(t *testing.T) {
	t.Parallel()

	c, err := train.LoadCurriculum()
	if err != nil {
		t.Fatal(err)
	}
	lesson, ok := c.Lesson("1.1")
	if !ok {
		t.Fatal("no lesson 1.1")
	}

	p := train.NewProfile(train.PlacementResult{
		AssignedTier:   train.TierFoundation,
		AssignedLesson: "1.1",
	})
	pass := train.EvaluateLesson(&p, c, lesson, model.SessionResult{
		Metrics:           model.SessionMetrics{NetWPM: 20, Accuracy: 95},
		TotalKeystrokes:   100,
		CorrectKeystrokes: 95,
	})
	if !pass.Passed {
		t.Fatalf("expected pass: %s", pass.Message)
	}
	if p.Progress.CurrentLesson != "1.2" {
		t.Fatalf("current lesson = %q, want 1.2", p.Progress.CurrentLesson)
	}

	p2 := train.NewProfile(train.PlacementResult{
		AssignedTier:   train.TierFoundation,
		AssignedLesson: "1.1",
	})
	fail := train.EvaluateLesson(&p2, c, lesson, model.SessionResult{
		Metrics:           model.SessionMetrics{NetWPM: 5, Accuracy: 70},
		TotalKeystrokes:   100,
		CorrectKeystrokes: 70,
	})
	if fail.Passed {
		t.Fatal("expected fail")
	}
}

func TestEvaluateLessonStrictUsesKeystrokeAccuracy(t *testing.T) {
	t.Parallel()

	c, err := train.LoadCurriculum()
	if err != nil {
		t.Fatal(err)
	}
	lesson, ok := c.Lesson("1.1")
	if !ok {
		t.Fatal("no lesson 1.1")
	}
	if !lesson.Strict {
		t.Fatal("lesson 1.1 should be strict")
	}

	p := train.NewProfile(train.PlacementResult{
		AssignedTier:   train.TierFoundation,
		AssignedLesson: "1.1",
	})
	// Perfect final text but too many wrong keys along the way.
	out := train.EvaluateLesson(&p, c, lesson, model.SessionResult{
		Metrics:           model.SessionMetrics{NetWPM: 20, Accuracy: 100},
		TotalKeystrokes:   200,
		CorrectKeystrokes: 170, // 85% keystroke acc
	})
	if out.Passed {
		t.Fatalf("expected fail on low keystroke accuracy: %s", out.Message)
	}
	if !strings.Contains(out.Message, "keystroke acc") {
		t.Fatalf("message should mention keystroke acc: %q", out.Message)
	}

	p2 := train.NewProfile(train.PlacementResult{
		AssignedTier:   train.TierFoundation,
		AssignedLesson: "1.1",
	})
	pass := train.EvaluateLesson(&p2, c, lesson, model.SessionResult{
		Metrics:           model.SessionMetrics{NetWPM: 20, Accuracy: 100},
		TotalKeystrokes:   200,
		CorrectKeystrokes: 190, // 95% keystroke acc
	})
	if !pass.Passed {
		t.Fatalf("expected pass: %s", pass.Message)
	}
}

func TestKeystrokeAccuracy(t *testing.T) {
	t.Parallel()

	if got := train.KeystrokeAccuracy(model.SessionResult{}); got >= 0 {
		t.Fatalf("empty result = %v, want unavailable", got)
	}
	got := train.KeystrokeAccuracy(model.SessionResult{
		TotalKeystrokes:   10,
		CorrectKeystrokes: 9,
	})
	if got != 90 {
		t.Fatalf("got %v, want 90", got)
	}
}

func TestEffectiveRounds(t *testing.T) {
	t.Parallel()

	foundation := train.Lesson{Tier: train.TierFoundation}
	if foundation.EffectiveRounds() != 4 {
		t.Fatalf("foundation rounds = %d", foundation.EffectiveRounds())
	}
	timed := train.Lesson{Tier: train.TierFoundation, TimedMS: 60000}
	if timed.EffectiveRounds() != 1 {
		t.Fatalf("timed rounds = %d", timed.EffectiveRounds())
	}
	custom := train.Lesson{Tier: train.TierFoundation, Rounds: 5}
	if custom.EffectiveRounds() != 5 {
		t.Fatalf("custom rounds = %d", custom.EffectiveRounds())
	}
}

func TestBuildLessonContentSeedLines(t *testing.T) {
	t.Parallel()

	lesson := train.Lesson{
		ID:      "test",
		Tier:    train.TierFoundation,
		Prompts: []string{"ff jj", "dd kk", "ss ll", "aa ;;"},
		Rounds:  4,
	}
	content, err := train.BuildLessonContent(lesson, nil)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(content, "\n")
	if len(lines) != 4 {
		t.Fatalf("lines = %d, want 4", len(lines))
	}
	want := []string{"ff jj", "dd kk", "ss ll", "aa ;;"}
	for i, line := range lines {
		if line != want[i] {
			t.Fatalf("line %d = %q, want %q", i+1, line, want[i])
		}
	}
}

func TestUpdateStreak(t *testing.T) {
	t.Parallel()

	p := train.NewProfile(train.PlacementResult{AssignedTier: train.TierFoundation, AssignedLesson: "1.1"})
	today := time.Now().UTC().Format("2006-01-02")
	train.UpdateStreak(&p)
	if p.Progress.StreakDays != 1 || p.Progress.LastPracticeDate != today {
		t.Fatalf("streak = %d date = %q", p.Progress.StreakDays, p.Progress.LastPracticeDate)
	}
	train.UpdateStreak(&p)
	if p.Progress.StreakDays != 1 {
		t.Fatalf("same-day streak should stay 1, got %d", p.Progress.StreakDays)
	}
}
