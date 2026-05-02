package session

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"typer/internal/model"
)

const inputHint = "Hint: type the highlighted word, then Space to advance."

// Caret markers are Unicode combining characters (overlay on the glyph; base letter keeps passage color).
const (
	markUserCaret  = "\u0332" // COMBINING UNDERLINE LOW LINE — bar below
	markGhostCaret = "\u0305" // COMBINING OVERLINE — bar above
)

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
}

func newTypingSessionModel(prompt model.Prompt, strict bool, now func() time.Time, indefinite bool, replay *model.SessionResult, showReplayUI bool, fingerHint bool, noInput bool, hideHint bool) *typingSessionModel {
	words := strings.Fields(prompt.Content)

	shadowStrict := false
	var shadowTrace []model.ReplayEvent
	if replay != nil {
		shadowStrict = model.SessionOptionsForReplay(*replay).Strict
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
		shadowTrace:       shadowTrace,
		shadowStrict:      shadowStrict,
		shadowWords:       make([]string, 0, n),
		shadowWordMatches: make([]bool, 0, n),
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
		return m, nil
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.aborted = true
			m.endedAt = m.now().UTC()
			return m, tea.Quit
		}
		if m.isDone() {
			return m, tea.Quit
		}
		switch msg.Type {
		case tea.KeyBackspace:
			m.applyBackspace()
			m.status = ""
			return m, nil
		case tea.KeySpace, tea.KeyEnter:
			cmd := m.commitCurrentWord()
			if m.isDone() {
				m.endedAt = m.now().UTC()
				return m, tea.Quit
			}
			return m, cmd
		default:
			if msg.Type == tea.KeyRunes {
				cmd := m.appendRunes(msg.Runes)
				m.status = ""
				return m, cmd
			}
		}
	}
	return m, nil
}

func (m *typingSessionModel) View() string {
	if len(m.words) == 0 {
		return "No text available for this session.\n"
	}

	tw := m.wrapWidth()
	var b strings.Builder
	b.WriteString(m.styles.meta.Width(tw).Render(m.sessionHeadingLine()))
	b.WriteString("\n")
	if !m.hideHint {
		b.WriteString(m.styles.meta.Width(tw).Render(inputHint))
		b.WriteString("\n")
	}
	if m.showReplayUI && m.replay != nil {
		b.WriteString(m.styles.meta.Width(tw).Render(formatReplayCompactLine(m.replay)))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(m.renderPassageFrame())
	if m.fingerHint {
		sf := m.suggestedFinger()
		if sf != FingerUnknown {
			b.WriteString("\n\n")
			b.WriteString(m.renderFingerHandsFrame(sf))
		}
	}
	if !m.noInput {
		b.WriteString("\n")
		b.WriteString(m.styles.promptPrefix.Render("> "))
		b.WriteString(m.renderInputWord())
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
	return b.String()
}

func runTypingSession(ctx context.Context, input io.Reader, output io.Writer, prompt model.Prompt, strict, indefinite bool, now func() time.Time, replayBaseline *model.SessionResult, showReplayUI bool, fingerHint bool, noInput bool, hideHint bool) (typingSessionResult, error) {
	m := newTypingSessionModel(prompt, strict, now, indefinite, replayBaseline, showReplayUI, fingerHint, noInput, hideHint)
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
