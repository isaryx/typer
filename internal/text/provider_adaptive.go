package text

import (
	"context"

	"typer/internal/model"
	"typer/internal/train"
)

type AdaptiveProvider struct {
	filter   *train.WordFilter
	weakKeys []string
	words    int
	prepared bool
	prompt   model.Prompt
	lastNote string
}

func NewAdaptiveProvider(filter *train.WordFilter, weakKeys []string, words int) *AdaptiveProvider {
	if words <= 0 {
		words = 15
	}
	return &AdaptiveProvider{
		filter:   filter,
		weakKeys: append([]string(nil), weakKeys...),
		words:    words,
	}
}

func (p *AdaptiveProvider) Name() string {
	return "train-adaptive"
}

func (p *AdaptiveProvider) prepare() (model.Prompt, string) {
	if p.prepared {
		return p.prompt, p.lastNote
	}
	content, note := train.BuildAdaptivePrompt(p.filter, p.weakKeys, p.words)
	p.lastNote = note
	p.prompt = model.Prompt{
		ID:      "adaptive",
		Content: content,
		Source:  "train-adaptive",
		Mode:    model.ModeTrain,
	}
	p.prepared = true
	return p.prompt, p.lastNote
}

func (p *AdaptiveProvider) Next(_ context.Context, _ Constraints) (model.Prompt, error) {
	got, _ := p.prepare()
	return got, nil
}

// LastNote returns a non-empty warning when the prompt used a fallback word selection.
func (p *AdaptiveProvider) LastNote() string {
	_, note := p.prepare()
	return note
}
