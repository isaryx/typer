package session

import (
	"strings"
	"testing"
	"time"

	"typer/internal/model"
	"typer/internal/ui"
)

func TestViewQuoteModeSourceRelocatedToBottomWhenInputOnTopBorder(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{
			Content: "hello world",
			Mode:    model.ModeQuote,
			Source:  "zenquotes",
			Author:  "Someone",
		},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
		false,
		false,
		false,
		false,
		model.InputPlacement{V: model.InputVerticalOnTopBorder, H: model.InputHorizontalCenter},
		nil,
		nil,
	)
	m.width = 88
	v := m.View().Content
	if strings.Contains(v, "> ") {
		t.Fatalf("on-top border mode should not use > prompt: %q", v)
	}
	if !strings.Contains(v, "|") {
		t.Fatalf("expected | … | input chrome on top border: %q", v)
	}
	if !strings.Contains(v, "@ZenQuotes") {
		t.Fatalf("expected @ZenQuotes relocated to bottom border: %q", v)
	}
	if !strings.Contains(v, "by Someone") || !strings.Contains(v, "by Someone · @ZenQuotes") {
		t.Fatalf("expected author and source combined on bottom border, got:\n%s", v)
	}
	idxTop := strings.Index(v, "╭")
	idxZen := strings.Index(v, "@ZenQuotes")
	if idxTop < 0 || idxZen < 0 || idxZen <= idxTop {
		t.Fatalf("expected frame before @ZenQuotes (bottom row): top=%d zen=%d", idxTop, idxZen)
	}
}

func TestViewOnBottomBorderCombinesAuthorAndSource(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{
			Content: "hello world",
			Mode:    model.ModeQuote,
			Source:  "zenquotes",
			Author:  "Someone",
		},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
		false,
		false,
		false,
		false,
		model.InputPlacement{V: model.InputVerticalOnBottomBorder, H: model.InputHorizontalCenter},
		nil,
		nil,
	)
	m.width = 88
	v := m.View().Content
	if !strings.Contains(v, "by Someone · @ZenQuotes") {
		t.Fatalf("expected author and source combined, got:\n%s", v)
	}
	if strings.Count(v, "@ZenQuotes") != 1 {
		t.Fatalf("expected single @ZenQuotes, got %d in:\n%s", strings.Count(v, "@ZenQuotes"), v)
	}
	i := strings.Index(v, "╭")
	if i < 0 {
		t.Fatal("expected frame start")
	}
	rest := v[i:]
	ln := rest
	if j := strings.Index(rest, "\n"); j >= 0 {
		ln = rest[:j]
	}
	if !strings.Contains(ln, "@ZenQuotes") || !strings.Contains(ln, "by Someone") {
		t.Fatalf("on-bottom: combined meta should be on top border row, got: %q", ln)
	}
}

func TestViewQuoteModeShowsSourceInTopBorder(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{
			Content: "hello world",
			Mode:    model.ModeQuote,
			Source:  "zenquotes",
			Author:  "Someone",
		},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
		false,
		false,
		false,
		false,
		model.InputPlacement{},
		nil,
		nil,
	)
	m.width = 88
	v := m.View().Content
	if !strings.Contains(v, "@ZenQuotes") {
		t.Fatalf("expected @ZenQuotes in top border, got:\n%s", v)
	}
	// First row of the rounded frame should include the source (top-right half)
	if i := strings.Index(v, "╭"); i < 0 {
		t.Fatalf("expected frame start")
	}
}

func TestViewQuoteModeSeedDoesNotShowRemoteCaption(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: "hello world", Mode: model.ModeQuote, Source: "seed"},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
		false,
		false,
		false,
		false,
		model.InputPlacement{},
		nil,
		nil,
	)
	m.width = 88
	v := m.View().Content
	if strings.Contains(v, "@ZenQuotes") || strings.Contains(v, "@type.fit") {
		t.Fatalf("bundled quote should not show remote API caption:\n%s", v)
	}
}

func TestActiveWordContentOffsetNoWords(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: "   "},
		false,
		func() time.Time { return time.Unix(0, 0) },
		false,
		nil,
		false,
		false,
		false,
		false,
		model.InputPlacement{},
		nil,
		nil,
	)
	m.width = 80
	pl := m.ensurePlainLayout()
	_, _, vis := m.activeWordContentOffset(pl, 0)
	if vis {
		t.Fatalf("expected visible=false with no words")
	}
}

func TestBorderDynamicCaretXMovesWithActiveWord(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: "aa bb cc"},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
		false,
		false,
		false,
		false,
		model.InputPlacement{V: model.InputVerticalOnTopBorderDynamic, H: model.InputHorizontalCenter},
		nil,
		nil,
	)
	m.width = 80
	cxFirst, _, ok := m.passageFrameForTest(0)
	if !ok {
		t.Fatal("expected border cursor")
	}
	m.wordIndex = 2
	cxThird, _, ok2 := m.passageFrameForTest(0)
	if !ok2 {
		t.Fatal("expected border cursor on third word")
	}
	if cxThird == cxFirst {
		t.Fatalf("dynamic border should shift caret with active word: cxFirst=%d cxThird=%d", cxFirst, cxThird)
	}
}

func TestGhostCaretXAtWordStartMatchesPassageInnerOffset(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: "hello world"},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		&model.SessionResult{TypingTrace: []model.ReplayEvent{{AtMS: 0, Kind: model.ReplayEventKey, Rune: "x"}}},
		false,
		false,
		false,
		false,
		model.InputPlacement{},
		nil,
		nil,
	)
	m.width = 80
	m.shadowWordIndex = 0
	m.shadowCurrent = ""
	x, _, ok := m.ghostCaretForTest(0)
	if !ok {
		t.Fatal("expected ghost in viewport")
	}
	if x != ui.PassageSideInnerStartCells {
		t.Fatalf("ghost caret x=%d want %d (RenderRoundedSide inner start)", x, ui.PassageSideInnerStartCells)
	}
}

func TestTypingReduceMotionEnv(t *testing.T) {
	cases := []struct {
		env  string
		want bool
	}{
		{"1", true},
		{"true", true},
		{"yes", true},
		{"on", true},
		{"", false},
		{"0", false},
		{"no", false},
	}
	for _, tc := range cases {
		t.Run(tc.env, func(t *testing.T) {
			if tc.env == "" {
				t.Setenv("TYPER_REDUCE_MOTION", "")
			} else {
				t.Setenv("TYPER_REDUCE_MOTION", tc.env)
			}
			if got := typingReduceMotion(); got != tc.want {
				t.Fatalf("TYPER_REDUCE_MOTION=%q: got %v want %v", tc.env, got, tc.want)
			}
		})
	}
}

func TestTypingReduceMotionInlineCaretHasNoBlinkSGR(t *testing.T) {
	t.Setenv("TYPER_REDUCE_MOTION", "1")
	m := newTypingSessionModel(
		model.Prompt{Content: "hi"},
		false,
		func() time.Time { return time.Unix(0, 0) },
		false,
		nil,
		false,
		false,
		false,
		false,
		model.InputPlacement{},
		nil,
		nil,
	)
	m.wordIndex = 0
	m.current = "h"
	s := m.renderBlinkBlockCaret()
	// Standard slow-blink and rapid-blink SGRs (lipgloss Blink uses one of these).
	if strings.Contains(s, "\x1b[5m") || strings.Contains(s, "\x1b[6m") {
		t.Fatalf("reduce-motion caret should not emit blink SGR: %q", s)
	}
}

func TestViewQuoteModeTypefitSourceLabel(t *testing.T) {
	m := newTypingSessionModel(
		model.Prompt{Content: "a b", Mode: model.ModeQuote, Source: "type.fit"},
		false,
		func() time.Time { return time.Unix(100, 0) },
		false,
		nil,
		false,
		false,
		false,
		false,
		model.InputPlacement{},
		nil,
		nil,
	)
	m.width = 80
	v := m.View().Content
	if !strings.Contains(v, "@type.fit") {
		t.Fatalf("expected @type.fit in frame: %q", v)
	}
}
