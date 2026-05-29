package session

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"

	"typer/internal/model"
	"typer/internal/text"
	"typer/internal/ui"
)

const (
	// passageViewportLines is how many wrapped passage rows show inside the rounded frame.
	passageViewportLines = 3
	// passageScrollLeadLines keeps the active line this many rows above the viewport bottom
	// when possible so the window scrolls earlier and one line of upcoming text stays visible.
	passageScrollLeadLines = 1
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
	if mode == model.ModeHangman || m.hangman != nil {
		return fmt.Sprintf("typer · %s", mode)
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

// lineIndexForWord returns the wrapped line index containing words[wordIdx].
func lineIndexForWord(firstWordIdx []int, wordIdx int) int {
	if len(firstWordIdx) == 0 {
		return 0
	}
	for k := len(firstWordIdx) - 1; k >= 0; k-- {
		if firstWordIdx[k] <= wordIdx {
			return k
		}
	}
	return 0
}

// passageViewportStart picks the first visible line so the active line stays in view.
// With passageScrollLeadLines, the active row is placed one line above the viewport bottom
// when possible (early scroll / read-ahead below).
func passageViewportStart(activeLine, totalLines, viewportH int) int {
	if totalLines <= 0 || viewportH <= 0 {
		return 0
	}
	if totalLines <= viewportH {
		return 0
	}
	maxStart := totalLines - viewportH
	targetRow := viewportH - 1 - passageScrollLeadLines
	if targetRow < 0 {
		targetRow = 0
	}
	start := activeLine - targetRow
	if start < 0 {
		start = 0
	}
	if start > maxStart {
		start = maxStart
	}
	return start
}

func (m *typingSessionModel) promptInputStyled() string {
	return m.styles.promptPrefix.Render("> ") + m.renderInputWord()
}

// borderInsertionPrefix is "| …typed" up to the caret (pipe matches frame tone via border style).
func (m *typingSessionModel) borderInsertionPrefix() string {
	return m.styles.border.Render("|") + " " + m.renderInputWord()
}

// borderChromeCenterStyled is "| … |" between the horizontal rules (ASCII pipes read cleanly in most fonts).
func (m *typingSessionModel) borderChromeCenterStyled() string {
	return m.borderInsertionPrefix() + " " + m.styles.border.Render("|")
}

func (m *typingSessionModel) inputOnTopBorder() bool {
	if m.noInput || m.isDone() {
		return false
	}
	switch m.inputPlacement.V {
	case model.InputVerticalOnTopBorder, model.InputVerticalOnTopBorderDynamic:
		return true
	default:
		return false
	}
}

func (m *typingSessionModel) inputOnBottomBorder() bool {
	if m.noInput || m.isDone() {
		return false
	}
	switch m.inputPlacement.V {
	case model.InputVerticalOnBottomBorder, model.InputVerticalOnBottomBorderDynamic:
		return true
	default:
		return false
	}
}

// activeWordContentOffset returns the active word's start column and width in passage inner coordinates (0..lineWidth-1)
// when that word's wrapped line is visible in the passage viewport. Otherwise visible is false (caller should fall back to centered border input).
func (m *typingSessionModel) activeWordContentOffset(pl passageLayout, activeW int) (wordStart, wordWidth int, visible bool) {
	if len(m.words) == 0 || activeW >= len(m.words) {
		return 0, 0, false
	}
	if pl.lineCount == 0 {
		return 0, 0, false
	}
	activeLine := pl.lineIndexForWord(activeW)
	start := passageViewportStart(activeLine, pl.lineCount, passageViewportLines)
	end := start + passageViewportLines
	if end > pl.lineCount {
		end = pl.lineCount
	}
	if activeLine < start || activeLine >= end {
		return 0, 0, false
	}
	return pl.wordCol[activeW], pl.wordPlainWidth[activeW], true
}

// borderInputLeftDashes chooses how many ─ cells precede the input segment on the top/bottom border row.
// When dynamic is true, the segment is shifted so it roughly centers on the active word in the passage body; otherwise centered in the rule.
func (m *typingSessionModel) borderInputLeftDashes(inner int, centerStyled string, dynamic bool, pl passageLayout, activeW int) int {
	w := lipgloss.Width(centerStyled)
	if inner < w {
		return 0
	}
	rem := inner - w
	centered := rem / 2
	if !dynamic {
		return centered
	}
	wordStart, wordWidth, ok := m.activeWordContentOffset(pl, activeW)
	if !ok {
		return centered
	}
	// Frame rule is two cells wider than passage inner (╭╮ gutters); map content column c to middle coordinate c+1.
	wordCenter := 1 + wordStart + wordWidth/2
	left := wordCenter - w/2
	if left < 0 {
		left = 0
	}
	if left > rem {
		left = rem
	}
	return left
}

// footerMetaCombinedLabel combines "by Author" and @source for on-top (bottom border halves) and
// on-bottom (top-right half alongside word count).
func (m *typingSessionModel) footerMetaCombinedLabel() string {
	author := strings.TrimSpace(formatByCaption(m.prompt.Author))
	var src string
	if m.prompt.Mode == model.ModeQuote {
		src = strings.TrimSpace(text.QuoteFrameSourceCaption(m.prompt.Source))
	}
	switch {
	case author != "" && src != "":
		return author + " · " + src
	case author != "":
		return author
	case src != "":
		return src
	default:
		return ""
	}
}

// relocatedFooterBottomLine draws the bottom border when the top border is used for input (on-top),
// carrying word count on the left and footerMetaCombinedLabel on the right when present.
func (m *typingSessionModel) relocatedFooterBottomLine(inner int) string {
	if right := m.footerMetaCombinedLabel(); right != "" {
		return ui.RenderRoundedBottomHalves("", m.styles.border, m.styles.meta, m.wordCountTopLabel(), right, inner)
	}
	capPlain, d1 := ui.FitBottomCaption(m.wordCountTopLabel(), inner, 1)
	return ui.RenderRoundedBottomCaption("", m.styles.border, m.styles.meta, d1, capPlain, 1)
}

// passageWrappedLayout runs one soft-wrap pass for the passage body (same lines used for the frame
// and for ghostCaretViewCoords). Call once per View to avoid duplicate layout work.
// Implemented in tui_layout_cache.go.

// renderPassageFrame renders the rounded passage frame. startRow is the 0-based line index of the
// frame's top border in the overall view. visibleLines and firstWordIdx come from styledViewportLines.
func (m *typingSessionModel) renderPassageFrame(startRow int, pl passageLayout, visibleLines []string, firstWordIdx []int) (content string, cx, cy int, ok bool) {
	tw := m.wrapWidth()
	inner := tw - 2 // between ╭ and ╮ (excluding corner runes)
	lineWidth := pl.lineWidth

	wordLbl := m.wordCountTopLabel()
	onTopIn := m.inputOnTopBorder()
	onBottomIn := m.inputOnBottomBorder()
	combinedMeta := m.footerMetaCombinedLabel()

	activeW := m.wordIndex
	if len(m.words) > 0 && activeW >= len(m.words) {
		activeW = len(m.words) - 1
	}

	var topLine string
	switch {
	case onTopIn:
		cs := m.borderChromeCenterStyled()
		dynTop := m.inputPlacement.V == model.InputVerticalOnTopBorderDynamic
		leftD := m.borderInputLeftDashes(inner, cs, dynTop, pl, activeW)
		topLine = ui.RenderRoundedTopCenterBorderLeft("", m.styles.border, cs, inner, leftD)
		cx = ui.CenterBorderCaretXLeft(true, "", m.styles.border, inner, cs, m.borderInsertionPrefix(), leftD)
		cy = startRow
		ok = true
	case onBottomIn:
		// Input uses the bottom rule; combined author · @source on the top-right border (like default quote).
		if combinedMeta != "" {
			topLine = ui.RenderRoundedTopHalves("", m.styles.border, m.styles.meta, wordLbl, combinedMeta, inner)
		} else {
			topLine = ui.RenderRoundedTop("", m.styles.border, m.styles.meta, wordLbl, inner)
		}
	case m.prompt.Mode == model.ModeQuote:
		if cap := strings.TrimSpace(text.QuoteFrameSourceCaption(m.prompt.Source)); cap != "" {
			topLine = ui.RenderRoundedTopHalves("", m.styles.border, m.styles.meta, wordLbl, cap, inner)
		} else {
			topLine = ui.RenderRoundedTop("", m.styles.border, m.styles.meta, wordLbl, inner)
		}
	default:
		topLine = ui.RenderRoundedTop("", m.styles.border, m.styles.meta, wordLbl, inner)
	}

	contentW := lineWidth

	author := strings.TrimSpace(formatByCaption(m.prompt.Author))

	var b strings.Builder
	b.WriteString(topLine)
	b.WriteString("\n")
	for _, line := range visibleLines {
		b.WriteString(ui.RenderRoundedSide("", m.styles.border, contentW, line))
		b.WriteString("\n")
	}

	if onBottomIn {
		cs := m.borderChromeCenterStyled()
		dynBot := m.inputPlacement.V == model.InputVerticalOnBottomBorderDynamic
		leftD := m.borderInputLeftDashes(inner, cs, dynBot, pl, activeW)
		b.WriteString(ui.RenderRoundedBottomCenterBorderLeft("", m.styles.border, cs, inner, leftD))
		cx = ui.CenterBorderCaretXLeft(false, "", m.styles.border, inner, cs, m.borderInsertionPrefix(), leftD)
		cy = startRow + 1 + len(visibleLines)
		ok = true
		return b.String(), cx, cy, ok
	}

	if author != "" && !onTopIn {
		capPlain, d1 := ui.FitBottomCaption(author, inner, 1)
		b.WriteString(ui.RenderRoundedBottomCaption("", m.styles.border, m.styles.meta, d1, capPlain, 1))
		return b.String(), 0, 0, false
	}
	if onTopIn {
		b.WriteString(m.relocatedFooterBottomLine(inner))
		return b.String(), cx, cy, ok
	}
	b.WriteString(ui.RenderRoundedBottomPlain("", m.styles.border, inner))
	return b.String(), 0, 0, false
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
	if laneWordIndex >= len(m.words) || wordIdx != laneWordIndex || wordIdx >= len(m.wordLens) {
		return false
	}
	return utf8.RuneCountInString(partial) >= m.wordLens[wordIdx]
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
	st := m.styles.upcoming
	if userOnSpace {
		st = m.styles.activePlain.Underline(true)
	}
	return st.Render(" ")
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

// renderPassageWordSegment styles one word for passage layout; with a shadow trace,
// each segment is rendered via renderMergedWordPiece.
func (m *typingSessionModel) renderPassageWordSegment(i int, w string) string {
	if !m.hasShadowReplay() {
		return m.renderWordPiece(i, w)
	}
	return m.renderMergedWordPiece(i, w)
}

// collectWrappedWordLines lays out words into soft-wrapped lines using plain geometry
// and styled segments. Prefer ensurePlainLayout + styledViewportLines in the View hot path.
func (m *typingSessionModel) collectWrappedWordLines(lineWidth int, seg func(int, string) string) ([]string, []int) {
	pl := buildPlainLayout(m.words, lineWidth)
	var lines []string
	for li := 0; li < pl.lineCount; li++ {
		start := pl.firstWordIdx[li]
		end := pl.lastWordOnLine(li) + 1
		parts := make([]string, 0, end-start)
		for i := start; i < end; i++ {
			parts = append(parts, seg(i, m.words[i]))
		}
		lines = append(lines, m.joinLineParts(parts, start))
	}
	return lines, pl.firstWordIdx
}

func (m *typingSessionModel) renderWords(lineWidth int) string {
	m.layoutWidth = 0
	m.plainLayout = buildPlainLayout(m.words, lineWidth)
	m.layoutWidth = lineWidth
	lines, _, _ := m.passageWrappedLayout()
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
	// ahead of the ghost — we must keep showing the ghost here (otherwise the ghost
	// caret highlight vanishes until the shadow catches up).
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
		return m.renderShadowActiveWord(w)
	}
	return m.styles.upcoming.Render(w)
}

// Both on the same word: user underline on the target letter; ghost position uses the terminal cursor (see ghostCaretViewCoords).
func (m *typingSessionModel) renderMergedActiveWord(target string) string {
	tu := []rune(target)
	cu := len([]rune(m.current))
	if len(tu) == 0 {
		return m.styles.active.Render(target)
	}
	var b strings.Builder
	for j := 0; j < len(tu); j++ {
		ch := string(tu[j])
		userAt := cu < len(tu) && j == cu
		var st lipgloss.Style
		if j < cu {
			st = m.styles.activeTyped
		} else {
			st = m.styles.activePlain
		}
		if userAt {
			st = st.Underline(true)
		}
		b.WriteString(st.Render(ch))
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
			b.WriteString(m.styles.activePlain.Underline(true).Render(" "))
		}
		return b.String()
	}

	var b strings.Builder
	if cursor > 0 {
		b.WriteString(m.styles.activeTyped.Render(string(targetRunes[:cursor])))
	}
	for j := cursor; j < len(targetRunes); j++ {
		ch := string(targetRunes[j])
		st := m.styles.activePlain
		if j == cursor {
			st = st.Underline(true)
		}
		b.WriteString(st.Render(ch))
	}
	return b.String()
}

func (m *typingSessionModel) renderBlinkBlockCaret() string {
	st := lipgloss.NewStyle().Foreground(m.inputCursorColor())
	if !typingReduceMotion() {
		st = st.Blink(true)
	}
	return st.Render("█")
}

// typingReduceMotion reports whether TYPER_REDUCE_MOTION is set (e.g. 1, true, yes).
// When true, the inline █ caret does not use ANSI blink (accessibility / vestibular comfort).
func typingReduceMotion() bool {
	s := strings.TrimSpace(strings.ToLower(os.Getenv("TYPER_REDUCE_MOTION")))
	switch s {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// renderInputWord groups consecutive good/bad runes into single styled runs so
// each render emits one ANSI escape per run instead of per rune.
func (m *typingSessionModel) renderInputWord() string {
	body := m.renderInputWordBody()
	if m.showInputChrome() {
		return body + m.renderBlinkBlockCaret()
	}
	return body
}

func (m *typingSessionModel) renderInputWordBody() string {
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

// ghostCaretViewCoords returns the terminal cell (x, y) for the shadow caret on passage text,
// matching RenderRoundedSide geometry (│ + inner). passStart is the view line index of the passage top border row.
func (m *typingSessionModel) ghostCaretViewCoords(passStart int, pl passageLayout, firstWordIdx []int) (x, y int, ok bool) {
	if !m.hasShadowReplay() || len(m.words) == 0 {
		return 0, 0, false
	}
	wi := m.shadowWordIndex
	if wi < 0 || wi >= len(m.words) {
		return 0, 0, false
	}
	if pl.lineCount == 0 {
		return 0, 0, false
	}
	ghostLine := pl.lineIndexForWord(wi)
	vpStart := passageViewportStart(ghostLine, pl.lineCount, passageViewportLines)
	vpEnd := vpStart + passageViewportLines
	if vpEnd > pl.lineCount {
		vpEnd = pl.lineCount
	}
	if ghostLine < vpStart || ghostLine >= vpEnd {
		return 0, 0, false
	}

	innerX := pl.wordCol[wi]
	cg := len([]rune(m.shadowCurrent))
	if cg > m.wordLens[wi] {
		cg = m.wordLens[wi]
	}
	if cg >= m.wordLens[wi] {
		innerX += pl.wordPlainWidth[wi]
	} else {
		innerX += plainRunePrefixWidth(m.words[wi], cg)
	}

	x = ui.PassageSideInnerStartCells + innerX
	y = passStart + 1 + (ghostLine - vpStart)
	return x, y, true
}

// Ghost alone on this word: plain passage styling; ghost caret uses the terminal cursor.
func (m *typingSessionModel) renderShadowActiveWord(target string) string {
	tu := []rune(target)
	if len(tu) == 0 {
		return m.styles.upcoming.Render(target)
	}
	var b strings.Builder
	for j := 0; j < len(tu); j++ {
		b.WriteString(m.styles.upcoming.Render(string(tu[j])))
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
