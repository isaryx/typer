package session

import (
	"fmt"
	"io"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"typer/internal/model"
)

// maxWrapWidth caps soft-wrap and lipgloss layout on wide terminals so lines stay
// readable (roughly 65–95 characters is a common prose band; 88 matches e.g. rustfmt).
const maxWrapWidth = 88

type wordSessionResult struct {
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
}

type wordSessionModel struct {
	prompt            model.Prompt
	words             []string
	typedWords        []string
	wordMatches       []bool
	wordIndex         int
	current           string
	strict            bool
	now               func() time.Time
	startedAt         time.Time
	endedAt           time.Time
	styles            tuiStyles
	status            string
	aborted           bool
	wpmSamples        []float64
	totalErrors       int
	totalKeystrokes   int
	correctKeystrokes int
	uncorrectedErrors int
	// width is terminal width in cells (from WindowSizeMsg). Used to soft-wrap long prompts.
	width int
	// indefinite shows a hint that Ctrl+C ends the run with a summary (CLI indefinite mode).
	indefinite bool
}

func newWordSessionModel(prompt model.Prompt, strict bool, now func() time.Time, indefinite bool) wordSessionModel {
	if now == nil {
		now = time.Now
	}
	words := strings.Fields(prompt.Content)
	return wordSessionModel{
		prompt:     prompt,
		words:      words,
		strict:     strict,
		now:        now,
		startedAt:  now().UTC(),
		styles:     defaultStyles(),
		indefinite: indefinite,
	}
}

func (m wordSessionModel) Init() tea.Cmd {
	return nil
}

func (m wordSessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		if m.width < 20 {
			m.width = 20
		}
		return m, nil
	case tea.KeyMsg:
		if m.isDone() {
			return m, tea.Quit
		}
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.aborted = true
			m.endedAt = m.now().UTC()
			return m, tea.Quit
		case tea.KeyBackspace:
			m.current = removeLastRune(m.current)
			m.status = ""
			return m, nil
		case tea.KeySpace, tea.KeyEnter:
			m.commitCurrentWord()
			if m.isDone() {
				m.endedAt = m.now().UTC()
				return m, tea.Quit
			}
			return m, nil
		default:
			if msg.Type == tea.KeyRunes {
				m.appendRunes(msg.Runes)
				m.status = ""
			}
		}
	}
	return m, nil
}

func (m wordSessionModel) View() string {
	if len(m.words) == 0 {
		return "No text available for this session.\n"
	}

	tw := m.wrapWidth()
	var b strings.Builder
	b.WriteString(m.styles.title.Width(tw).Render("Guide: Type the current word, then press Space to advance"))
	b.WriteString("\n")
	strictMode := "strict"
	if !m.strict {
		strictMode = "non-strict"
	}
	b.WriteString(m.styles.meta.Width(tw).Render(fmt.Sprintf("Mode: %s  |  Progress: %d/%d", strictMode, m.wordIndex, len(m.words))))
	b.WriteString("\n")
	if m.prompt.Author != "" {
		b.WriteString("\n")
		b.WriteString(m.styles.meta.Width(tw).Render(fmt.Sprintf("Author: %s", m.prompt.Author)))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	prompt := m.renderWords(m.promptInnerWidth())
	b.WriteString(m.styles.promptBox.Width(tw).Render(prompt))
	b.WriteString("\n\n")
	b.WriteString("> ")
	b.WriteString(m.renderInputWord())
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

func (m wordSessionModel) termWidth() int {
	if m.width > 0 {
		return m.width
	}
	return 80
}

// wrapWidth is min(terminal width, maxWrapWidth) so content does not stretch across ultra-wide screens.
func (m wordSessionModel) wrapWidth() int {
	tw := m.termWidth()
	if tw > maxWrapWidth {
		return maxWrapWidth
	}
	return tw
}

// promptInnerWidth is the line width for text inside the bordered prompt (account for left+right border runes).
func (m wordSessionModel) promptInnerWidth() int {
	w := m.wrapWidth() - 2
	if w < 1 {
		return 1
	}
	return w
}

func (m wordSessionModel) renderWordPiece(i int, w string) string {
	switch {
	case i < m.wordIndex:
		if i < len(m.wordMatches) && !m.wordMatches[i] {
			return m.styles.completedBad.Render(w)
		}
		return m.styles.completed.Render(w)
	case i == m.wordIndex:
		return m.renderActiveWord(w)
	default:
		return m.styles.upcoming.Render(w)
	}
}

func (m wordSessionModel) renderWords(lineWidth int) string {
	if lineWidth < 1 {
		lineWidth = 1
	}
	var lines []string
	var parts []string
	cur := 0

	for i, w := range m.words {
		seg := m.renderWordPiece(i, w)
		sw := lipgloss.Width(seg)
		sep := 0
		if len(parts) > 0 {
			sep = 1
		}
		if cur+sep+sw > lineWidth && len(parts) > 0 {
			lines = append(lines, strings.Join(parts, " "))
			parts = nil
			cur = 0
			sep = 0
		}
		if len(parts) > 0 {
			cur++
		}
		parts = append(parts, seg)
		cur += sw
	}
	if len(parts) > 0 {
		lines = append(lines, strings.Join(parts, " "))
	}
	return strings.Join(lines, "\n")
}

func (m wordSessionModel) renderActiveWord(target string) string {
	targetRunes := []rune(target)
	if len(targetRunes) == 0 {
		return m.styles.active.Render(target)
	}

	cursor := len([]rune(strings.TrimSpace(m.current)))
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= len(targetRunes) {
		return m.styles.active.Render(target)
	}

	var b strings.Builder
	if cursor > 0 {
		b.WriteString(m.styles.activeTyped.Render(string(targetRunes[:cursor])))
	}
	b.WriteString(m.styles.activeCursor.Render(string(targetRunes[cursor])))
	if cursor+1 < len(targetRunes) {
		b.WriteString(m.styles.active.Render(string(targetRunes[cursor+1:])))
	}
	return b.String()
}

func (m wordSessionModel) renderInputWord() string {
	if m.wordIndex >= len(m.words) {
		return m.styles.input.Render(m.current)
	}
	target := []rune(m.words[m.wordIndex])
	typed := []rune(m.current)
	var b strings.Builder
	for i, ch := range typed {
		if i < len(target) && ch != target[i] {
			b.WriteString(m.styles.inputBad.Render(string(ch)))
		} else {
			b.WriteString(m.styles.input.Render(string(ch)))
		}
	}
	return b.String()
}

func (m *wordSessionModel) commitCurrentWord() {
	if m.wordIndex >= len(m.words) {
		return
	}

	currentInput := strings.TrimSpace(m.current)
	if currentInput == "" {
		m.status = "Type the current word before advancing."
		return
	}
	targetWord := m.words[m.wordIndex]
	matched := currentInput == targetWord
	if m.strict && !matched {
		m.status = ""
		return
	}

	if !matched {
		m.totalErrors++
	} else {
		m.status = ""
	}

	m.uncorrectedErrors += countRuneMismatches([]rune(targetWord), []rune(currentInput))

	m.typedWords = append(m.typedWords, currentInput)
	m.wordMatches = append(m.wordMatches, matched)
	m.wordIndex++
	m.current = ""
	m.appendWPMSample()
}

func countRuneMismatches(target, typed []rune) int {
	n := len(target)
	if len(typed) < n {
		n = len(typed)
	}
	mismatches := 0
	for i := 0; i < n; i++ {
		if target[i] != typed[i] {
			mismatches++
		}
	}
	if len(typed) > len(target) {
		mismatches += len(typed) - len(target)
	}
	if len(target) > len(typed) {
		mismatches += len(target) - len(typed)
	}
	return mismatches
}

func (m *wordSessionModel) appendWPMSample() {
	elapsedMinutes := m.now().Sub(m.startedAt).Minutes()
	if elapsedMinutes <= 0 {
		return
	}
	charCount := len([]rune(strings.Join(m.typedWords, " ")))
	gross := (float64(charCount) / 5.0) / elapsedMinutes
	m.wpmSamples = append(m.wpmSamples, gross)
}

func (m wordSessionModel) isDone() bool {
	return m.wordIndex >= len(m.words) && len(m.words) > 0
}

func (m wordSessionModel) result() wordSessionResult {
	ended := m.endedAt
	if ended.IsZero() {
		ended = m.now().UTC()
	}
	return wordSessionResult{
		TypedText:         strings.Join(m.typedWords, " "),
		StartedAt:         m.startedAt,
		EndedAt:           ended,
		Completed:         m.isDone() && !m.aborted,
		Aborted:           m.aborted,
		WPMSamples:        append([]float64(nil), m.wpmSamples...),
		Strict:            m.strict,
		TotalErrors:       m.totalErrors,
		TotalKeystrokes:   m.totalKeystrokes,
		CorrectKeystrokes: m.correctKeystrokes,
		UncorrectedErrors: m.uncorrectedErrors,
	}
}

func runWordSession(input io.Reader, output io.Writer, prompt model.Prompt, strict, indefinite bool, now func() time.Time) (wordSessionResult, error) {
	m := newWordSessionModel(prompt, strict, now, indefinite)
	if len(m.words) == 0 {
		return wordSessionResult{}, fmt.Errorf("prompt contains no words")
	}

	// Full clear + cursor home so each session (including indefinite rounds) starts uncluttered.
	fmt.Fprint(output, "\x1b[2J\x1b[H")

	p := tea.NewProgram(
		m,
		tea.WithInput(input),
		tea.WithOutput(output),
	)
	finalModel, err := p.Run()
	if err != nil {
		return wordSessionResult{}, err
	}

	fm, ok := finalModel.(wordSessionModel)
	if !ok {
		return wordSessionResult{}, fmt.Errorf("unexpected final session model")
	}
	return fm.result(), nil
}

func (m *wordSessionModel) appendRunes(runes []rune) {
	if m.wordIndex >= len(m.words) {
		return
	}

	target := []rune(m.words[m.wordIndex])
	current := []rune(m.current)
	for _, r := range runes {
		pos := len(current)
		m.totalKeystrokes++
		expected := rune(0)
		if pos < len(target) {
			expected = target[pos]
		}
		matched := pos < len(target) && r == expected
		if matched {
			m.correctKeystrokes++
		}
		if m.strict {
			if !matched {
				continue
			}
		}
		current = append(current, r)
	}
	m.current = string(current)
}

func removeLastRune(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	return string(runes[:len(runes)-1])
}
