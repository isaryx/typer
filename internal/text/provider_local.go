package text

import (
	"context"
	"errors"
	"math/rand/v2"
	"strings"
	"time"

	"typer/assets"
	"typer/internal/model"
)

type LocalProvider struct {
	passages []string
	sourceID string
}

func splitPassageChunks(raw string) []string {
	chunks := strings.Split(raw, "\n\n")
	out := make([]string, 0, len(chunks))
	for _, c := range chunks {
		t := strings.TrimSpace(c)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// NewLocalProvider loads passages from passagesFile when non-empty (blank-line-separated
// blocks, same layout as the embedded bundle); otherwise it uses the built-in corpus.
func NewLocalProvider(passagesFile string) (*LocalProvider, error) {
	raw, sourceID, err := loadCorpusOrBuiltin(passagesFile, assets.Passages, "assets/passages.txt")
	if err != nil {
		return nil, err
	}
	passages := splitPassageChunks(raw)
	if len(passages) == 0 {
		return nil, errors.New("no passages found (use blank-line-separated blocks in the file)")
	}
	return &LocalProvider{passages: passages, sourceID: sourceID}, nil
}

func (p *LocalProvider) Name() string {
	return p.sourceID
}

func (p *LocalProvider) Next(_ context.Context, _ Constraints) (model.Prompt, error) {
	return model.Prompt{
		ID:      time.Now().UTC().Format(time.RFC3339Nano),
		Content: p.passages[rand.IntN(len(p.passages))],
		Source:  p.Name(),
		Mode:    model.ModePassage,
	}, nil
}
