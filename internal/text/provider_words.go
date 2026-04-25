package text

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"typer/assets"
	"typer/internal/model"
)

const defaultWordCount = 15

type WordsProvider struct {
	rng      *rand.Rand
	words    []string
	sourceID string
}

func NewWordsProvider(wordsFile string) (*WordsProvider, error) {
	rawWords := assets.Words
	sourceID := "builtin:assets/words.txt"
	if strings.TrimSpace(wordsFile) != "" {
		path, err := filepath.Abs(wordsFile)
		if err != nil {
			return nil, fmt.Errorf("resolve words file path: %w", err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read words file %q: %w", path, err)
		}
		rawWords = string(data)
		sourceID = "file:" + path
	}

	lines := strings.Split(rawWords, "\n")
	words := make([]string, 0, len(lines))
	for _, l := range lines {
		w := strings.TrimSpace(l)
		if w != "" {
			words = append(words, w)
		}
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("word list is empty")
	}

	return &WordsProvider{
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
		words:    words,
		sourceID: sourceID,
	}, nil
}

func (p *WordsProvider) Name() string {
	return p.sourceID
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
		b.WriteString(p.words[p.rng.Intn(len(p.words))])
	}
	return model.Prompt{
		ID:      time.Now().UTC().Format(time.RFC3339Nano),
		Content: b.String(),
		Source:  p.Name(),
		Mode:    "words",
	}, nil
}
