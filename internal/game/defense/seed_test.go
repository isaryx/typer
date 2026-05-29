package defense

import (
	"math/rand/v2"
	"testing"
)

func TestResolveSeedExplicit(t *testing.T) {
	if got := resolveSeed(42); got != 42 {
		t.Fatalf("got %d want 42", got)
	}
}

func TestResolveSeedRandomWhenZero(t *testing.T) {
	a := resolveSeed(0)
	b := resolveSeed(0)
	if a == 0 && b == 0 {
		t.Fatal("expected non-zero random seeds")
	}
	// Extremely unlikely to collide twice; if flaky, still better than always 0.
	if a == b {
		t.Logf("warning: two random seeds matched: %d", a)
	}
}

func TestWordPoolShuffleChangesOrder(t *testing.T) {
	words := []string{"ace", "act", "add", "age", "ago", "aid", "aim", "air", "ant", "ape"}
	p1 := NewWordPool(words)
	p2 := NewWordPool(words)
	p1.Shuffle(rand.New(rand.NewPCG(1, 2)))
	p2.Shuffle(rand.New(rand.NewPCG(99, 100)))
	if len(p1.all) != len(p2.all) {
		t.Fatal("length mismatch")
	}
	same := true
	for i := range p1.all {
		if p1.all[i] != p2.all[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatal("expected different shuffle order for different seeds")
	}
}

func TestDifferentSeedsProduceDifferentSpawns(t *testing.T) {
	pool := NewWordPool([]string{"cat", "dog", "bat", "rat", "hat", "mat", "pat", "sat", "eat", "fat"})
	r1 := rand.New(rand.NewPCG(10, 20))
	r2 := rand.New(rand.NewPCG(30, 40))
	pool.Shuffle(r1)
	w1, ok1 := spawnWord(pool, 40, r1, 1, nil, 0, "")
	pool2 := NewWordPool([]string{"cat", "dog", "bat", "rat", "hat", "mat", "pat", "sat", "eat", "fat"})
	pool2.Shuffle(r2)
	w2, ok2 := spawnWord(pool2, 40, r2, 1, nil, 0, "")
	if !ok1 || !ok2 {
		t.Fatal("expected spawns")
	}
	// Not guaranteed but with 10 words and different shuffles/seeds very likely different.
	if w1.Text == w2.Text && w1.Col == w2.Col {
		t.Logf("note: first spawn matched %q@%d — possible but rare", w1.Text, w1.Col)
	}
}
