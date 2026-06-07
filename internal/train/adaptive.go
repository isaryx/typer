package train

import (
	"math/rand/v2"
	"strings"
)

// WordFilter indexes words by character set for lesson and adaptive generation.
type WordFilter struct {
	all []string
}

func NewWordFilter(words []string) *WordFilter {
	return &WordFilter{all: append([]string(nil), words...)}
}

// WordsContaining returns up to n words that contain all required keys (case-insensitive).
// The bool is false when no matching words were found and a random sample was returned instead.
func (f *WordFilter) WordsContaining(required []string, n int) ([]string, bool) {
	if n <= 0 || len(f.all) == 0 {
		return nil, true
	}
	req := normalizeKeys(required)
	var matches []string
	for _, w := range f.all {
		if wordHasKeys(w, req) {
			matches = append(matches, w)
		}
	}
	if len(matches) == 0 {
		return f.sample(f.all, n), false
	}
	return f.sample(matches, n), true
}

// WordsForWeakKeys picks words emphasizing weak keys.
func (f *WordFilter) WordsForWeakKeys(weakKeys []string, n int) ([]string, bool) {
	if len(weakKeys) == 0 {
		return f.sample(f.all, n), false
	}
	return f.WordsContaining(weakKeys[:min(len(weakKeys), 3)], n)
}

func (f *WordFilter) sample(src []string, n int) []string {
	if len(src) == 0 {
		return nil
	}
	if len(src) <= n {
		out := append([]string(nil), src...)
		rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
		return out
	}
	perm := rand.Perm(len(src))[:n]
	out := make([]string, n)
	for i, idx := range perm {
		out[i] = src[idx]
	}
	return out
}

func normalizeKeys(keys []string) []rune {
	var out []rune
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		r := []rune(strings.ToLower(k))
		if len(r) == 1 {
			out = append(out, r[0])
		}
	}
	return out
}

func wordHasKeys(word string, keys []rune) bool {
	if len(keys) == 0 {
		return true
	}
	lower := strings.ToLower(word)
	for _, k := range keys {
		if !strings.ContainsRune(lower, k) {
			return false
		}
	}
	return true
}

// BuildAdaptivePrompt builds a practice line from weak keys using the word filter.
// note is non-empty when words could not be matched to weak keys and a fallback was used.
func BuildAdaptivePrompt(filter *WordFilter, weakKeys []string, wordCount int) (content, note string) {
	if wordCount <= 0 {
		wordCount = 15
	}
	if len(filter.all) == 0 {
		return "the quick brown fox jumps over the lazy dog", "word list empty — using default phrase"
	}
	words, matched := filter.WordsForWeakKeys(weakKeys, wordCount)
	if len(words) == 0 {
		return "the quick brown fox jumps over the lazy dog", "word list empty — using default phrase"
	}
	if len(weakKeys) == 0 {
		return strings.Join(words, " "), "no weak keys tracked — random words"
	}
	if !matched {
		return strings.Join(words, " "), "no words matched weak keys — random words"
	}
	return strings.Join(words, " "), ""
}
