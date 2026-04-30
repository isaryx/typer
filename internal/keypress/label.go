package keypress

import (
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

// maxDisplayLabelRunes caps pasted / IME text so the UI and history stay usable.
const maxDisplayLabelRunes = 72

// DisplayLabel returns a readable string for a key event. It is based on
// bubbletea's KeyMsg.String() with small adjustments for empty/space and
// truncation for very long input (e.g. bracketed paste).
func DisplayLabel(msg tea.KeyMsg) string {
	s := msg.String()
	if s == "" {
		return "?"
	}
	if s == " " {
		return "space"
	}
	return truncateRunes(s, maxDisplayLabelRunes)
}

func truncateRunes(s string, max int) string {
	if max <= 0 {
		return s
	}
	// One full rune count avoids mis-handling max==1 (old code treated every
	// non-empty string as "…") and matches "at most max runes in output" when
	// we add an ellipsis.
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		if n >= max-1 {
			b.WriteRune('…')
			return b.String()
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}

// AppendHistory appends label and keeps at most max entries (oldest dropped).
func AppendHistory(h []string, label string, max int) []string {
	h = append(h, label)
	if len(h) > max {
		return h[len(h)-max:]
	}
	return h
}
