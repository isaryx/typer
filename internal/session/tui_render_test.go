package session

import (
	"strings"
	"testing"
	"time"

	"typer/internal/model"
)

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
