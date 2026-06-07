package text

import (
	"fmt"
	"strings"
	"sync"

	"typer/assets"
)

// WordCorpus is a parsed newline-separated word list. Words must not be mutated.
type WordCorpus struct {
	Words    []string
	SourceID string
}

var (
	builtinCorpusOnce sync.Once
	builtinCorpus     WordCorpus
	builtinCorpusErr  error
)

// LoadWordCorpus loads words from path. An empty path uses the embedded assets/words.txt
// (parsed once per process). A non-empty path reads that file via loadCorpusOrBuiltin.
func LoadWordCorpus(path string) (WordCorpus, error) {
	if strings.TrimSpace(path) == "" {
		builtinCorpusOnce.Do(func() {
			words, err := parseWordLines(assets.Words)
			if err != nil {
				builtinCorpusErr = err
				return
			}
			builtinCorpus = WordCorpus{
				Words:    words,
				SourceID: "builtin:assets/words.txt",
			}
		})
		return builtinCorpus, builtinCorpusErr
	}

	rawWords, sourceID, err := loadCorpusOrBuiltin(path, "", "")
	if err != nil {
		return WordCorpus{}, err
	}
	words, err := parseWordLines(rawWords)
	if err != nil {
		return WordCorpus{}, err
	}
	return WordCorpus{Words: words, SourceID: sourceID}, nil
}

func parseWordLines(raw string) ([]string, error) {
	words := make([]string, 0, strings.Count(raw, "\n")+1)
	for len(raw) > 0 {
		i := strings.IndexByte(raw, '\n')
		if i < 0 {
			if w := strings.TrimSpace(raw); w != "" {
				words = append(words, w)
			}
			break
		}
		if w := strings.TrimSpace(raw[:i]); w != "" {
			words = append(words, w)
		}
		raw = raw[i+1:]
	}
	if len(words) == 0 {
		return nil, fmt.Errorf("word list is empty")
	}
	return words, nil
}
