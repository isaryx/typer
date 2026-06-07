package defense

import (
	"fmt"

	"typer/internal/text"
)

// LoadWordPool loads and filters a word list for defense (lengths MinWordLen–MaxWordLen).
func LoadWordPool(wordsFile string) (WordPool, error) {
	corpus, err := text.LoadWordCorpus(wordsFile)
	if err != nil {
		return WordPool{}, err
	}
	pool := NewWordPool(corpus.Words)
	if pool.Len() == 0 {
		return WordPool{}, fmt.Errorf("no words between %d and %d characters in word list", MinWordLen, MaxWordLen)
	}
	return pool, nil
}
