package session

import (
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	"typer/internal/model"
)

func TestBuildPlainLayoutWrap(t *testing.T) {
	words := []string{"hello", "world", "foo", "bar"}
	pl := buildPlainLayout(words, 12)
	if pl.lineCount < 2 {
		t.Fatalf("lineCount = %d, want multiple lines for width 12", pl.lineCount)
	}
	if pl.wordCol[0] != 0 {
		t.Fatalf("wordCol[0] = %d, want 0", pl.wordCol[0])
	}
	if pl.wordPlainWidth[0] != lipgloss.Width("hello") {
		t.Fatalf("wordPlainWidth[0] = %d", pl.wordPlainWidth[0])
	}
}

func TestPlainLayoutStyledWidthInvariant(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: "hello world foo"},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false, nil, false, false, false, false,
		model.DefaultInputPlacement(),
		nil,
		nil,
	)
	m.width = 80
	pl := m.ensurePlainLayout()
	for i, w := range m.words {
		styled := m.renderPassageWordSegmentCached(i, w)
		if got := lipgloss.Width(styled); got != pl.wordPlainWidth[i] {
			t.Fatalf("word %d styled width %d != plain %d", i, got, pl.wordPlainWidth[i])
		}
	}
}

func TestPlainRunePrefixWidthASCII(t *testing.T) {
	if got := plainRunePrefixWidth("hello", 3); got != 3 {
		t.Fatalf("prefix width = %d, want 3", got)
	}
}
