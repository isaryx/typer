package text_test

import (
	"context"
	"strings"
	"testing"

	"typer/internal/model"
	"typer/internal/text"
	"typer/internal/train"
)

func TestTrainProviderNext(t *testing.T) {
	t.Parallel()

	c, err := train.LoadCurriculum()
	if err != nil {
		t.Fatal(err)
	}
	lesson, ok := c.Lesson("1.1")
	if !ok {
		t.Fatal("missing 1.1")
	}
	p := text.NewTrainProvider(train.NewWordFilter([]string{"fad", "jad"}), lesson)
	got, err := p.Next(context.Background(), text.Constraints{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != model.ModeTrain || got.Content == "" {
		t.Fatalf("prompt = %+v", got)
	}
}

func TestTrainProviderCombinesAllLines(t *testing.T) {
	t.Parallel()

	c, err := train.LoadCurriculum()
	if err != nil {
		t.Fatal(err)
	}
	lesson, ok := c.Lesson("1.1")
	if !ok {
		t.Fatal("missing 1.1")
	}
	p := text.NewTrainProvider(train.NewWordFilter(nil), lesson)
	got, err := p.Next(context.Background(), text.Constraints{})
	if err != nil {
		t.Fatal(err)
	}
	lines := lesson.PromptLines()
	if len(lines) < 2 {
		t.Fatalf("expected multiple lines, got %d", len(lines))
	}
	for _, line := range lines {
		for _, word := range strings.Fields(line) {
			if !strings.Contains(got.Content, word) {
				t.Fatalf("content missing word %q from line %q in %q", word, line, got.Content)
			}
		}
	}
	if !strings.Contains(got.Content, "\n") {
		t.Fatalf("expected newline-separated lesson content")
	}
	for _, line := range strings.Split(got.Content, "\n") {
		if strings.TrimSpace(line) == "" {
			t.Fatalf("empty lesson line in %q", got.Content)
		}
	}
}

func TestAdaptiveProviderNext(t *testing.T) {
	t.Parallel()

	filter := train.NewWordFilter([]string{"apple", "apply", "happy", "zip"})
	p := text.NewAdaptiveProvider(filter, []string{"p"}, 5)
	got, err := p.Next(context.Background(), text.Constraints{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Mode != model.ModeTrain || got.Content == "" {
		t.Fatalf("prompt = %+v", got)
	}
}
