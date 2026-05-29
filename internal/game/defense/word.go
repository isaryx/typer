package defense

import (
	"math/rand/v2"
	"unicode/utf8"
)

// Word is a falling target on the playfield.
type Word struct {
	ID   int
	Text string
	Col  int
	Row  float64
}

// WordPool indexes defense words by length for progressive spawn selection.
type WordPool struct {
	byLength map[int][]string
	all      []string
}

// NewWordPool builds a length-indexed pool from raw word strings.
func NewWordPool(src []string) WordPool {
	filtered := FilterWordPool(src)
	byLength := make(map[int][]string)
	for _, w := range filtered {
		n := utf8.RuneCountInString(w)
		byLength[n] = append(byLength[n], w)
	}
	return WordPool{byLength: byLength, all: filtered}
}

// Shuffle randomizes word order within each length bucket and in all.
func (p *WordPool) Shuffle(rng *rand.Rand) {
	shuffleStrings(p.all, rng)
	for n := range p.byLength {
		shuffleStrings(p.byLength[n], rng)
	}
}

func shuffleStrings(s []string, rng *rand.Rand) {
	for i := len(s) - 1; i > 0; i-- {
		j := rng.IntN(i + 1)
		s[i], s[j] = s[j], s[i]
	}
}

// MaxWordLenForScore returns the longest word length allowed after score words destroyed.
func MaxWordLenForScore(score int) int {
	if score < 0 {
		score = 0
	}
	extra := score / wordsPerLenTier
	maxLen := startMaxWordLen + extra
	if maxLen > MaxWordLen {
		return MaxWordLen
	}
	if maxLen < startMaxWordLen {
		return startMaxWordLen
	}
	return maxLen
}

func (p WordPool) eligibleWords(minLen, maxLen int) []string {
	if len(p.all) == 0 {
		return nil
	}
	out := make([]string, 0)
	for n := minLen; n <= maxLen; n++ {
		out = append(out, p.byLength[n]...)
	}
	if len(out) == 0 {
		return p.all
	}
	return out
}

// FilterWordPool keeps words suitable for defense (length bounds, non-empty).
func FilterWordPool(src []string) []string {
	out := make([]string, 0, len(src))
	for _, w := range src {
		n := utf8.RuneCountInString(w)
		if n >= MinWordLen && n <= MaxWordLen {
			out = append(out, w)
		}
	}
	return out
}

const spawnColumnGap = 1

// lockBracketCols is horizontal space reserved for "[" and "]" when a word is locked.
const lockBracketCols = 2

func maxSpawnCol(innerWidth, wlen int) int {
	return innerWidth - wlen - lockBracketCols
}

func clampWordCol(col, wlen, innerWidth int) int {
	maxCol := maxSpawnCol(innerWidth, wlen)
	if maxCol < 0 {
		return 0
	}
	if col > maxCol {
		return maxCol
	}
	if col < 0 {
		return 0
	}
	return col
}

func spawnWord(pool WordPool, innerWidth int, rng *rand.Rand, nextID int, existing []Word, score int, lastSpawned string) (Word, bool) {
	if len(pool.all) == 0 || innerWidth < MinWordLen+lockBracketCols {
		return Word{}, false
	}
	maxLen := MaxWordLenForScore(score)
	eligible := pool.eligibleWords(MinWordLen, maxLen)
	if len(eligible) == 0 {
		return Word{}, false
	}
	for attempt := 0; attempt < 24; attempt++ {
		text := eligible[rng.IntN(len(eligible))]
		if text == lastSpawned && len(eligible) > 1 {
			continue
		}
		wlen := utf8.RuneCountInString(text)
		if wlen > innerWidth {
			continue
		}
		maxCol := maxSpawnCol(innerWidth, wlen)
		if maxCol < 0 {
			continue
		}
		col := 0
		if maxCol > 0 {
			col = rng.IntN(maxCol + 1)
		}
		if overlapsExisting(col, wlen, 0, existing) {
			continue
		}
		return Word{ID: nextID, Text: text, Col: col, Row: 0}, true
	}
	return Word{}, false
}

func overlapsExisting(col, width int, row int, existing []Word) bool {
	for _, w := range existing {
		if int(w.Row) != row {
			continue
		}
		wlen := utf8.RuneCountInString(w.Text)
		wEnd := w.Col + wlen + lockBracketCols
		end := col + width + lockBracketCols
		if col <= wEnd+spawnColumnGap && w.Col <= end+spawnColumnGap {
			return true
		}
	}
	return false
}

func wordByID(words []Word, id int) (Word, int, bool) {
	for i, w := range words {
		if w.ID == id {
			return w, i, true
		}
	}
	return Word{}, -1, false
}

// selectLockCandidate picks the lowest (most dangerous) word matching the next keystroke.
func selectLockCandidate(words []Word, r rune, typed string) (int, bool) {
	bestID := -1
	bestRow := -1.0
	for _, w := range words {
		target := []rune(w.Text)
		cur := []rune(typed)
		if len(cur) > len(target) {
			continue
		}
		if !runesPrefix(cur, target) {
			continue
		}
		if len(cur) >= len(target) {
			continue
		}
		if target[len(cur)] != r {
			continue
		}
		if w.Row > bestRow {
			bestRow = w.Row
			bestID = w.ID
		}
	}
	return bestID, bestID >= 0
}

func runesPrefix(prefix, full []rune) bool {
	if len(prefix) > len(full) {
		return false
	}
	for i := range prefix {
		if prefix[i] != full[i] {
			return false
		}
	}
	return true
}

func strictAppendTyped(word, typed string, r rune) (string, bool) {
	target := []rune(word)
	cur := []rune(typed)
	if len(cur) >= len(target) {
		return typed, false
	}
	if target[len(cur)] != r {
		return typed, false
	}
	return typed + string(r), true
}
