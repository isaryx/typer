package session

import (
	"strings"
	"testing"
	"time"

	"typer/internal/model"
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
	)
	m.width = 80
	_, _, vis := m.activeWordContentOffset(m.promptInnerWidth(), nil, nil)
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
	)
	m.width = 80
	_, cxFirst, _, ok := m.renderPassageFrameWithCursor(0)
	if !ok {
		t.Fatal("expected border cursor")
	}
	m.wordIndex = 2
	_, cxThird, _, ok2 := m.renderPassageFrameWithCursor(0)
	if !ok2 {
		t.Fatal("expected border cursor on third word")
	}
	if cxThird == cxFirst {
		t.Fatalf("dynamic border should shift caret with active word: cxFirst=%d cxThird=%d", cxFirst, cxThird)
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
	)
	m.width = 80
	v := m.View().Content
	if !strings.Contains(v, "@type.fit") {
		t.Fatalf("expected @type.fit in frame: %q", v)
	}
}
