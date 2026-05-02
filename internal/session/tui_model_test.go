package session

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"typer/internal/model"
	"typer/internal/ui"
)

const helloWorldPrompt = "hello world"

func TestCommitCurrentWordStrictMismatchBlocksAdvance(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		true,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
		false,
		false,
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

func TestAppendRunesStrictRejectsWrongActiveChar(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		true,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
		false,
		false,
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
		false,
		false,
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

func TestCommitCurrentWordStrictMatchAdvancesAndClearsPrompt(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		true,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
		false,
		false,
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

func TestAppendRunesTracksKeystrokeAccuracy(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
		false,
		false,
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

func TestCommitCurrentWordTracksUncorrectedErrors(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
		false,
		false,
		false,
	)
	m.appendRunes([]rune("hxllo"))
	m.commitCurrentWord()

	if m.uncorrectedErrors != 1 {
		t.Fatalf("expected 1 uncorrected error, got %d", m.uncorrectedErrors)
	}
}

func TestLineIndexForWord(t *testing.T) {
	first := []int{0, 8, 15}
	if got := lineIndexForWord(first, 0); got != 0 {
		t.Fatalf("word 0: got %d", got)
	}
	if got := lineIndexForWord(first, 7); got != 0 {
		t.Fatalf("word 7: got %d", got)
	}
	if got := lineIndexForWord(first, 8); got != 1 {
		t.Fatalf("word 8: got %d", got)
	}
	if got := lineIndexForWord(first, 20); got != 2 {
		t.Fatalf("word 20: got %d", got)
	}
}

func TestPassageViewportStart(t *testing.T) {
	const h = passageViewportLines
	if got := passageViewportStart(0, 10, h); got != 0 {
		t.Fatalf("start at top: got %d", got)
	}
	if got := passageViewportStart(1, 10, h); got != 0 {
		t.Fatalf("active on line 1: got %d want 0", got)
	}
	if got := passageViewportStart(2, 10, h); got != 1 {
		t.Fatalf("active on line 2 (early scroll): got %d want 1", got)
	}
	if got := passageViewportStart(8, 10, h); got != 7 {
		t.Fatalf("active on line 8: got %d want 7", got)
	}
	if got := passageViewportStart(9, 10, h); got != 7 {
		t.Fatalf("active on last line: got %d want 7", got)
	}
	if got := passageViewportStart(0, 3, h); got != 0 {
		t.Fatalf("short passage: got %d", got)
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
		false,
		false,
		false,
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
		false,
		false,
		false,
	)
	m.width = 200
	if got := m.wrapWidth(); got != ui.MaxContentWidth {
		t.Fatalf("wrapWidth(200) = %d, want %d", got, ui.MaxContentWidth)
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
		false,
		false,
		false,
	)
	got := m.View()
	if !strings.Contains(got, "typer · words · non-strict") {
		t.Fatalf("expected session heading in view, got %q", got)
	}
}

func TestViewNoInputPlacesHintUnderHeadingHidesPrompt(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt, Mode: model.ModeWords},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
		false,
		false,
		true,
	)
	got := m.View()
	if strings.Contains(got, "> ") {
		t.Fatalf("expected no input prompt, got:\n%s", got)
	}
	if !strings.Contains(got, inputHint) {
		t.Fatalf("expected hint in view:\n%s", got)
	}
	idxHead := strings.Index(got, "typer · words · non-strict")
	idxHint := strings.Index(got, inputHint)
	idxPassage := strings.Index(got, "world")
	if idxHead < 0 || idxHint < 0 || idxPassage < 0 {
		t.Fatalf("missing parts: head=%d hint=%d passage=%d in:\n%s", idxHead, idxHint, idxPassage, got)
	}
	if !(idxHead < idxHint && idxHint < idxPassage) {
		t.Fatalf("want heading, then hint, then passage; head=%d hint=%d passage=%d", idxHead, idxHint, idxPassage)
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
		false,
		false,
		false,
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
		false,
		false,
		false,
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
		false,
		false,
		false,
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
		false,
		false,
		false,
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
		false,
		false,
		false,
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
		false,
		false,
		false,
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
				false,
				false,
				false,
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

func TestMergedWordPieceShowsGhostWhenUserAheadOfShadow(t *testing.T) {
	// Regression: i < wordIndex used to always paint completed styling, which hid the
	// ghost on any word the user had committed but the shadow had not finished yet.
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		&model.SessionResult{TypingTrace: []model.ReplayEvent{{AtMS: 0, Kind: model.ReplayEventKey, Rune: "x"}}},
		false,
		false,
		false,
	)
	m.wordIndex = 1
	m.current = "w"
	m.shadowWordIndex = 0
	m.shadowCurrent = "hell"
	s := m.renderMergedWordPiece(0, "hello")
	if !strings.Contains(s, markGhostCaret) {
		t.Fatalf("expected ghost mark on prior word: %q", s)
	}
}

func TestMergedActiveWordShowsGhostWhenUserWordComplete(t *testing.T) {
	// Regression: used to delegate to renderActiveWord when len(typed) == len(target),
	// which dropped the ghost overline until Space (looked like the caret flickered off).
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		&model.SessionResult{TypingTrace: []model.ReplayEvent{{AtMS: 0, Kind: model.ReplayEventKey, Rune: "x"}}},
		false,
		false,
		false,
	)
	m.wordIndex = 0
	m.current = "hello"
	m.shadowWordIndex = 0
	m.shadowCurrent = "he"
	s := m.renderMergedActiveWord("hello")
	if !strings.Contains(s, markGhostCaret) {
		t.Fatalf("expected ghost mark in merged render: %q", s)
	}
}

func TestFormatReplaySessionLabel(t *testing.T) {
	t.Setenv("TZ", "UTC")
	ts := time.Date(2026, 4, 30, 12, 4, 5, 0, time.UTC)
	const ulid = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	if got := formatReplaySessionLabel(&model.SessionResult{ID: ulid, StartedAt: ts}); got != ulid {
		t.Fatalf("prefer id: got %q want %q", got, ulid)
	}
	if got, want := formatReplaySessionLabel(&model.SessionResult{StartedAt: ts}), "2026-04-30 12:04:05"; got != want {
		t.Fatalf("no id, use StartedAt: got %q want %q", got, want)
	}
	if formatReplaySessionLabel(nil) != "" {
		t.Fatal("nil session")
	}
	if formatReplaySessionLabel(&model.SessionResult{}) != "" {
		t.Fatal("no id and no StartedAt should yield empty label")
	}
}

func TestViewReplayTitleShowsSessionID(t *testing.T) {
	t.Setenv("TZ", "UTC")
	const sid = "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	m := newTypingSessionModel(
		model.Prompt{Content: "a b"},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		&model.SessionResult{
			ID:        sid,
			StartedAt: time.Date(2026, 4, 30, 12, 4, 5, 0, time.UTC),
			Metrics:   model.SessionMetrics{NetWPM: 10, Accuracy: 90, Errors: 0},
			ElapsedMS: 1000,
		},
		true,
		false,
		false,
	)
	v := m.View()
	want := "Replay · " + sid + " | net 10.00 wpm"
	if !strings.Contains(v, want) {
		t.Fatalf("expected %q in view, got:\n%s", want, v)
	}
}

func TestViewGhostFromHistoryUsesNormalStartChrome(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: "a b"},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		&model.SessionResult{
			TypingTrace: []model.ReplayEvent{{AtMS: 0, Kind: model.ReplayEventKey, Rune: "x"}},
			StartedAt:   time.Date(2026, 4, 30, 12, 4, 5, 0, time.UTC),
			Metrics:     model.SessionMetrics{NetWPM: 99, Accuracy: 100, Errors: 0},
		},
		false,
		false,
		false,
	)
	v := m.View()
	if strings.Contains(v, "Replay ·") || strings.Contains(v, "Previous run:") {
		t.Fatalf("ghost-only session should not show replay chrome, got:\n%s", v)
	}
	if !strings.Contains(v, inputHint) {
		t.Fatal("expected hint above input")
	}
}

func TestViewFingerHintShowsFramedHands(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: helloWorldPrompt},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
		false,
		true,
		false,
	)
	m.width = 88
	v := m.View()
	if !strings.Contains(v, "Left hand") || !strings.Contains(v, "Right hand") {
		t.Fatalf("expected hands labels in view:\n%s", v)
	}
	if !strings.Contains(v, "Lp") || !strings.Contains(v, "Rp") {
		t.Fatalf("expected finger abbreviations:\n%s", v)
	}
	if strings.Contains(v, "Next ") {
		t.Fatalf("should not show Next line:\n%s", v)
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
