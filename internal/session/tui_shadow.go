package session

import (
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"

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
	var cmd tea.Cmd
	if m.shadowTracePos < len(m.shadowTrace) {
		cmd = tea.Tick(16*time.Millisecond, func(time.Time) tea.Msg { return shadowTickMsg{} })
	}
	return m, cmd
}

func (m *typingSessionModel) applyShadowKey(r rune) {
	if m.shadowWordIndex >= len(m.words) {
		return
	}
	target := []rune(m.words[m.shadowWordIndex])
	current := []rune(m.shadowCurrent)
	pos := len(current)
	matched := pos < len(target) && r == target[pos]
	if m.shadowStrict && !matched {
		return
	}
	m.shadowCurrent = string(append(current, r))
}

func (m *typingSessionModel) applyShadowBackspace() {
	m.shadowCurrent = removeLastRune(m.shadowCurrent)
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
	m.shadowWordMatches = append(m.shadowWordMatches, matched)
	m.shadowWords = append(m.shadowWords, m.shadowCurrent)
	m.shadowWordIndex++
	m.shadowCurrent = ""
}
