package text

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseWordLines(t *testing.T) {
	words, err := parseWordLines("alpha\n\n beta \ngamma\n")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if strings.Join(words, ",") != "alpha,beta,gamma" {
		t.Fatalf("got %v", words)
	}
}

func TestParseWordLines_Empty(t *testing.T) {
	_, err := parseWordLines("\n\n  \n")
	if err == nil {
		t.Fatal("expected error for empty list")
	}
}

func TestLoadWordCorpus_Builtin(t *testing.T) {
	corpus, err := LoadWordCorpus("")
	if err != nil {
		t.Fatalf("load builtin: %v", err)
	}
	if !strings.HasPrefix(corpus.SourceID, "builtin:") {
		t.Fatalf("source = %q, want builtin prefix", corpus.SourceID)
	}
	if len(corpus.Words) == 0 {
		t.Fatal("expected non-empty builtin word list")
	}
}

func TestLoadWordCorpus_CustomFile(t *testing.T) {
	tmpDir := t.TempDir()
	customPath := filepath.Join(tmpDir, "words.txt")
	if err := os.WriteFile(customPath, []byte("alpha\nbeta\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	corpus, err := LoadWordCorpus(customPath)
	if err != nil {
		t.Fatalf("load custom: %v", err)
	}
	if !strings.HasPrefix(corpus.SourceID, "file:") {
		t.Fatalf("source = %q, want file prefix", corpus.SourceID)
	}
	if strings.Join(corpus.Words, ",") != "alpha,beta" {
		t.Fatalf("words = %v", corpus.Words)
	}
}

func TestLoadWordCorpus_BuiltinCached(t *testing.T) {
	c1, err := LoadWordCorpus("")
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	c2, err := LoadWordCorpus("")
	if err != nil {
		t.Fatalf("second load: %v", err)
	}
	if &c1.Words[0] != &c2.Words[0] {
		t.Fatal("expected builtin corpus to share the same word slice")
	}
}

func TestPracticeSlotCounts(t *testing.T) {
	counts := practiceSlotCounts(15)
	sum := counts[bucketShort] + counts[bucketCore] + counts[bucketMid] + counts[bucketLong]
	if sum != 15 {
		t.Fatalf("sum = %d, want 15", sum)
	}
	if counts[bucketShort] != 3 || counts[bucketCore] != 6 || counts[bucketMid] != 4 || counts[bucketLong] != 2 {
		t.Fatalf("unexpected counts for n=15: %v", counts)
	}

	counts10 := practiceSlotCounts(10)
	sum10 := counts10[bucketShort] + counts10[bucketCore] + counts10[bucketMid] + counts10[bucketLong]
	if sum10 != 10 {
		t.Fatalf("sum = %d, want 10", sum10)
	}
}
