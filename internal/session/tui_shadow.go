package session

import (
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"

	"typer/internal/model"
)

type shadowTickMsg struct{}

func (m *typingSessionModel) applyShadowTick() (tea.Model, tea.Cmd) {
	if m.replayClockStart.IsZero() {
		return m, nil
	}
	elapsed := m.now().Sub(m.replayClockStart).Milliseconds()
	for m.shadowTracePos < len(m.shadowTrace) && m.shadowTrace[m.shadowTracePos].AtMS <= elapsed {
		ev := m.shadowTrace[m.shadowTracePos]
		switch ev.Kind {
		case model.ReplayEventKey:
			if ev.Rune != "" {
				r, _ := utf8.DecodeRuneInString(ev.Rune)
				if r != utf8.RuneError {
					m.applyShadowKey(r)
				}
			}
		case model.ReplayEventBackspace:
			m.applyShadowBackspace()
		case model.ReplayEventCommit:
			m.applyShadowCommit()
		}
		m.shadowTracePos++
	}
	return m, m.scheduleShadowTick()
}

func (m *typingSessionModel) scheduleShadowTick() tea.Cmd {
	if m.shadowTracePos >= len(m.shadowTrace) {
		return nil
	}
	elapsed := m.now().Sub(m.replayClockStart).Milliseconds()
	nextAt := m.shadowTrace[m.shadowTracePos].AtMS
	gap := nextAt - elapsed
	var delay time.Duration
	switch {
	case gap <= 0:
		delay = 16 * time.Millisecond
	case gap < 16:
		delay = time.Duration(gap) * time.Millisecond
	default:
		delay = time.Duration(gap) * time.Millisecond
		if delay > 250*time.Millisecond {
			delay = 250 * time.Millisecond
		}
	}
	return tea.Tick(delay, func(time.Time) tea.Msg { return shadowTickMsg{} })
}

func (m *typingSessionModel) applyShadowKey(r rune) {
	if m.shadowWordIndex >= len(m.words) {
		return
	}
	target := m.wordRunes[m.shadowWordIndex]
	current := []rune(m.shadowCurrent)
	pos := len(current)
	matched := pos < len(target) && r == target[pos]
	if m.shadowStrict && !matched {
		return
	}
	m.shadowCurrent = string(append(current, r))
	m.bumpWordStyle(m.shadowWordIndex)
}

func (m *typingSessionModel) applyShadowBackspace() {
	m.shadowCurrent = removeLastRune(m.shadowCurrent)
	m.bumpWordStyle(m.shadowWordIndex)
}

func (m *typingSessionModel) applyShadowCommit() {
	if m.shadowWordIndex >= len(m.words) {
		return
	}
	if m.shadowCurrent == "" {
		return
	}
	targetWord := m.words[m.shadowWordIndex]
	matched := m.shadowCurrent == targetWord
	if m.shadowStrict && !matched {
		return
	}
	m.bumpWordStyle(m.shadowWordIndex)
	m.shadowWordMatches = append(m.shadowWordMatches, matched)
	m.shadowWords = append(m.shadowWords, m.shadowCurrent)
	m.shadowWordIndex++
	m.shadowCurrent = ""
	if m.shadowWordIndex < len(m.wordStyleGen) {
		m.bumpWordStyle(m.shadowWordIndex)
	}
}
