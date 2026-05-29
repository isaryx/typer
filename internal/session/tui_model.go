package session

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"typer/internal/game/hangman"
	"typer/internal/model"
	"typer/internal/ui"
)

const inputHint = "Hint: type the highlighted word, then Space to advance."

type typingSessionResult struct {
	TypedText         string
	StartedAt         time.Time
	EndedAt           time.Time
	Completed         bool
	Aborted           bool
	WPMSamples        []float64
	Strict            bool
	TotalErrors       int
	TotalKeystrokes   int
	CorrectKeystrokes int
	UncorrectedErrors int
	TypingTrace       []model.ReplayEvent
	HangmanOutcome    string // "", "win", "lose"
	HangmanStage      int
}

type typingSessionModel struct {
	*typingState
	prompt  model.Prompt
	endedAt time.Time
	styles  tuiStyles
	status  string
	aborted bool
	// width is terminal width in cells (from WindowSizeMsg). Used to soft-wrap long prompts.
	width int
	// indefinite shows a hint that Ctrl+C ends the run with a summary (CLI indefinite mode).
	indefinite bool
	// fingerHint enables touch-typing finger hints (US QWERTY diagram).
	fingerHint bool
	// noInput hides the "> …" input line (hint still follows the session heading).
	noInput bool
	// hideHint hides the typing hint line under the session heading (from settings or SessionOptions).
	hideHint bool
	// inputPlacement positions and aligns the "> …" input row (from settings).
	inputPlacement model.InputPlacement
	// replay, when set, holds the prior session for shadow trace and (if showReplayUI) replay chrome.
	replay *model.SessionResult
	// showReplayUI is true only for `typer replay`; ghost-from-start uses the normal header without replay lines.
	showReplayUI bool

	// Shadow replay (previous session animation), driven by replay.TypingTrace.
	shadowTrace       []model.ReplayEvent
	shadowTracePos    int
	shadowStrict      bool
	shadowWords       []string
	shadowWordMatches []bool
	shadowCurrent     string
	shadowWordIndex   int
	replayClockStart  time.Time
	// bellOut receives the terminal bell (ASCII 7) on mistakes; nil disables audible feedback.
	bellOut io.Writer

	// Plain wrap geometry (rebuilt on terminal width change).
	plainLayout passageLayout
	layoutWidth int

	// Per-word styled segment cache (Phase 3).
	wordStyleGen   []uint64
	styledAtGen    []uint64
	styledSegments []string

	// Finger hint frame cache.
	lastFingerHintKey   rune
	cachedFingerHands   string
	hasCachedFingerHands bool

	// Hangman game overlay (nil for normal sessions).
	hangman            *hangman.State
	hangmanOutcome     string
	hangmanPendingQuit bool
}

func newTypingSessionModel(prompt model.Prompt, strict bool, now func() time.Time, indefinite bool, replay *model.SessionResult, showReplayUI bool, fingerHint bool, noInput bool, hideHint bool, inputPlacement model.InputPlacement, bellOut io.Writer, hm *hangman.State) *typingSessionModel {
	words := strings.Fields(prompt.Content)

	shadowStrict := false
	var shadowTrace []model.ReplayEvent
	if replay != nil {
		// Ghost overlay (no replay chrome) should not use strict shadow replay: a prior strict
		// run can include mistyped keys in TypingTrace; strict shadow would ignore them and
		// the animation would appear frozen. `typer replay` keeps strict for a faithful shadow.
		if showReplayUI {
			shadowStrict = model.SessionOptionsForReplay(*replay).Strict
		}
		if len(replay.TypingTrace) > 0 {
			shadowTrace = append([]model.ReplayEvent(nil), replay.TypingTrace...)
		}
	}

	n := len(words)
	return &typingSessionModel{
		typingState:       newTypingState(words, strict, now),
		prompt:            prompt,
		styles:            defaultStyles(),
		indefinite:        indefinite,
		replay:            replay,
		showReplayUI:      showReplayUI,
		fingerHint:        fingerHint,
		noInput:           noInput,
		hideHint:          hideHint,
		inputPlacement:    inputPlacement,
		shadowTrace:       shadowTrace,
		shadowStrict:      shadowStrict,
		shadowWords:       make([]string, 0, n),
		shadowWordMatches: make([]bool, 0, n),
		wordStyleGen:      make([]uint64, n),
		styledAtGen:       make([]uint64, n),
		styledSegments:    make([]string, n),
		bellOut:           bellOut,
		hangman:           hm,
	}
}

func (m *typingSessionModel) Init() tea.Cmd {
	return nil
}

func (m *typingSessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case shadowTickMsg:
		return m.applyShadowTick()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		if m.width < 20 {
			m.width = 20
		}
		m.invalidatePlainLayout()
		m.bumpAllWordStyles()
		return m, nil
	case tea.PasteMsg:
		// v2: bracketed paste is a separate message (not KeyPressMsg with long Text).
		if m.isDone() || m.hangmanPendingQuit {
			return m, tea.Quit
		}
		cmd := m.appendRunes([]rune(msg.Content))
		m.status = ""
		if m.hangmanPendingQuit {
			m.endedAt = m.now().UTC()
			return m, tea.Quit
		}
		return m, cmd
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.aborted = true
			m.endedAt = m.now().UTC()
			return m, tea.Quit
		}
		if m.isDone() || m.hangmanPendingQuit {
			return m, tea.Quit
		}
		switch msg.String() {
		case "backspace":
			m.applyBackspace()
			m.status = ""
			return m, nil
		case "space", "enter":
			cmd := m.commitCurrentWord()
			if m.isDone() {
				m.endedAt = m.now().UTC()
				if m.hangman != nil {
					m.hangmanOutcome = "win"
				}
				return m, tea.Quit
			}
			return m, cmd
		default:
			if len(msg.Text) > 0 {
				cmd := m.appendRunes([]rune(msg.Text))
				m.status = ""
				if m.hangmanPendingQuit {
					m.endedAt = m.now().UTC()
					return m, tea.Quit
				}
				return m, cmd
			}
		}
	}
	return m, nil
}

// showInputChrome is true when the "> …" / border input row is painted (mirrors View's showInput).
func (m *typingSessionModel) showInputChrome() bool {
	return !m.noInput && !m.isDone()
}

func (m *typingSessionModel) View() tea.View {
	if len(m.words) == 0 {
		return tea.NewView("No text available for this session.\n")
	}

	tw := m.wrapWidth()
	var b strings.Builder
	viewLines := 0
	writeViewLine := func(s string) {
		b.WriteString(s)
		b.WriteString("\n")
		viewLines++
	}

	writeViewLine(m.styles.meta.Width(tw).Render(m.sessionHeadingLine()))
	if m.hangman != nil {
		b.WriteString(hangman.RenderBoxWithStats(m.hangman, tw, m.styles.border, m.styles.meta))
		b.WriteString("\n")
	}
	if !m.hideHint {
		writeViewLine(m.styles.meta.Width(tw).Render(inputHint))
	}
	if m.showReplayUI && m.replay != nil {
		writeViewLine(m.styles.meta.Width(tw).Render(formatReplayCompactLine(m.replay)))
	}
	// Hide the input line once the passage is finished (same effect as --no-input for the final paint).
	showInput := m.showInputChrome()
	topInput := showInput && m.inputPlacement.V == model.InputVerticalTop
	bottomInput := showInput && m.inputPlacement.V == model.InputVerticalBottom
	if topInput {
		align := m.inputAlign()
		line := lipgloss.NewStyle().Width(tw).Align(align).Render(m.promptInputStyled())
		writeViewLine(line)
	}

	passStart := viewLines
	pl := m.ensurePlainLayout()
	visibleLines, firstWordIdx := m.styledViewportLines(pl)
	passStr, _, _, _ := m.renderPassageFrame(passStart, pl, visibleLines, firstWordIdx)
	b.WriteString(passStr)

	if m.fingerHint {
		sf := m.suggestedFinger()
		if sf != FingerUnknown {
			b.WriteString("\n\n")
			b.WriteString(m.renderFingerHandsFrameCached(sf))
		}
	}
	if bottomInput {
		b.WriteString("\n")
		align := m.inputAlign()
		line := lipgloss.NewStyle().Width(tw).Align(align).Render(m.promptInputStyled())
		b.WriteString(line)
	}
	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(m.styles.errorMessage.Render(m.status))
	}
	if m.indefinite {
		b.WriteString("\n")
		b.WriteString(m.styles.meta.Width(tw).Render("Indefinite mode — Ctrl+C or Esc to stop"))
	}
	b.WriteString("\n")

	// User typing caret is a blinking block (█) in the view string. The only hardware cursor is
	// the ghost bar on the passage when shadow replay is active and the ghost is in view.
	var cur *tea.Cursor
	if m.hasShadowReplay() {
		if gx, gy, gOk := m.ghostCaretViewCoords(passStart, pl, firstWordIdx); gOk {
			cur = m.ghostBarCursor(gx, gy)
		}
	}

	v := tea.NewView(b.String())
	v.Cursor = cur
	return v
}

func (m *typingSessionModel) inputAlign() lipgloss.Position {
	switch m.inputPlacement.H {
	case model.InputHorizontalCenter:
		return lipgloss.Center
	case model.InputHorizontalRight:
		return lipgloss.Right
	default:
		return lipgloss.Left
	}
}

// ghostBarCursor is the passage ghost caret: bar shape, steady, always white (not input good/bad tint).
func (m *typingSessionModel) ghostBarCursor(x, y int) *tea.Cursor {
	c := tea.NewCursor(x, y)
	c.Shape = tea.CursorBar
	c.Blink = false
	c.Color = lipgloss.Color(ui.ColorBorderHex)
	return c
}

func (m *typingSessionModel) inputCursorColor() color.Color {
	if m.wordIndex >= len(m.words) {
		return lipgloss.Color(ui.ColorInputFg)
	}
	target := []rune(m.words[m.wordIndex])
	typed := []rune(m.current)
	for i := range typed {
		if i < len(target) && typed[i] != target[i] {
			return lipgloss.Color(ui.ColorInputBadFg)
		}
		if i >= len(target) {
			return lipgloss.Color(ui.ColorInputBadFg)
		}
	}
	return lipgloss.Color(ui.ColorInputFg)
}

// typingSessionRunOpts groups TUI flags passed from model.SessionOptions into runTypingSession.
type typingSessionRunOpts struct {
	strict         bool
	indefinite     bool
	now            func() time.Time
	replayBaseline *model.SessionResult
	showReplayUI   bool
	fingerHint     bool
	noInput        bool
	hideHint       bool
	inputPlacement model.InputPlacement
	noAudible      bool
	hangman        *hangman.State
}

func runTypingSession(ctx context.Context, input io.Reader, output io.Writer, prompt model.Prompt, o typingSessionRunOpts) (typingSessionResult, error) {
	var bellOut io.Writer
	if !o.noAudible {
		bellOut = output
	}
	m := newTypingSessionModel(prompt, o.strict, o.now, o.indefinite, o.replayBaseline, o.showReplayUI, o.fingerHint, o.noInput, o.hideHint, o.inputPlacement, bellOut, o.hangman)
	if len(m.words) == 0 {
		return typingSessionResult{}, fmt.Errorf("prompt contains no words")
	}

	// Full clear + cursor home so each session (including indefinite rounds) starts uncluttered.
	fmt.Fprint(output, "\x1b[2J\x1b[H")

	p := tea.NewProgram(
		m,
		tea.WithContext(ctx),
		tea.WithInput(input),
		tea.WithOutput(output),
	)
	finalModel, err := p.Run()
	if err != nil {
		return typingSessionResult{}, err
	}

	fm, ok := finalModel.(*typingSessionModel)
	if !ok {
		return typingSessionResult{}, fmt.Errorf("unexpected final session model")
	}
	return fm.result(), nil
}
