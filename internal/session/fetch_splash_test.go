package session

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"typer/internal/model"
)

func TestRunQuoteFetchSplash_returnsPrompt(t *testing.T) {
	var buf bytes.Buffer
	want := model.Prompt{Content: "hello there", Mode: model.ModeQuote}
	p, err := runQuoteFetchSplash(context.Background(), &buf, func() (model.Prompt, error) {
		return want, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Content != want.Content {
		t.Fatalf("prompt = %#v", p)
	}
	out := buf.String()
	if !strings.Contains(out, "\x1b[2K") || !strings.Contains(out, "\x1b[?25") {
		t.Fatalf("expected spinner terminal sequences in output: %q", out)
	}
}
