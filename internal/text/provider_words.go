package text

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"typer/assets"
	"typer/internal/model"
)

const defaultWordCount = 15

type WordsProvider struct {
	words    []string
	sourceID string
}

func NewWordsProvider(wordsFile string) (*WordsProvider, error) {
	rawWords, sourceID, err := loadCorpusOrBuiltin(wordsFile, assets.Words, "assets/words.txt")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(rawWords, "\n")
	words := make([]string, 0, len(lines))
	for _, l := range lines {
		if w := strings.TrimSpace(l); w != "" {
			words = append(words, w)
		}
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("word list is empty")
	}
	return &WordsProvider{words: words, sourceID: sourceID}, nil
}

func (p *WordsProvider) Name() string {
	return p.sourceID
}

func (p *WordsProvider) AllWords() []string {
	out := make([]string, len(p.words))
	copy(out, p.words)
	return out
}

func (p *WordsProvider) Next(_ context.Context, c Constraints) (model.Prompt, error) {
	n := c.Words
	if n <= 0 {
		n = defaultWordCount
	}
	var b strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(p.words[rand.IntN(len(p.words))])
	}
	return model.Prompt{
		ID:      time.Now().UTC().Format(time.RFC3339Nano),
		Content: b.String(),
		Source:  p.Name(),
		Mode:    model.ModeWords,
	}, nil
}
