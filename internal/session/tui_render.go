package session

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"

	"typer/internal/model"
	"typer/internal/ui"
)

func (m *typingSessionModel) termWidth() int {
	if m.width > 0 {
		return m.width
	}
	return 80
}

// wrapWidth is min(terminal width, ui.MaxContentWidth) so content does not stretch across ultra-wide screens.
func (m *typingSessionModel) wrapWidth() int {
	tw := m.termWidth()
	if tw > ui.MaxContentWidth {
		return ui.MaxContentWidth
	}
	return tw
}

// expectedNextKeystroke returns the next key that advances the prompt (caret-aligned).
func (m *typingSessionModel) expectedNextKeystroke() (rune, bool) {
	if m.wordIndex >= len(m.words) {
		return 0, false
	}
	target := []rune(m.words[m.wordIndex])
	typed := []rune(m.current)
	pos := len(typed)
	if pos < len(target) {
		return target[pos], true
	}
	if pos > len(target) {
		return ' ', true
	}
	return ' ', true
}

func (m *typingSessionModel) suggestedFinger() Finger {
	r, ok := m.expectedNextKeystroke()
	if !ok {
		return FingerUnknown
	}
	return FingerForRune(r)
}

// promptInnerWidth is the line width for passage text between "│ " and " │" manual borders.
func (m *typingSessionModel) promptInnerWidth() int {
	return ui.FrameBodyInnerWidth(m.wrapWidth())
}

func (m *typingSessionModel) sessionHeadingLine() string {
	mode := strings.TrimSpace(m.prompt.Mode)
	if mode == "" {
		mode = "unknown"
	}
	strictLabel := "strict"
	if !m.strict {
		strictLabel = "non-strict"
	}
	return fmt.Sprintf("typer · %s · %s", mode, strictLabel)
}

func (m *typingSessionModel) wordCountTopLabel() string {
	k, n := m.wordIndex, len(m.words)
	full := fmt.Sprintf("%d/%d words", k, n)
	inner := m.wrapWidth() - 2
	pad := " " + full + " "
	if lipgloss.Width(pad) <= inner {
		return full
	}
	return fmt.Sprintf("%d/%d", k, n)
}

func formatByCaption(author string) string {
	s := strings.TrimSpace(author)
	if s == "" {
		return ""
	}
	low := strings.ToLower(s)
	if strings.HasPrefix(low, "by ") {
		return s
	}
	return "by " + s
}

func formatReplayStatsSegment(sr *model.SessionResult) string {
	if sr == nil {
		return ""
	}
	return fmt.Sprintf(
		"net %.2f wpm · %s · %.1f%% acc · %d err",
		sr.Metrics.NetWPM,
		formatReplayDuration(sr.ElapsedMS),
		sr.Metrics.Accuracy,
		sr.Metrics.Errors,
	)
}

func formatReplayCompactLine(sr *model.SessionResult) string {
	if sr == nil {
		return ""
	}
	stats := formatReplayStatsSegment(sr)
	lbl := formatReplaySessionLabel(sr)
	if lbl == "" {
		return "Replay | " + stats
	}
	return fmt.Sprintf("Replay · %s | %s", lbl, stats)
}

func (m *typingSessionModel) renderPassageFrame() string {
	tw := m.wrapWidth()
	inner := tw - 2 // between ╭ and ╮ (excluding corner runes)

	wordLbl := m.wordCountTopLabel()
	topLine := ui.RenderRoundedTop("", m.styles.border, m.styles.meta, wordLbl, inner)

	passage := m.renderMergedWords(m.promptInnerWidth())
	lines := strings.Split(passage, "\n")
	contentW := m.promptInnerWidth()

	var b strings.Builder
	b.WriteString(topLine)
	b.WriteString("\n")
	for _, pl := range lines {
		b.WriteString(ui.RenderRoundedSide("", m.styles.border, contentW, pl))
		b.WriteString("\n")
	}

	caption := formatByCaption(m.prompt.Author)
	if caption == "" {
		b.WriteString(ui.RenderRoundedBottomPlain("", m.styles.border, inner))
		return b.String()
	}

	capPlain, d1 := ui.FitBottomCaption(caption, inner, 1)
	b.WriteString(ui.RenderRoundedBottomCaption("", m.styles.border, m.styles.meta, d1, capPlain, 1))
	return b.String()
}

func (m *typingSessionModel) renderWordPiece(i int, w string) string {
	switch {
	case i < m.wordIndex:
		if i < len(m.wordMatches) && !m.wordMatches[i] {
			return m.styles.completedBad.Render(w)
		}
		return m.styles.completed.Render(w)
	case i == m.wordIndex:
		return m.renderActiveWord(w, i == len(m.words)-1)
	default:
		return m.styles.upcoming.Render(w)
	}
}

// partialMatchesFullWord is true when the typing lane is still on wordIdx and partial
// contains every rune of that word (Space not yet applied). Used for both user and ghost.
func (m *typingSessionModel) partialMatchesFullWord(wordIdx, laneWordIndex int, partial string) bool {
	if laneWordIndex >= len(m.words) || wordIdx != laneWordIndex {
		return false
	}
	target := m.words[wordIdx]
	return utf8.RuneCountInString(partial) >= len([]rune(target))
}

// activeWordInputComplete is true when the user has typed every rune of the
// active word but has not yet committed with Space.
func (m *typingSessionModel) activeWordInputComplete() bool {
	if m.wordIndex >= len(m.words) {
		return false
	}
	return m.partialMatchesFullWord(m.wordIndex, m.wordIndex, m.current)
}

// shadowWordInputCompleteFor is true when the ghost has typed every rune of words[afterIdx]
// but the replay has not yet advanced past this word (Space not applied in trace).
func (m *typingSessionModel) shadowWordInputCompleteFor(afterIdx int) bool {
	if !m.hasShadowReplay() {
		return false
	}
	return m.partialMatchesFullWord(afterIdx, m.shadowWordIndex, m.shadowCurrent)
}

// interWordSeparator is the glue between word at afterIdx and word at afterIdx+1.
// When the user or ghost has finished that word, the corresponding caret sits on this space.
func (m *typingSessionModel) interWordSeparator(afterIdx int) string {
	userOnSpace := afterIdx == m.wordIndex && afterIdx < len(m.words) && m.activeWordInputComplete()
	ghostOnSpace := m.shadowWordInputCompleteFor(afterIdx)
	if !userOnSpace && !ghostOnSpace {
		return " "
	}
	var marks string
	if userOnSpace {
		marks += markUserCaret
	}
	if ghostOnSpace {
		marks += markGhostCaret
	}
	if userOnSpace {
		return m.styles.activePlain.Render(" " + marks)
	}
	return m.styles.upcoming.Render(" " + marks)
}

func (m *typingSessionModel) joinLineParts(parts []string, firstWordIndex int) string {
	if len(parts) == 0 {
		return ""
	}
	var b strings.Builder
	for j, part := range parts {
		if j > 0 {
			b.WriteString(m.interWordSeparator(firstWordIndex + j - 1))
		}
		b.WriteString(part)
	}
	return b.String()
}

func (m *typingSessionModel) renderWords(lineWidth int) string {
	if lineWidth < 1 {
		lineWidth = 1
	}
	var lines []string
	var parts []string
	cur := 0
	lineStartIdx := 0

	for i, w := range m.words {
		seg := m.renderWordPiece(i, w)
		sw := lipgloss.Width(seg)
		sep := 0
		if len(parts) > 0 {
			sep = lipgloss.Width(m.interWordSeparator(i - 1))
		}
		if cur+sep+sw > lineWidth && len(parts) > 0 {
			lines = append(lines, m.joinLineParts(parts, lineStartIdx))
			parts = nil
			cur = 0
			sep = 0
		}
		if len(parts) == 0 {
			lineStartIdx = i
		}
		if len(parts) > 0 {
			cur += sep
		}
		parts = append(parts, seg)
		cur += sw
	}
	if len(parts) > 0 {
		lines = append(lines, m.joinLineParts(parts, lineStartIdx))
	}
	return strings.Join(lines, "\n")
}

// renderMergedWords draws user + ghost on the same passage; without a trace it matches renderWords.
func (m *typingSessionModel) renderMergedWords(lineWidth int) string {
	if lineWidth < 1 {
		lineWidth = 1
	}
	if !m.hasShadowReplay() {
		return m.renderWords(lineWidth)
	}
	var lines []string
	var parts []string
	cur := 0
	lineStartIdx := 0
	for i, w := range m.words {
		seg := m.renderMergedWordPiece(i, w)
		sw := lipgloss.Width(seg)
		sep := 0
		if len(parts) > 0 {
			sep = lipgloss.Width(m.interWordSeparator(i - 1))
		}
		if cur+sep+sw > lineWidth && len(parts) > 0 {
			lines = append(lines, m.joinLineParts(parts, lineStartIdx))
			parts = nil
			cur = 0
			sep = 0
		}
		if len(parts) == 0 {
			lineStartIdx = i
		}
		if len(parts) > 0 {
			cur += sep
		}
		parts = append(parts, seg)
		cur += sw
	}
	if len(parts) > 0 {
		lines = append(lines, m.joinLineParts(parts, lineStartIdx))
	}
	return strings.Join(lines, "\n")
}

func (m *typingSessionModel) renderMergedWordPiece(i int, w string) string {
	if !m.hasShadowReplay() {
		return m.renderWordPiece(i, w)
	}

	uAct := i == m.wordIndex
	gAct := i == m.shadowWordIndex

	// User finished word i only when their caret moved on *and* the ghost is not still
	// typing this word. If i < wordIndex but shadowWordIndex == i, the user committed
	// ahead of the ghost — we must keep showing the ghost here (otherwise the overline
	// vanishes until the shadow catches up).
	if i < m.wordIndex && m.shadowWordIndex != i {
		return m.renderWordPiece(i, w)
	}
	if uAct && gAct {
		return m.renderMergedActiveWord(w)
	}
	if uAct && m.shadowWordIndex > i {
		return m.renderMergedActiveGhostCommitted(w, i)
	}
	if uAct {
		return m.renderActiveWord(w, i == len(m.words)-1)
	}
	if i < m.shadowWordIndex && m.wordIndex < i {
		return m.renderGhostCompletedAt(i, w)
	}
	if gAct {
		return m.renderShadowActiveWord(w, i == len(m.words)-1)
	}
	return m.styles.upcoming.Render(w)
}

// Both on the same word: only combining marks differ by column; rune styling does not depend on ghost position.
func (m *typingSessionModel) renderMergedActiveWord(target string) string {
	tu := []rune(target)
	cu := len([]rune(m.current))
	cg := len([]rune(m.shadowCurrent))
	if len(tu) == 0 {
		return m.styles.active.Render(target)
	}
	// Do not delegate to renderActiveWord when cu >= len(tu): that path omits the ghost
	// overline while the user has finished the runes but not yet pressed Space.
	var b strings.Builder
	for j := 0; j < len(tu); j++ {
		ch := string(tu[j])
		marks := ""
		if cu < len(tu) && j == cu {
			marks += markUserCaret
		}
		if cg < len(tu) && j == cg {
			marks += markGhostCaret
		}
		if j < cu {
			b.WriteString(m.styles.activeTyped.Render(ch + marks))
		} else {
			b.WriteString(m.styles.activePlain.Render(ch + marks))
		}
	}
	return b.String()
}

// User still typing a word the ghost has already finished — prompt stays target-only (same as a normal active word for user).
func (m *typingSessionModel) renderMergedActiveGhostCommitted(target string, wordIdx int) string {
	return m.renderActiveWord(target, wordIdx == len(m.words)-1)
}

func (m *typingSessionModel) renderGhostCompletedAt(_ int, w string) string {
	// Ghost progress does not change passage styling — looks like any word not yet reached.
	return m.styles.upcoming.Render(w)
}

func (m *typingSessionModel) renderActiveWord(target string, isLastWord bool) string {
	targetRunes := []rune(target)
	if len(targetRunes) == 0 {
		return m.styles.active.Render(target)
	}

	cursor := utf8.RuneCountInString(m.current)
	if cursor >= len(targetRunes) {
		var b strings.Builder
		b.WriteString(m.styles.activeTyped.Render(string(targetRunes)))
		if isLastWord {
			b.WriteString(m.styles.activePlain.Render(" " + markUserCaret))
		}
		return b.String()
	}

	var b strings.Builder
	if cursor > 0 {
		b.WriteString(m.styles.activeTyped.Render(string(targetRunes[:cursor])))
	}
	for j := cursor; j < len(targetRunes); j++ {
		ch := string(targetRunes[j])
		marks := ""
		if j == cursor {
			marks = markUserCaret
		}
		b.WriteString(m.styles.activePlain.Render(ch + marks))
	}
	return b.String()
}

// renderInputWord groups consecutive good/bad runes into single styled runs so
// each render emits one ANSI escape per run instead of per rune.
func (m *typingSessionModel) renderInputWord() string {
	if m.wordIndex >= len(m.words) {
		return m.styles.input.Render(m.current)
	}
	target := []rune(m.words[m.wordIndex])
	typed := []rune(m.current)
	if len(typed) == 0 {
		return ""
	}

	var b strings.Builder
	runStart := 0
	runBad := len(target) > 0 && typed[0] != target[0]
	isBad := func(i int) bool {
		return i < len(target) && typed[i] != target[i]
	}
	flush := func(end int) {
		segment := string(typed[runStart:end])
		if runBad {
			b.WriteString(m.styles.inputBad.Render(segment))
		} else {
			b.WriteString(m.styles.input.Render(segment))
		}
	}
	for i := 1; i < len(typed); i++ {
		if bad := isBad(i); bad != runBad {
			flush(i)
			runStart = i
			runBad = bad
		}
	}
	flush(len(typed))
	return b.String()
}

func (m *typingSessionModel) hasShadowReplay() bool {
	return len(m.shadowTrace) > 0
}

// Ghost alone on this word: overline moves with the ghost caret; when the ghost has
// finished all runes, the caret sits on the following space (see interWordSeparator).
func (m *typingSessionModel) renderShadowActiveWord(target string, isLastWord bool) string {
	tu := []rune(target)
	cg := len([]rune(m.shadowCurrent))
	if len(tu) == 0 {
		return m.styles.upcoming.Render(target)
	}
	if cg >= len(tu) {
		var b strings.Builder
		for j := 0; j < len(tu); j++ {
			b.WriteString(m.styles.upcoming.Render(string(tu[j])))
		}
		if isLastWord {
			b.WriteString(m.styles.upcoming.Render(" " + markGhostCaret))
		}
		return b.String()
	}
	var b strings.Builder
	for j := 0; j < len(tu); j++ {
		ch := string(tu[j])
		if j == cg {
			b.WriteString(m.styles.upcoming.Render(ch + markGhostCaret))
		} else {
			b.WriteString(m.styles.upcoming.Render(ch))
		}
	}
	return b.String()
}

// formatReplaySessionLabel is the suffix after "Replay ·" (session id when present, else time fallback).
func formatReplaySessionLabel(sr *model.SessionResult) string {
	if sr == nil {
		return ""
	}
	if sr.ID != "" {
		return sr.ID
	}
	if !sr.StartedAt.IsZero() {
		return sr.StartedAt.Local().Format("2006-01-02 15:04:05")
	}
	return ""
}

func formatReplayDuration(ms int64) string {
	if ms < 0 {
		return "0 ms"
	}
	if ms < 1000 {
		return fmt.Sprintf("%d ms", ms)
	}
	secs := ms / 1000
	if secs < 60 {
		if ms < 10_000 {
			return fmt.Sprintf("%.1f s", float64(ms)/1000)
		}
		return fmt.Sprintf("%d s", secs)
	}
	mins := secs / 60
	secs %= 60
	if mins < 60 {
		return fmt.Sprintf("%dm %02ds", mins, secs)
	}
	h := mins / 60
	mins %= 60
	return fmt.Sprintf("%dh %02dm %02ds", h, mins, secs)
}
