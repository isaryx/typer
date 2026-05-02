package keypress

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestDisplayLabel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		msg  tea.KeyPressMsg
		want string
	}{
		{
			name: "alt_rune",
			msg:  tea.KeyPressMsg{Code: 'a', Text: "a", Mod: tea.ModAlt},
			want: "alt+a",
		},
		{
			name: "ctrl_k",
			msg:  tea.KeyPressMsg{Code: 'k', Mod: tea.ModCtrl},
			want: "ctrl+k",
		},
		{
			name: "space",
			msg:  tea.KeyPressMsg{Code: tea.KeySpace},
			want: "space",
		},
		{
			name: "lowercase_a",
			msg:  tea.KeyPressMsg{Code: 'a', Text: "a"},
			want: "a",
		},
		{
			name: "esc",
			msg:  tea.KeyPressMsg{Code: tea.KeyEsc},
			want: "esc",
		},
		{
			name: "ctrl_c",
			msg:  tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl},
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
	if got := DisplayLabel(tea.KeyPressMsg{}); got != "?" {
		t.Fatalf("DisplayLabel(nil runes) = %q, want %q", got, "?")
	}
}

func TestDisplayLabel_truncatesLongInput(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("a", maxDisplayLabelRunes+20)
	msg := tea.KeyPressMsg{Text: long}
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
