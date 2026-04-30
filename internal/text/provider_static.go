package text

import (
	"context"

	"typer/internal/model"
)

// StaticProvider always returns the same prompt (used for replay).
type StaticProvider struct {
	prompt model.Prompt
}

func NewStaticProvider(p model.Prompt) *StaticProvider {
	return &StaticProvider{prompt: p}
}

func (p *StaticProvider) Name() string { return "static" }

func (p *StaticProvider) Next(context.Context, Constraints) (model.Prompt, error) {
	return p.prompt, nil
}
