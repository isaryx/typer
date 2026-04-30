package session

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"typer/internal/model"
)

const helloWorldPrompt = "hello world"

func TestCommitCurrentWordStrictMismatchBlocksAdvance(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		true,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
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

func TestAppendRunesStrictRejectsWrongActiveChar(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		true,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
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

func TestCommitCurrentWordNonStrictAdvancesAndClearsPrompt(t *testing.T) {
	now := time.Unix(100, 0)
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		false,
		func() time.Time {
			now = now.Add(time.Second)
			return now
		},
		false,
		nil,
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

func TestCommitCurrentWordStrictMatchAdvancesAndClearsPrompt(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		true,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
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

func TestAppendRunesTracksKeystrokeAccuracy(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
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

func TestCommitCurrentWordTracksUncorrectedErrors(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
	)
	m.appendRunes([]rune("hxllo"))
	m.commitCurrentWord()

	if m.uncorrectedErrors != 1 {
		t.Fatalf("expected 1 uncorrected error, got %d", m.uncorrectedErrors)
	}
}

func TestRenderWordsWrapsToTerminalWidth(t *testing.T) {
	// Many one-letter "words"; each styled segment is wider than a bare rune but still tiny.
	content := strings.TrimSpace(strings.Repeat("x ", 60))
	m := newTypingSessionModel(
		model.Prompt{Content: content},
		true,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
	)
	m.width = 16
	out := m.renderWords(14)
	if !strings.Contains(out, "\n") {
		t.Fatalf("expected soft-wrapped lines, got single line: %q", out)
	}
}

func TestWrapWidthCapsWideTerminal(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: "hello"},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
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

func TestViewIncludesPromptModeInMetaLine(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt, Mode: model.ModeWords},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
	)
	got := m.View()
	if !strings.Contains(got, "Mode: non-strict (words)") {
		t.Fatalf("expected mode label in header, got %q", got)
	}
}

func TestAppendRunesStartsTimerOnFirstKeystroke(t *testing.T) {
	t0 := time.Unix(100, 0)
	now := t0
	m := newTypingSessionModel(
					model.Prompt{Content: helloWorldPrompt},
		false,
		func() time.Time { return now },
		false,
		nil,
	)
	if !m.startedAt.IsZero() {
		t.Fatalf("expected zero startedAt before typing, got %v", m.startedAt)
	}

	now = t0.Add(2 * time.Second)
	m.appendRunes([]rune("h"))
	if got := m.startedAt; !got.Equal(now.UTC()) {
		t.Fatalf("startedAt after first keystroke = %v, want %v", got, now.UTC())
	}

	now = t0.Add(5 * time.Second)
	m.appendRunes([]rune("e"))
	if got := m.startedAt; !got.Equal(t0.Add(2 * time.Second).UTC()) {
		t.Fatalf("startedAt changed after subsequent keystroke = %v", got)
	}
}

func TestResultUsesEndTimeWhenNoTypingOccurred(t *testing.T) {
	now := time.Unix(200, 0)
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		false,
		func() time.Time { return now },
		false,
		nil,
	)
	m.aborted = true
	m.endedAt = now.UTC()

	got := m.result()
	if !got.StartedAt.Equal(now.UTC()) {
		t.Fatalf("StartedAt = %v, want endedAt %v when no typing occurred", got.StartedAt, now.UTC())
	}
	if !got.EndedAt.Equal(now.UTC()) {
		t.Fatalf("EndedAt = %v, want %v", got.EndedAt, now.UTC())
	}
}

func TestInitReturnsNilCommand(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
	)
	if cmd := m.Init(); cmd != nil {
		t.Fatalf("expected nil init command, got %v", cmd)
	}
}

func TestUpdateBackspaceRemovesLastRune(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: "hello"},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
	)
	m.current = "hé"
	gotModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	updated := gotModel.(*typingSessionModel)
	if updated.current != "h" {
		t.Fatalf("current = %q, want %q", updated.current, "h")
	}
}

func TestUpdateSpaceCommitsAndCompletes(t *testing.T) {
	now := time.Unix(100, 0)
	m := newTypingSessionModel(
		model.Prompt{Content: "hello"},
		false,
		func() time.Time {
			now = now.Add(time.Second)
			return now
		},
		false,
		nil,
	)
	m.current = "hello"
	gotModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	updated := gotModel.(*typingSessionModel)
	if !updated.isDone() {
		t.Fatal("expected session to be done after committing only word")
	}
	if updated.endedAt.IsZero() {
		t.Fatal("expected endedAt to be set on completion")
	}
	if cmd == nil {
		t.Fatal("expected quit command when session completes")
	}
}

func TestUpdateEnterCommitsAndCompletes(t *testing.T) {
	now := time.Unix(100, 0)
	m := newTypingSessionModel(
		model.Prompt{Content: "go"},
		false,
		func() time.Time {
			now = now.Add(time.Second)
			return now
		},
		false,
		nil,
	)
	m.current = "go"
	gotModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := gotModel.(*typingSessionModel)
	if !updated.isDone() {
		t.Fatal("expected enter to commit and finish")
	}
	if cmd == nil {
		t.Fatal("expected quit command on completion")
	}
}

func TestUpdateEscAndCtrlCAbort(t *testing.T) {
	for _, key := range []tea.KeyType{tea.KeyEsc, tea.KeyCtrlC} {
		t.Run(key.String(), func(t *testing.T) {
			m := newTypingSessionModel(
					model.Prompt{Content: helloWorldPrompt},
				false,
				func() time.Time { return time.Unix(123, 0) },
				false,
				nil,
			)
			gotModel, cmd := m.Update(tea.KeyMsg{Type: key})
			updated := gotModel.(*typingSessionModel)
			if !updated.aborted {
				t.Fatal("expected aborted flag to be set")
			}
			if updated.endedAt.IsZero() {
				t.Fatal("expected endedAt to be recorded")
			}
			if cmd == nil {
				t.Fatal("expected quit command")
			}
		})
	}
}

func TestRemoveLastRune(t *testing.T) {
	if got := removeLastRune(""); got != "" {
		t.Fatalf("empty input: got %q", got)
	}
	if got := removeLastRune("abc"); got != "ab" {
		t.Fatalf("ascii input: got %q", got)
	}
	if got := removeLastRune("hé"); got != "h" {
		t.Fatalf("unicode input: got %q", got)
	}
}
