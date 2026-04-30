package keypress

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDisplayLabel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		msg  tea.KeyMsg
		want string
	}{
		{
			name: "alt_rune",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}, Alt: true},
			want: "alt+a",
		},
		{
			name: "ctrl_k",
			msg:  tea.KeyMsg{Type: tea.KeyCtrlK},
			want: "ctrl+k",
		},
		{
			name: "space",
			msg:  tea.KeyMsg{Type: tea.KeySpace},
			want: "space",
		},
		{
			name: "lowercase_a",
			msg:  tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			want: "a",
		},
		{
			name: "esc",
			msg:  tea.KeyMsg{Type: tea.KeyEsc},
			want: "esc",
		},
		{
			name: "ctrl_c",
			msg:  tea.KeyMsg{Type: tea.KeyCtrlC},
			want: "ctrl+c",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := DisplayLabel(tc.msg)
			if got != tc.want {
				t.Fatalf("DisplayLabel() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDisplayLabel_emptyString(t *testing.T) {
	t.Parallel()
	// If String() is empty, we show "?".
	if got := DisplayLabel(tea.KeyMsg{Type: tea.KeyRunes, Runes: nil}); got != "?" {
		t.Fatalf("DisplayLabel(nil runes) = %q, want %q", got, "?")
	}
}

func TestDisplayLabel_truncatesLongInput(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("a", maxDisplayLabelRunes+20)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(long)}
	got := DisplayLabel(msg)
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("expected suffix …, got %q", got)
	}
	if got == long {
		t.Fatal("expected truncation")
	}
}

func TestTruncateRunes(t *testing.T) {
	t.Parallel()
	if got := truncateRunes("a", 1); got != "a" {
		t.Fatalf("truncateRunes(%q, 1) = %q, want %q", "a", got, "a")
	}
	if got := truncateRunes("ab", 1); got != "…" {
		t.Fatalf("truncateRunes(%q, 1) = %q, want %q", "ab", got, "…")
	}
	if got := truncateRunes("hi", 4); got != "hi" {
		t.Fatalf("truncateRunes(%q, 4) = %q, want %q", "hi", got, "hi")
	}
	long := strings.Repeat("x", 10)
	if got := truncateRunes(long, 4); got != "xxx…" {
		t.Fatalf("truncateRunes(10 x's, 4) = %q, want %q", got, "xxx…")
	}
}

func TestAppendHistory(t *testing.T) {
	t.Parallel()
	var h []string
	for i := 0; i < 5; i++ {
		h = AppendHistory(h, string(rune('a'+i)), 3)
	}
	want := []string{"c", "d", "e"}
	if len(h) != len(want) {
		t.Fatalf("len = %d, want %d", len(h), len(want))
	}
	for i := range want {
		if h[i] != want[i] {
			t.Fatalf("h[%d] = %q, want %q", i, h[i], want[i])
		}
	}
}
