package text

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"typer/assets"
	"typer/internal/model"
)

type LocalProvider struct {
	rng      *rand.Rand
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
	raw := assets.Passages
	sourceID := "builtin:assets/passages.txt"
	if strings.TrimSpace(passagesFile) != "" {
		path, err := filepath.Abs(passagesFile)
		if err != nil {
			return nil, fmt.Errorf("resolve passages file path: %w", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read passages file %q: %w", path, err)
		}
		raw = string(data)
		sourceID = "file:" + path
	}
	passages := splitPassageChunks(raw)
	if len(passages) == 0 {
		return nil, errors.New("no passages found (use blank-line-separated blocks in the file)")
	}
	return &LocalProvider{
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
		passages: passages,
		sourceID: sourceID,
	}, nil
}

func (p *LocalProvider) Name() string {
	return p.sourceID
}

func (p *LocalProvider) Next(_ context.Context, _ Constraints) (model.Prompt, error) {
	idx := p.rng.Intn(len(p.passages))
	return model.Prompt{
		ID:      time.Now().UTC().Format(time.RFC3339Nano),
		Content: p.passages[idx],
		Source:  p.Name(),
		Mode:    "passage",
	}, nil
}
