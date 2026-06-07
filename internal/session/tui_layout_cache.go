package session

import "typer/internal/model"

func (m *typingSessionModel) invalidatePlainLayout() {
	m.layoutWidth = 0
}

func (m *typingSessionModel) hardLineCount() int {
	if len(m.wordHardLine) == 0 {
		return 0
	}
	maxH := 0
	for _, h := range m.wordHardLine {
		if h+1 > maxH {
			maxH = h + 1
		}
	}
	return maxH
}

// passageViewportHeight returns how many passage rows are visible in the frame.
// Train lessons with multiple hard lines show all of them (e.g. 4 full drill rows).
func (m *typingSessionModel) passageViewportHeight() int {
	if m.prompt.Mode == model.ModeTrain {
		if n := m.hardLineCount(); n > passageViewportLines {
			if n > 6 {
				return 6
			}
			return n
		}
	}
	return passageViewportLines
}

func (m *typingSessionModel) canRefitWords() bool {
	return m.promptSeedContent != "" && m.wordIndex == 0 && m.current == "" && len(m.typedWords) == 0
}

func (m *typingSessionModel) replaceWords(words []string) {
	m.typingState.replaceWords(words)
}

func (m *typingSessionModel) maybeRefitWords() {
	if !m.canRefitWords() {
		return
	}
	lw := m.promptInnerWidth()
	if lw == m.wordsFitWidth {
		return
	}
	words, hard := parsePromptContent(m.promptSeedContent, lw)
	m.replaceWords(words)
	m.wordHardLine = hard
	m.wordsFitWidth = lw
	n := len(words)
	m.wordStyleGen = make([]uint64, n)
	m.styledAtGen = make([]uint64, n)
	m.styledSegments = make([]string, n)
}

func (m *typingSessionModel) ensurePlainLayout() passageLayout {
	m.maybeRefitWords()
	lw := m.promptInnerWidth()
	if m.layoutWidth != lw {
		m.plainLayout = buildPlainLayout(m.words, lw, m.wordHardLine)
		m.layoutWidth = lw
	}
	return m.plainLayout
}

func (m *typingSessionModel) bumpWordStyle(wordIdx int) {
	if wordIdx < 0 || wordIdx >= len(m.wordStyleGen) {
		return
	}
	m.wordStyleGen[wordIdx]++
}

func (m *typingSessionModel) bumpAllWordStyles() {
	for i := range m.wordStyleGen {
		m.wordStyleGen[i]++
	}
}

func (m *typingSessionModel) renderPassageWordSegmentCached(i int, w string) string {
	if i < 0 || i >= len(m.styledSegments) {
		return m.renderPassageWordSegment(i, w)
	}
	if m.styledAtGen[i] == m.wordStyleGen[i] && m.styledSegments[i] != "" {
		return m.styledSegments[i]
	}
	seg := m.renderPassageWordSegment(i, w)
	m.styledSegments[i] = seg
	m.styledAtGen[i] = m.wordStyleGen[i]
	return seg
}

func (m *typingSessionModel) styledLineForPlainLine(pl passageLayout, lineIdx int) string {
	start := pl.firstWordIdx[lineIdx]
	end := pl.lastWordOnLine(lineIdx) + 1
	parts := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		parts = append(parts, m.renderPassageWordSegmentCached(i, m.words[i]))
	}
	return m.joinLineParts(parts, start)
}

// styledViewportLines returns styled strings only for lines visible in the passage viewport.
func (m *typingSessionModel) styledViewportLines(pl passageLayout) (visibleLines []string, firstWordIdx []int) {
	firstWordIdx = pl.firstWordIdx
	if pl.lineCount == 0 {
		return nil, firstWordIdx
	}
	activeW := m.wordIndex
	if activeW >= len(m.words) {
		activeW = len(m.words) - 1
	}
	activeLine := pl.lineIndexForWord(activeW)
	vp := m.passageViewportHeight()
	start := passageViewportStart(activeLine, pl.lineCount, vp)
	end := start + vp
	if end > pl.lineCount {
		end = pl.lineCount
	}
	visibleLines = make([]string, end-start)
	for li := start; li < end; li++ {
		visibleLines[li-start] = m.styledLineForPlainLine(pl, li)
	}
	return visibleLines, firstWordIdx
}

func (m *typingSessionModel) renderFingerHandsFrameCached(f Finger) string {
	key, ok := m.expectedNextKeystroke()
	if !ok {
		return m.renderFingerHandsFrame(f)
	}
	if m.hasCachedFingerHands && key == m.lastFingerHintKey {
		return m.cachedFingerHands
	}
	frame := m.renderFingerHandsFrame(f)
	m.lastFingerHintKey = key
	m.cachedFingerHands = frame
	m.hasCachedFingerHands = true
	return frame
}

// test helpers for layout/render tests.

func (m *typingSessionModel) passageFrameForTest(startRow int) (cx, cy int, ok bool) {
	pl := m.ensurePlainLayout()
	visible, idx := m.styledViewportLines(pl)
	_, cx, cy, ok = m.renderPassageFrame(startRow, pl, visible, idx)
	return cx, cy, ok
}

func (m *typingSessionModel) ghostCaretForTest(passStart int) (x, y int, ok bool) {
	pl := m.ensurePlainLayout()
	return m.ghostCaretViewCoords(passStart, pl, pl.firstWordIdx)
}

// passageWrappedLayout builds styled lines for the full passage (tests and legacy callers).
func (m *typingSessionModel) passageWrappedLayout() (lines []string, firstWordIdx []int, lineWidth int) {
	pl := m.ensurePlainLayout()
	lineWidth = pl.lineWidth
	firstWordIdx = pl.firstWordIdx
	if pl.lineCount == 0 {
		return nil, firstWordIdx, lineWidth
	}
	lines = make([]string, pl.lineCount)
	for li := 0; li < pl.lineCount; li++ {
		lines[li] = m.styledLineForPlainLine(pl, li)
	}
	return lines, firstWordIdx, lineWidth
}
