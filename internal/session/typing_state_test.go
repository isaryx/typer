package session

import (
	"testing"
	"time"

	"typer/internal/model"
)

func TestTypingStateApplyRunesStrict(t *testing.T) {
	s := newTypingState([]string{"hello", "world"}, true, func() time.Time { return time.Unix(100, 0) })

	clock, _ := s.applyRunes([]rune("x"))
	if !clock {
		t.Fatal("first keystroke batch starts session clock even when strict rejects every rune")
	}
	if s.current != "" {
		t.Fatalf("want empty current, got %q", s.current)
	}
	if s.totalKeystrokes != 1 {
		t.Fatalf("wrong key still counts toward keystrokes, got %d", s.totalKeystrokes)
	}

	if clock, _ := s.applyRunes([]rune("he")); clock {
		t.Fatal("later batch should not restart clock")
	}
	if s.current != "he" {
		t.Fatalf("want he, got %q", s.current)
	}

	if clock, _ := s.applyRunes([]rune("x")); clock {
		t.Fatal("should not restart clock")
	}
	if s.current != "he" {
		t.Fatalf("wrong rune rejected: got %q", s.current)
	}
}

func TestTypingStateApplyRunesMistakeAfterPrefixBroken(t *testing.T) {
	// "hal" is not a prefix of "hello"; the next "l" matches target[3] but the word is already wrong.
	s := newTypingState([]string{"hello"}, false, func() time.Time { return time.Unix(100, 0) })
	s.applyRunes([]rune("hal"))
	_, mistake := s.applyRunes([]rune("l"))
	if !mistake {
		t.Fatal("want mistake feedback once prefix no longer matches target")
	}
}

func TestTypingStateApplyCommitWordTable(t *testing.T) {
	t.Parallel()
	now := time.Unix(1000, 0)
	clk := func() time.Time { return now }

	t.Run("empty_buffer_shows_empty_flag", func(t *testing.T) {
		s := newTypingState([]string{"go"}, false, clk)
		r := s.applyCommitWord()
		if !r.emptyCurrent || r.advanced || r.strictBlocked || r.sessionClockStarted {
			t.Fatalf("unexpected %+v", r)
		}
	})

	t.Run("strict_mismatch_blocked", func(t *testing.T) {
		s := newTypingState([]string{"hi"}, true, clk)
		s.current = "ho"
		r := s.applyCommitWord()
		if !r.strictBlocked || r.advanced || r.emptyCurrent {
			t.Fatalf("unexpected %+v", r)
		}
		if s.wordIndex != 0 {
			t.Fatal("should not advance")
		}
	})

	t.Run("non_strict_mismatch_advances", func(t *testing.T) {
		s := newTypingState([]string{"hi"}, false, clk)
		s.applyRunes([]rune("ho"))
		r := s.applyCommitWord()
		if !r.advanced || r.strictBlocked || r.emptyCurrent {
			t.Fatalf("unexpected %+v", r)
		}
		if s.wordIndex != 1 || s.totalErrors != 1 {
			t.Fatalf("wordIndex=%d errors=%d", s.wordIndex, s.totalErrors)
		}
	})

	t.Run("commit_without_prior_keystroke_starts_clock", func(t *testing.T) {
		s := newTypingState([]string{"a"}, false, clk)
		s.current = "a"
		r := s.applyCommitWord()
		if !r.advanced || !r.sessionClockStarted {
			t.Fatalf("unexpected %+v", r)
		}
		var sawCommit bool
		for _, ev := range s.typingTrace {
			if ev.Kind == model.ReplayEventCommit {
				sawCommit = true
				break
			}
		}
		if !sawCommit {
			t.Fatal("expected commit event")
		}
	})
}

// TestTypingStateSessionCountersGolden walks a multi-word session and locks counter semantics
// for performance refactors (rune buffer, layout) to preserve.
func TestTypingStateSessionCountersGolden(t *testing.T) {
	start := time.Unix(1000, 0)
	tick := 0
	clk := func() time.Time {
		return start.Add(time.Duration(tick) * time.Second)
	}
	words := []string{"the", "quick", "brown", "fox"}
	s := newTypingState(words, false, clk)

	// Word 0: correct
	tick = 1
	s.applyRunes([]rune("the"))
	tick = 2
	s.applyCommitWord()

	// Word 1: typo then fix — "quik", backspace, "ck"
	tick = 3
	s.applyRunes([]rune("quik"))
	tick = 4
	s.applyBackspace()
	tick = 5
	s.applyRunes([]rune("ck"))
	tick = 6
	s.applyCommitWord()

	// Word 2: correct
	tick = 7
	s.applyRunes([]rune("brown"))
	tick = 8
	s.applyCommitWord()

	// Word 3: partial mismatch committed in non-strict
	tick = 9
	s.applyRunes([]rune("fix"))
	tick = 10
	s.applyCommitWord()

	if s.totalKeystrokes != 17 {
		t.Fatalf("totalKeystrokes = %d, want 17", s.totalKeystrokes)
	}
	if s.correctKeystrokes != 15 {
		t.Fatalf("correctKeystrokes = %d, want 15", s.correctKeystrokes)
	}
	if s.totalErrors != 1 {
		t.Fatalf("totalErrors = %d, want 1", s.totalErrors)
	}
	if s.uncorrectedErrors != 1 {
		t.Fatalf("uncorrectedErrors = %d, want 1", s.uncorrectedErrors)
	}
	if s.typedCharCount != 19 {
		t.Fatalf("typedCharCount = %d, want 19", s.typedCharCount)
	}
	if len(s.wpmSamples) != 4 {
		t.Fatalf("wpmSamples len = %d, want 4", len(s.wpmSamples))
	}
	if s.wordIndex != 4 {
		t.Fatalf("wordIndex = %d, want 4", s.wordIndex)
	}
	if !s.isDone() {
		t.Fatal("expected session done")
	}
}

func TestTypingStateRuneBufferUTF8(t *testing.T) {
	s := newTypingState([]string{"café", "世界"}, false, func() time.Time { return time.Unix(100, 0) })
	s.applyRunes([]rune("café"))
	r := s.applyCommitWord()
	if !r.advanced {
		t.Fatalf("commit café: %+v", r)
	}
	s.applyRunes([]rune("世界"))
	r = s.applyCommitWord()
	if !r.advanced || !s.isDone() {
		t.Fatalf("commit 世界: %+v done=%v", r, s.isDone())
	}
	if s.typedCharCount != 7 { // space + 世界(2) after café(4) with space between words
		t.Fatalf("typedCharCount = %d, want 7", s.typedCharCount)
	}
}

func TestTypingStatePasteAndBackspaceRuneBuffer(t *testing.T) {
	s := newTypingState([]string{"hello"}, false, func() time.Time { return time.Unix(100, 0) })
	s.applyRunes([]rune("hellx"))
	s.applyBackspace()
	s.applyRunes([]rune("o"))
	if s.current != "hello" {
		t.Fatalf("current = %q, want hello", s.current)
	}
}
