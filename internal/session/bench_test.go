package session

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"typer/internal/model"
)

func benchWords(n int) []string {
	// Deterministic word list from embedded-style vocabulary.
	pool := []string{
		"the", "quick", "brown", "fox", "jumps", "over", "lazy", "dog",
		"typing", "practice", "session", "words", "passage", "quote",
	}
	out := make([]string, n)
	for i := range out {
		out[i] = pool[i%len(pool)]
	}
	return out
}

func benchPassageContent(n int) string {
	return strings.Join(benchWords(n), " ")
}

func benchTypingModel(wordCount int, width int) *typingSessionModel {
	words := benchWords(wordCount)
	m := newTypingSessionModel(
		model.Prompt{Content: strings.Join(words, " "), Mode: model.ModeWords},
		false,
		func() time.Time { return time.Unix(1000, 0) },
		false, nil, false, false, false, false,
		model.DefaultInputPlacement(),
		nil,
		nil,
	)
	m.width = width
	return m
}

func BenchmarkApplyRunesSingleKey(b *testing.B) {
	words := []string{"hello"}
	s := newTypingState(words, false, func() time.Time { return time.Unix(100, 0) })
	b.ReportAllocs()
	for b.Loop() {
		s.current = ""
		s.applyRunes([]rune("h"))
	}
}

func BenchmarkApplyRunesPaste(b *testing.B) {
	words := []string{"hello"}
	s := newTypingState(words, false, func() time.Time { return time.Unix(100, 0) })
	paste := []rune("hello")
	b.ReportAllocs()
	for b.Loop() {
		s.current = ""
		s.applyRunes(paste)
	}
}

func BenchmarkPassageLayout(b *testing.B) {
	for _, width := range []int{80, 120} {
		b.Run(widthString(width), func(b *testing.B) {
			m := benchTypingModel(200, width)
			b.ReportAllocs()
			for b.Loop() {
				pl := m.ensurePlainLayout()
				m.styledViewportLines(pl)
			}
		})
	}
}

func BenchmarkViewKeystroke(b *testing.B) {
	m := benchTypingModel(200, 120)
	// Advance to middle of passage.
	for i := 0; i < 50; i++ {
		w := m.words[i]
		m.applyRunes([]rune(w))
		m.applyCommitWord()
	}
	b.ReportAllocs()
	for b.Loop() {
		m.applyRunes([]rune("x"))
		_ = m.View()
		m.applyBackspace()
	}
}

func widthString(w int) string {
	return fmt.Sprintf("w%d", w)
}
