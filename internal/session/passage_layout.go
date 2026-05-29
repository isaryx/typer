// Passage layout computes plain-text wrap geometry for the typing passage frame.
// It must not read or modify typingState counters — session metrics flow only through typingState.
package session

import (
	"charm.land/lipgloss/v2"
)

const plainWordSeparatorWidth = 1

type passageLayout struct {
	lineWidth      int
	firstWordIdx   []int
	lineCount      int
	wordPlainWidth []int
	wordLine       []int
	wordCol        []int
}

func buildPlainLayout(words []string, lineWidth int) passageLayout {
	if lineWidth < 1 {
		lineWidth = 1
	}
	n := len(words)
	pl := passageLayout{
		lineWidth:      lineWidth,
		wordPlainWidth: make([]int, n),
		wordLine:       make([]int, n),
		wordCol:        make([]int, n),
	}
	for i, w := range words {
		pl.wordPlainWidth[i] = lipgloss.Width(w)
	}
	if n == 0 {
		return pl
	}

	var firstWordIdx []int
	cur := 0
	lineStartIdx := 0
	lineNum := 0

	for i := range words {
		ww := pl.wordPlainWidth[i]
		sep := 0
		if i > lineStartIdx {
			sep = plainWordSeparatorWidth
		}
		if cur+sep+ww > lineWidth && i > lineStartIdx {
			firstWordIdx = append(firstWordIdx, lineStartIdx)
			lineNum++
			lineStartIdx = i
			cur = 0
			sep = 0
		}
		pl.wordLine[i] = lineNum
		if i == lineStartIdx {
			pl.wordCol[i] = 0
			cur = ww
		} else {
			pl.wordCol[i] = cur + sep
			cur = pl.wordCol[i] + ww
		}
	}
	firstWordIdx = append(firstWordIdx, lineStartIdx)
	pl.firstWordIdx = firstWordIdx
	pl.lineCount = len(firstWordIdx)
	return pl
}

func plainRunePrefixWidth(word string, runeCount int) int {
	if runeCount <= 0 {
		return 0
	}
	ru := []rune(word)
	if runeCount >= len(ru) {
		return lipgloss.Width(word)
	}
	return lipgloss.Width(string(ru[:runeCount]))
}

func (pl passageLayout) lineIndexForWord(wordIdx int) int {
	if wordIdx < 0 || wordIdx >= len(pl.wordLine) {
		return 0
	}
	return pl.wordLine[wordIdx]
}

func (pl passageLayout) lastWordOnLine(lineIdx int) int {
	if lineIdx < 0 || lineIdx >= len(pl.firstWordIdx) {
		return -1
	}
	if lineIdx+1 < len(pl.firstWordIdx) {
		return pl.firstWordIdx[lineIdx+1] - 1
	}
	return len(pl.wordPlainWidth) - 1
}
