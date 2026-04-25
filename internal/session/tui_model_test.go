package session

import (
	"strings"
	"testing"
	"time"

	"typer/internal/model"
)

func TestCommitCurrentWord_StrictMismatchBlocksAdvance(t *testing.T) {
	m := newWordSessionModel(
		model.Prompt{Content: "hello world"},
		true,
		func() time.Time { return time.Unix(100, 0) },
		false,
	)
	m.current = "hullo"
	m.commitCurrentWord()

	if m.wordIndex != 0 {
		t.Fatalf("expected wordIndex to remain 0, got %d", m.wordIndex)
	}
	if m.current != "hullo" {
		t.Fatalf("expected current input to be preserved")
	}
}

func TestAppendRunes_StrictRejectsWrongActiveChar(t *testing.T) {
	m := newWordSessionModel(
		model.Prompt{Content: "hello world"},
		true,
		func() time.Time { return time.Unix(100, 0) },
		false,
	)
	m.appendRunes([]rune("x"))
	if m.current != "" {
		t.Fatalf("expected wrong strict rune to be rejected, got %q", m.current)
	}

	m.appendRunes([]rune("he"))
	if m.current != "he" {
		t.Fatalf("expected matching runes to be accepted, got %q", m.current)
	}

	m.appendRunes([]rune("x"))
	if m.current != "he" {
		t.Fatalf("expected wrong next rune to be rejected, got %q", m.current)
	}
}

func TestCommitCurrentWord_NonStrictAdvancesAndClearsPrompt(t *testing.T) {
	now := time.Unix(100, 0)
	m := newWordSessionModel(
		model.Prompt{Content: "hello world"},
		false,
		func() time.Time {
			now = now.Add(time.Second)
			return now
		},
		false,
	)
	m.current = "hullo"
	m.commitCurrentWord()

	if m.wordIndex != 1 {
		t.Fatalf("expected wordIndex to advance to 1, got %d", m.wordIndex)
	}
	if m.current != "" {
		t.Fatalf("expected current input to clear after advance")
	}
	if m.totalErrors != 1 {
		t.Fatalf("expected mismatch error to be tracked")
	}
}

func TestCommitCurrentWord_StrictMatchAdvancesAndClearsPrompt(t *testing.T) {
	m := newWordSessionModel(
		model.Prompt{Content: "hello world"},
		true,
		func() time.Time { return time.Unix(100, 0) },
		false,
	)
	m.current = "hello"
	m.commitCurrentWord()

	if m.wordIndex != 1 {
		t.Fatalf("expected wordIndex to advance to 1, got %d", m.wordIndex)
	}
	if m.current != "" {
		t.Fatalf("expected current input to clear after valid advance")
	}
}

func TestAppendRunes_TracksKeystrokeAccuracy(t *testing.T) {
	m := newWordSessionModel(
		model.Prompt{Content: "hello world"},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
	)
	m.appendRunes([]rune("hx"))
	if m.totalKeystrokes != 2 {
		t.Fatalf("expected 2 total keystrokes, got %d", m.totalKeystrokes)
	}
	if m.correctKeystrokes != 1 {
		t.Fatalf("expected 1 correct keystroke, got %d", m.correctKeystrokes)
	}
	if m.current != "hx" {
		t.Fatalf("non-strict mode should keep wrong char, got %q", m.current)
	}
}

func TestCommitCurrentWord_TracksUncorrectedErrors(t *testing.T) {
	m := newWordSessionModel(
		model.Prompt{Content: "hello world"},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
	)
	m.appendRunes([]rune("hxllo"))
	m.commitCurrentWord()

	if m.uncorrectedErrors != 1 {
		t.Fatalf("expected 1 uncorrected error, got %d", m.uncorrectedErrors)
	}
}

func TestRenderWords_WrapsToTerminalWidth(t *testing.T) {
	// Many one-letter "words"; each styled segment is wider than a bare rune but still tiny.
	content := strings.TrimSpace(strings.Repeat("x ", 60))
	m := newWordSessionModel(
		model.Prompt{Content: content},
		true,
		func() time.Time { return time.Unix(100, 0) },
		false,
	)
	m.width = 16
	out := m.renderWords(14)
	if !strings.Contains(out, "\n") {
		t.Fatalf("expected soft-wrapped lines, got single line: %q", out)
	}
}

func TestWrapWidth_capsWideTerminal(t *testing.T) {
	m := newWordSessionModel(
		model.Prompt{Content: "hello"},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
	)
	m.width = 200
	if got := m.wrapWidth(); got != maxWrapWidth {
		t.Fatalf("wrapWidth(200) = %d, want %d", got, maxWrapWidth)
	}
	m.width = 40
	if got := m.wrapWidth(); got != 40 {
		t.Fatalf("wrapWidth(40) = %d, want 40", got)
	}
}
