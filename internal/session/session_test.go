package session

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

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

func TestRunner_PropagatesProviderError(t *testing.T) {
	wantErr := errors.New("provider offline")
	r := NewRunner(&fakeProvider{err: wantErr})

	_, err := r.Run(context.Background(), model.SessionOptions{Mode: model.ModeWords}, strings.NewReader(""), &bytes.Buffer{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected provider error, got %v", err)
	}
}
