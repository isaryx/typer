package train

import (
	"strings"
	"unicode"
)

// LetterKeyCoverage counts a-z occurrences in prompt text (words joined by spaces).
func LetterKeyCoverage(content string) map[string]int {
	return countTargetChars([]rune(strings.Join(strings.Fields(content), " ")))
}

// MissingPlacementLetters returns lowercase a-z absent from static placement segment content.
func MissingPlacementLetters(segments []PlacementSegment) []string {
	seen := map[rune]bool{}
	for _, seg := range segments {
		if seg.Content == "" {
			continue
		}
		for ch := range LetterKeyCoverage(seg.Content) {
			if len(ch) == 1 {
				seen[rune(ch[0])] = true
			}
		}
	}
	var missing []string
	for ch := 'a'; ch <= 'z'; ch++ {
		if !seen[ch] {
			missing = append(missing, string(ch))
		}
	}
	return missing
}

// HasAllLetters reports whether content includes every a-z letter at least once.
func HasAllLetters(content string) bool {
	seen := map[rune]bool{}
	for _, r := range strings.ToLower(content) {
		if unicode.IsLetter(r) {
			seen[r] = true
		}
	}
	for ch := 'a'; ch <= 'z'; ch++ {
		if !seen[ch] {
			return false
		}
	}
	return true
}
