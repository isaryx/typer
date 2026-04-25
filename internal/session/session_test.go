package session

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"typer/internal/model"
	"typer/internal/text"
)

type fakeProvider struct {
	prompt model.Prompt
	err    error
}

func (f *fakeProvider) Next(context.Context, text.Constraints) (model.Prompt, error) {
	return f.prompt, f.err
}

func (f *fakeProvider) Name() string { return "fake" }

func TestRunnerPropagatesProviderError(t *testing.T) {
	wantErr := errors.New("provider offline")
	r := NewRunner(&fakeProvider{err: wantErr})

	_, err := r.Run(context.Background(), model.SessionOptions{Mode: model.ModeWords}, strings.NewReader(""), &bytes.Buffer{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected provider error, got %v", err)
	}
}

func TestRunnerRunSuccessBuildsSessionResult(t *testing.T) {
	prompt := model.Prompt{
		Content: "hello",
		Mode:    model.ModeWords,
		Source:  "seed",
	}
	r := NewRunner(&fakeProvider{prompt: prompt})

	base := time.Unix(1_000, 0)
	calls := 0
	r.Now = func() time.Time {
		calls++
		return base.Add(time.Duration(calls) * time.Second)
	}

	var out bytes.Buffer
	got, err := r.Run(
		context.Background(),
		model.SessionOptions{Mode: model.ModeWords},
		strings.NewReader("hello "),
		&out,
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if got.Aborted {
		t.Fatal("expected completed run, got aborted")
	}
	if got.Prompt.Content != "hello" || got.Prompt.Mode != model.ModeWords || got.Prompt.Source != "seed" {
		t.Fatalf("unexpected prompt in result: %#v", got.Prompt)
	}
	if got.TypedText != "hello" {
		t.Fatalf("TypedText = %q, want hello", got.TypedText)
	}
	if got.ElapsedMS <= 0 {
		t.Fatalf("ElapsedMS = %d, want > 0", got.ElapsedMS)
	}
	if !got.EndedAt.After(got.StartedAt) {
		t.Fatalf("expected EndedAt > StartedAt, got %v <= %v", got.EndedAt, got.StartedAt)
	}
	if got.ID != got.StartedAt.UTC().Format(time.RFC3339Nano) {
		t.Fatalf("ID = %q, want %q", got.ID, got.StartedAt.UTC().Format(time.RFC3339Nano))
	}
	if got.Metrics.Accuracy != 100 {
		t.Fatalf("Accuracy = %.2f, want 100", got.Metrics.Accuracy)
	}
	if got.Metrics.Errors != 0 {
		t.Fatalf("Errors = %d, want 0", got.Metrics.Errors)
	}
	if got.Metrics.GrossWPM <= 0 || got.Metrics.NetWPM <= 0 || got.Metrics.AdjustedWPM <= 0 {
		t.Fatalf("expected positive WPM metrics, got %#v", got.Metrics)
	}
	if got.Metrics.Consistency != 100 {
		t.Fatalf("expected 100 consistency with <=1 sample, got %.2f", got.Metrics.Consistency)
	}
	clearScreen := "\x1b[2J\x1b[H"
	if !strings.Contains(out.String(), clearScreen) {
		t.Fatalf("expected clear-screen sequence in output, got %q", out.String())
	}
}

func TestRunnerRunPassesConstraints(t *testing.T) {
	wantWords := 12
	wantSource := "remote"
	fp := &fakeProvider{
		prompt: model.Prompt{Content: "go", Mode: model.ModeQuote, Source: "remote"},
	}
	var gotConstraints text.Constraints
	provider := &capturingProvider{
		inner: fp,
		onNext: func(c text.Constraints) {
			gotConstraints = c
		},
	}
	r := NewRunner(provider)
	r.Now = func() time.Time { return time.Unix(2_000, 0) }

	_, err := r.Run(
		context.Background(),
		model.SessionOptions{Words: wantWords, Source: wantSource},
		strings.NewReader("\x1b"),
		&bytes.Buffer{},
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if gotConstraints.Words != wantWords || gotConstraints.Source != wantSource {
		t.Fatalf("constraints = %#v, want words=%d source=%q", gotConstraints, wantWords, wantSource)
	}
}

type capturingProvider struct {
	inner  text.Provider
	onNext func(text.Constraints)
}

func (c *capturingProvider) Next(ctx context.Context, constraints text.Constraints) (model.Prompt, error) {
	if c.onNext != nil {
		c.onNext(constraints)
	}
	return c.inner.Next(ctx, constraints)
}

func (c *capturingProvider) Name() string {
	if c.inner == nil {
		return "capturing"
	}
	return fmt.Sprintf("capture:%s", c.inner.Name())
}
