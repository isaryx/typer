package text

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestNewWordsProvider_DefaultBuiltinList(t *testing.T) {
	p, err := NewWordsProvider("")
	if err != nil {
		t.Fatalf("new words provider: %v", err)
	}
	if !strings.HasPrefix(p.Name(), "builtin:") {
		t.Fatalf("expected builtin source name, got %q", p.Name())
	}

	prompt, err := p.Next(context.Background(), Constraints{Words: 5})
	if err != nil {
		t.Fatalf("next prompt: %v", err)
	}
	words := strings.Fields(prompt.Content)
	if len(words) != 5 {
		t.Fatalf("expected 5 words in prompt, got %q", prompt.Content)
	}
}

func TestNewWordsProvider_CustomFile(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "my-words.txt")
	if err := os.WriteFile(customPath, []byte("alpha\nbeta\ngamma\n"), 0o644); err != nil {
		t.Fatalf("write custom words file: %v", err)
	}

	p, err := NewWordsProvider(customPath)
	if err != nil {
		t.Fatalf("new words provider with custom file: %v", err)
	}
	if !strings.HasPrefix(p.Name(), "file:") {
		t.Fatalf("expected custom file source name, got %q", p.Name())
	}

	prompt, err := p.Next(context.Background(), Constraints{Words: 20})
	if err != nil {
		t.Fatalf("next prompt: %v", err)
	}
	for _, word := range strings.Fields(prompt.Content) {
		switch word {
		case "alpha", "beta", "gamma":
		default:
			t.Fatalf("unexpected word from custom list: %q", word)
		}
	}
}

func TestWordsProvider_StratifiedLengthMix(t *testing.T) {
	p, err := NewWordsProvider("")
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}

	want := practiceSlotCounts(15)
	for range 20 {
		prompt, err := p.Next(context.Background(), Constraints{Words: 15})
		if err != nil {
			t.Fatalf("next: %v", err)
		}
		words := strings.Fields(prompt.Content)
		if len(words) != 15 {
			t.Fatalf("got %d words, want 15", len(words))
		}
		got := countPracticeBuckets(words)
		for b := range practiceBucketCount {
			if diff(got[b], want[b]) > 1 {
				t.Fatalf("bucket %d: got %d want %d (±1); full counts %v want %v", b, got[b], want[b], got, want)
			}
		}
		nonEmpty := 0
		for _, n := range got {
			if n > 0 {
				nonEmpty++
			}
		}
		if nonEmpty < 3 {
			t.Fatalf("expected words from at least 3 length buckets, got %v", got)
		}
	}
}

func TestWordsProvider_StratifiedScaling(t *testing.T) {
	p, err := NewWordsProvider("")
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	want := practiceSlotCounts(5)
	prompt, err := p.Next(context.Background(), Constraints{Words: 5})
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	words := strings.Fields(prompt.Content)
	if len(words) != 5 {
		t.Fatalf("got %d words, want 5", len(words))
	}
	got := countPracticeBuckets(words)
	sum := 0
	for b := range practiceBucketCount {
		sum += got[b]
		if diff(got[b], want[b]) > 1 {
			t.Fatalf("bucket %d: got %d want %d (±1)", b, got[b], want[b])
		}
	}
	if sum != 5 {
		t.Fatalf("bucket sum = %d, want 5", sum)
	}
}

func TestWordsProvider_RedistributeEmptyBuckets(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "core-only.txt")
	lines := []string{"apple", "banana", "cherry", "donut", "elder", "floral", "grape", "honey"}
	if err := os.WriteFile(customPath, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	p, err := NewWordsProvider(customPath)
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}

	counts := p.effectiveSlotCounts(15)
	if counts[bucketShort] != 0 || counts[bucketMid] != 0 || counts[bucketLong] != 0 {
		t.Fatalf("expected empty non-core slot targets after redistribution, got %v", counts)
	}
	if counts[bucketCore] != 15 {
		t.Fatalf("core slots = %d, want 15", counts[bucketCore])
	}

	prompt, err := p.Next(context.Background(), Constraints{Words: 15})
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	for _, w := range strings.Fields(prompt.Content) {
		b, ok := practiceBucketForWord(w)
		if !ok || b != bucketCore {
			t.Fatalf("expected core-bucket word, got %q", w)
		}
	}
}

func TestEffectiveSlotCounts_RedistributesToAdjacent(t *testing.T) {
	p := &WordsProvider{
		buckets: [practiceBucketCount]bucketState{
			bucketShort: {wordIdx: []int{0}},
			bucketCore:  {},
			bucketMid:   {wordIdx: []int{1}},
			bucketLong:  {},
		},
	}
	base := practiceSlotCounts(10)
	got := p.effectiveSlotCounts(10)

	if got[bucketCore] != 0 || got[bucketLong] != 0 {
		t.Fatalf("empty buckets should have 0 slots after redistribution, got %v", got)
	}
	wantShort := base[bucketShort] + base[bucketCore]
	wantMid := base[bucketMid] + base[bucketLong]
	if got[bucketShort] != wantShort || got[bucketMid] != wantMid {
		t.Fatalf("got %v, want short=%d mid=%d", got, wantShort, wantMid)
	}
}

func countPracticeBuckets(words []string) [practiceBucketCount]int {
	var counts [practiceBucketCount]int
	for _, w := range words {
		b, ok := practiceBucketForWord(w)
		if !ok {
			continue
		}
		counts[b]++
	}
	return counts
}

func diff(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}

func TestWordsProvider_NextNoDuplicatesWithinDeck(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "words.txt")
	var b strings.Builder
	for i := range 20 {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("w" + string(rune('a'+i)) + "oo")
	}
	if err := os.WriteFile(customPath, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write words file: %v", err)
	}

	p, err := NewWordsProvider(customPath)
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	prompt, err := p.Next(context.Background(), Constraints{Words: 15})
	if err != nil {
		t.Fatalf("next: %v", err)
	}
	words := strings.Fields(prompt.Content)
	if len(words) != 15 {
		t.Fatalf("got %d words, want 15", len(words))
	}
	if hasDuplicate(words) {
		t.Fatalf("expected no duplicates within deck, got %q", prompt.Content)
	}
	for _, w := range words {
		if utf8.RuneCountInString(w) < 3 || utf8.RuneCountInString(w) > 12 {
			t.Fatalf("unexpected word length: %q", w)
		}
	}
}

func hasDuplicate(words []string) bool {
	seen := make(map[string]struct{}, len(words))
	for _, w := range words {
		if _, ok := seen[w]; ok {
			return true
		}
		seen[w] = struct{}{}
	}
	return false
}
