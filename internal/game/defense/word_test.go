package defense

import (
	"math/rand/v2"
	"testing"
	"time"
)

func TestSpawnWordAvoidsSameRowOverlap(t *testing.T) {
	existing := []Word{{ID: 1, Text: "hello", Col: 5, Row: 0}}
	pool := NewWordPool([]string{"world", "help", "tests", "cat", "dog"})
	rng := rand.New(rand.NewPCG(3, 4))
	for attempt := 0; attempt < 40; attempt++ {
		w, ok := spawnWord(pool, 40, rng, 2, existing, 50, "")
		if !ok {
			continue
		}
		end := lockedSpanEnd(w.Col, len(w.Text))
		existEnd := lockedSpanEnd(existing[0].Col, len(existing[0].Text))
		if lockedSpanStart(w.Col) <= existEnd+spawnColumnGap && lockedSpanStart(existing[0].Col) <= end+spawnColumnGap {
			t.Fatalf("spawned overlapping word %q at col %d", w.Text, w.Col)
		}
	}
}

func TestMaxWordLenForScore(t *testing.T) {
	if got := MaxWordLenForScore(0); got != 4 {
		t.Fatalf("score 0: got %d want 4", got)
	}
	if got := MaxWordLenForScore(9); got != 4 {
		t.Fatalf("score 9: got %d want 4", got)
	}
	if got := MaxWordLenForScore(10); got != 5 {
		t.Fatalf("score 10: got %d want 5", got)
	}
	if got := MaxWordLenForScore(100); got != MaxWordLen {
		t.Fatalf("score 100: got %d want %d", got, MaxWordLen)
	}
}

func TestSpawnWordStartsWithShortWords(t *testing.T) {
	pool := NewWordPool([]string{"go", "cat", "code", "terminal", "programming"})
	rng := rand.New(rand.NewPCG(9, 10))
	for i := 0; i < 50; i++ {
		w, ok := spawnWord(pool, 40, rng, i+1, nil, 0, "")
		if !ok {
			t.Fatal("expected spawn")
		}
		n := len([]rune(w.Text))
		if n < MinWordLen || n > 4 {
			t.Fatalf("early spawn should be 3-4 chars, got %q (len %d)", w.Text, n)
		}
	}
}

func TestSpawnWordAddsLongerWordsAfterCompletions(t *testing.T) {
	pool := NewWordPool([]string{"cat", "code", "hello", "world", "terminal"})
	rng := rand.New(rand.NewPCG(11, 12))
	sawLong := false
	for i := 0; i < 80; i++ {
		w, ok := spawnWord(pool, 40, rng, i+1, nil, 20, "")
		if !ok {
			t.Fatal("expected spawn")
		}
		if len([]rune(w.Text)) > 4 {
			sawLong = true
			break
		}
	}
	if !sawLong {
		t.Fatal("expected longer words after 20 words completed")
	}
}

func TestSpawnAvoidsImmediateRepeat(t *testing.T) {
	pool := NewWordPool([]string{"cat", "dog", "bat", "rat", "hat"})
	rng := rand.New(rand.NewPCG(5, 6))
	last := "cat"
	repeated := 0
	for i := 0; i < 30; i++ {
		w, ok := spawnWord(pool, 40, rng, i+1, nil, 0, last)
		if !ok {
			continue
		}
		if w.Text == last {
			repeated++
		}
		last = w.Text
	}
	if repeated > 2 {
		t.Fatalf("too many immediate repeats: %d", repeated)
	}
}

func TestDisplayScore(t *testing.T) {
	r := Result{Score: 10, Elapsed: 15 * time.Second}
	if got := r.DisplayScore(); got != 25 {
		t.Fatalf("got %d want 25", got)
	}
}
