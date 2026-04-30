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
	"typer/internal/version"

	"github.com/oklog/ulid/v2"
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

	_, err := r.Run(context.Background(), model.SessionOptions{Mode: model.ModeWords}, strings.NewReader(""), &bytes.Buffer{}, nil)
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

	wantOpts := model.SessionOptions{Mode: model.ModeWords}

	var out bytes.Buffer
	got, err := r.Run(
		context.Background(),
		wantOpts,
		strings.NewReader("hello "),
		&out,
		nil,
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
	wantHash := model.PromptContentHash(prompt.Content)
	if got.ContentHash != wantHash {
		t.Fatalf("ContentHash = %q, want %q", got.ContentHash, wantHash)
	}
	if got.ElapsedMS <= 0 {
		t.Fatalf("ElapsedMS = %d, want > 0", got.ElapsedMS)
	}
	if !got.EndedAt.After(got.StartedAt) {
		t.Fatalf("expected EndedAt > StartedAt, got %v <= %v", got.EndedAt, got.StartedAt)
	}
	u, err := ulid.ParseStrict(got.ID)
	if err != nil {
		t.Fatalf("ID should be a ULID: %v", err)
	}
	if u.Timestamp().UnixMilli() != got.StartedAt.UTC().UnixMilli() {
		t.Fatalf("ULID embedded time %v != StartedAt (ms) %v", u.Timestamp().UnixMilli(), got.StartedAt.UTC().UnixMilli())
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
	if got.ResultSchema != model.SessionResultSchema {
		t.Fatalf("ResultSchema = %d, want %d", got.ResultSchema, model.SessionResultSchema)
	}
	if got.TyperVersion != version.Version {
		t.Fatalf("TyperVersion = %q, want %q", got.TyperVersion, version.Version)
	}
	if got.Options != model.OptionsSnapshot(wantOpts) {
		t.Fatalf("Options = %#v, want %#v", got.Options, model.OptionsSnapshot(wantOpts))
	}
	if got.TotalKeystrokes <= 0 {
		t.Fatalf("TotalKeystrokes = %d, want > 0", got.TotalKeystrokes)
	}
	if got.CorrectKeystrokes <= 0 {
		t.Fatalf("CorrectKeystrokes = %d, want > 0", got.CorrectKeystrokes)
	}
	if len(got.WPMSamples) < 1 {
		t.Fatalf("WPMSamples = %v, want at least one sample after completing a word", got.WPMSamples)
	}
	if len(got.TypingTrace) < 1 {
		t.Fatalf("expected typing trace to be stored, got %d events", len(got.TypingTrace))
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
		nil,
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
