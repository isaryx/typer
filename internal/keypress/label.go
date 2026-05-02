package keypress

import (
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

// maxDisplayLabelRunes caps pasted / IME text so the UI and history stay usable.
const maxDisplayLabelRunes = 512

// DisplayLabel returns a readable string for a key event (tea.Key String vs Keystroke
// for modifiers), with small adjustments for empty/space and truncation. Bracketed paste
// is not a key press in Bubble Tea v2; handle tea.PasteMsg separately.
func DisplayLabel(msg tea.KeyPressMsg) string {
	k := tea.Key(msg)
	if k.Mod == 0 && k.Text == "" && k.Code == 0 {
		return "?"
	}
	var s string
	if k.Mod != 0 {
		s = k.Keystroke()
	} else {
		s = k.String()
	}
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
