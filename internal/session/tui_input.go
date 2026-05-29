package session

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"typer/internal/model"
)

func (m *typingSessionModel) commitCurrentWord() tea.Cmd {
	res := m.typingState.applyCommitWord()
	if res.emptyCurrent {
		m.status = "Type the current word before advancing."
		return nil
	}
	if res.strictBlocked {
		if m.bellOut != nil {
			fmt.Fprint(m.bellOut, "\a")
		}
		m.status = ""
		return nil
	}
	if !res.advanced {
		return nil
	}
	committed := m.wordIndex - 1
	m.bumpWordStyle(committed)
	if m.wordIndex < len(m.wordStyleGen) {
		m.bumpWordStyle(m.wordIndex)
	}
	if m.wordIndex > 0 && m.wordMatches[m.wordIndex-1] {
		m.status = ""
	}
	if res.sessionClockStarted {
		return m.afterSessionClockStart()
	}
	return nil
}

func (m *typingSessionModel) appendRunes(runes []rune) tea.Cmd {
	clock, mistake := m.typingState.applyRunes(runes)
	m.bumpWordStyle(m.wordIndex)
	if mistake && m.bellOut != nil {
		fmt.Fprint(m.bellOut, "\a")
	}
	if clock {
		return m.afterSessionClockStart()
	}
	return nil
}

func (m *typingSessionModel) applyBackspace() {
	m.typingState.applyBackspace()
	m.bumpWordStyle(m.wordIndex)
}

func (m *typingSessionModel) afterSessionClockStart() tea.Cmd {
	if len(m.shadowTrace) > 0 && m.replayClockStart.IsZero() {
		m.replayClockStart = m.now()
		return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg { return shadowTickMsg{} })
	}
	return nil
}

func (m *typingSessionModel) result() typingSessionResult {
	ended := m.endedAt
	if ended.IsZero() {
		ended = m.now().UTC()
	}
	started := m.startedAt
	if started.IsZero() {
		started = ended
	}
	return typingSessionResult{
		TypedText:         strings.Join(m.typedWords, " "),
		StartedAt:         started,
		EndedAt:           ended,
		Completed:         m.isDone() && !m.aborted,
		Aborted:           m.aborted,
		WPMSamples:        append([]float64(nil), m.wpmSamples...),
		Strict:            m.strict,
		TotalErrors:       m.totalErrors,
		TotalKeystrokes:   m.totalKeystrokes,
		CorrectKeystrokes: m.correctKeystrokes,
		UncorrectedErrors: m.uncorrectedErrors,
		TypingTrace:       append([]model.ReplayEvent(nil), m.typingTrace...),
	}
}
