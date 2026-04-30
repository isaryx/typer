package text

import (
	"context"
	"testing"

	"typer/internal/model"
)

func TestStaticProviderReturnsFixedPrompt(t *testing.T) {
	want := model.Prompt{ID: "x", Content: "alpha beta", Mode: model.ModeWords, Source: "seed"}
	p := NewStaticProvider(want)
	for range 3 {
		got, err := p.Next(context.Background(), Constraints{Words: 99, Source: "remote"})
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		if got != want {
			t.Fatalf("got %#v, want %#v", got, want)
		}
	}
}
