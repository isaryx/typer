package keypress

import (
	"strings"

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
	if max == 1 {
		for range s {
			return "…"
		}
		return s
	}
	var b strings.Builder
	written := 0
	for _, r := range s {
		if written >= max-1 {
			b.WriteRune('…')
			return b.String()
		}
		b.WriteRune(r)
		written++
	}
	return s
}

// AppendHistory appends label and keeps at most max entries (oldest dropped).
func AppendHistory(h []string, label string, max int) []string {
	h = append(h, label)
	if len(h) > max {
		return h[len(h)-max:]
	}
	return h
}
