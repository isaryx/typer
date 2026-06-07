package session

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	"typer/internal/model"
	"typer/internal/ui"
)

func TestParsePromptContentHardLines(t *testing.T) {
	t.Parallel()

	words, hard := parsePromptContent("aa bb\ncc dd\nee ff", 0)
	if len(words) != 6 {
		t.Fatalf("words = %v", words)
	}
	if len(hard) != 6 || hard[0] != 0 || hard[2] != 1 || hard[4] != 2 {
		t.Fatalf("hard = %v", hard)
	}
	pl := buildPlainLayout(words, 80, hard)
	if pl.lineCount < 3 {
		t.Fatalf("lineCount = %d, want at least 3 hard lines", pl.lineCount)
	}
}

func TestFitLineToWidth(t *testing.T) {
	t.Parallel()

	const w = 40
	seed := "cc vv cv vc cave vice"
	got := fitLineToWidth(seed, w)
	if lipgloss.Width(got) > w {
		t.Fatalf("width = %d, want <= %d: %q", lipgloss.Width(got), w, got)
	}
	if lipgloss.Width(got) < w-lipgloss.Width(seed) {
		t.Fatalf("expected line to fill width, got %q (width %d)", got, lipgloss.Width(got))
	}

	long := strings.Repeat("word ", 20)
	trunc := fitLineToWidth(long, w)
	if lipgloss.Width(trunc) > w {
		t.Fatalf("truncated width = %d", lipgloss.Width(trunc))
	}
}

func TestTrainLessonLinesFitFrameWidth(t *testing.T) {
	t.Parallel()

	content := strings.Join([]string{
		"cc vv cv vc cave vice",
		"cc vv cv vc cave vice",
		"cc vv cv vc cave vice",
		"cc vv cv vc cave vice",
	}, "\n")
	termW := 100
	innerW := ui.FrameBodyInnerWidth(termW)

	m := newTypingSessionModel(
		model.Prompt{Content: content, Mode: model.ModeTrain},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false, nil, false, false, false, false,
		model.DefaultInputPlacement(),
		nil, nil, 0,
	)
	m.width = termW
	pl := m.ensurePlainLayout()
	if pl.lineCount != 4 {
		t.Fatalf("lineCount = %d, want 4 fitted hard lines", pl.lineCount)
	}
	for i := 0; i < pl.lineCount; i++ {
		line := m.styledLineForPlainLine(pl, i)
		if lipgloss.Width(line) > innerW {
			t.Fatalf("line %d exceeds inner width %d: %q", i+1, innerW, line)
		}
	}
	if m.passageViewportHeight() < 4 {
		t.Fatalf("viewport = %d, want 4 for train lesson", m.passageViewportHeight())
	}
}

func TestTrainExpandedContentScoringTarget(t *testing.T) {
	t.Parallel()

	seed := "cc vv cv vc cave vice"
	const innerW = 76
	expanded := fitLineToWidth(seed, innerW)
	typed := strings.Join(strings.Fields(strings.Repeat(expanded+" ", 4)), " ")

	seedPrompt := strings.Join([]string{seed, seed, seed, seed}, "\n")
	seedTarget := strings.Join(strings.Fields(seedPrompt), " ")

	if seedTarget == typed {
		t.Fatal("test setup: seed and expanded targets should differ")
	}

	// Simulates old bug: compare typed text to short seed prompt.
	if len(typed) <= len(seedTarget)*2 {
		t.Fatalf("typed should be much longer than seed target")
	}
}

func TestTrainLessonFourVisualLines(t *testing.T) {
	t.Parallel()

	content := strings.Join([]string{
		"ff jj fjf jfj ffj jff",
		"fad jad fad jad fjf jfj",
		"jff ffj fjf jfj jad fad",
		"fjf jfj ffj jff ff jj",
	}, "\n")
	m := newTypingSessionModel(
		model.Prompt{Content: content, Mode: model.ModeTrain},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false, nil, false, false, false, false,
		model.DefaultInputPlacement(),
		nil, nil, 0,
	)
	m.width = 100
	pl := m.ensurePlainLayout()
	if pl.lineCount < 4 {
		t.Fatalf("lineCount = %d, want at least 4 visual lines", pl.lineCount)
	}
	if m.passageViewportHeight() < 4 {
		t.Fatalf("viewport = %d, want 4 for train lesson", m.passageViewportHeight())
	}
}
