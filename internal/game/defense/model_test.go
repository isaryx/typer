package defense

import (
	"math/rand/v2"
	"testing"
	"time"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

func testModel(t *testing.T) *defenseModel {
	t.Helper()
	m := newDefenseModel(NewWordPool([]string{"code", "cold", "cat"}), DefaultConfig(), nil, 42)
	m.width = 80
	m.height = 24
	m.startedAt = time.Now()
	m.lastTick = m.startedAt
	return m
}

func TestWordHitsBottomDecrementsLives(t *testing.T) {
	m := testModel(t)
	m.lives = 6
	m.words = []Word{{ID: 1, Text: "code", Col: 0, Row: float64(ShieldRow()) - 0.05}}
	now := m.startedAt.Add(200 * time.Millisecond)
	m.lastTick = m.startedAt
	m.applyTick(now)
	if m.lives != 5 {
		t.Fatalf("lives=%d want 5", m.lives)
	}
	if len(m.words) != 0 {
		t.Fatalf("word should be removed")
	}
}

func TestMaxOneLifeLostPerTick(t *testing.T) {
	m := testModel(t)
	m.lives = 3
	shield := float64(ShieldRow())
	m.words = []Word{
		{ID: 1, Text: "one", Col: 0, Row: shield},
		{ID: 2, Text: "two", Col: 5, Row: shield},
		{ID: 3, Text: "six", Col: 10, Row: shield},
	}
	now := m.startedAt.Add(TickInterval)
	m.applyTick(now)
	if m.lives != 2 {
		t.Fatalf("lives=%d want 2 (only one life lost per tick)", m.lives)
	}
	if len(m.words) != 0 {
		t.Fatalf("all shield words should be removed")
	}
}

func TestSpawnWaitNotConsumedOnFailure(t *testing.T) {
	m := testModel(t)
	m.pool = WordPool{}
	m.spawnWait = DefaultSpawnSeconds + 0.5
	before := m.spawnWait
	now := m.startedAt.Add(time.Second)
	m.lastTick = m.startedAt
	m.applyTick(now)
	if m.spawnWait < before {
		t.Fatalf("spawnWait should not decrease when spawn fails: before=%g after=%g", before, m.spawnWait)
	}
}

func TestStrictWrongKeyRejected(t *testing.T) {
	m := testModel(t)
	m.words = []Word{{ID: 1, Text: "code", Col: 0, Row: 1}}
	m.handleRune('c')
	if m.lockID != 1 || m.typed != "c" {
		t.Fatalf("lock=%d typed=%q", m.lockID, m.typed)
	}
	m.handleRune('x')
	if m.typed != "c" {
		t.Fatalf("typed should stay c, got %q", m.typed)
	}
}

func TestUppercaseNormalized(t *testing.T) {
	m := testModel(t)
	m.words = []Word{{ID: 1, Text: "code", Col: 0, Row: 1}}
	m.handleRune('C')
	if m.lockID != 1 || m.typed != "c" {
		t.Fatalf("uppercase should lock, got lock=%d typed=%q", m.lockID, m.typed)
	}
}

func TestCompleteWordIncrementsScore(t *testing.T) {
	m := testModel(t)
	m.words = []Word{{ID: 1, Text: "cat", Col: 0, Row: 1}}
	m.handleRune('c')
	m.handleRune('a')
	m.handleRune('t')
	if m.score != 1 || len(m.words) != 0 {
		t.Fatalf("score=%d words=%d", m.score, len(m.words))
	}
}

func TestGameOverAtZeroLives(t *testing.T) {
	m := testModel(t)
	m.lives = 1
	m.words = []Word{{ID: 1, Text: "code", Col: 0, Row: float64(ShieldRow())}}
	now := m.startedAt.Add(TickInterval)
	m.applyTick(now)
	if !m.over || m.lives != 0 {
		t.Fatalf("over=%v lives=%d", m.over, m.lives)
	}
}

func TestEscClearsLock(t *testing.T) {
	m := testModel(t)
	m.words = []Word{{ID: 1, Text: "code", Col: 0, Row: 1}}
	m.handleRune('c')
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if m.lockID != 0 || m.typed != "" {
		t.Fatalf("lock should clear")
	}
}

func TestTabClearsLock(t *testing.T) {
	m := testModel(t)
	m.words = []Word{{ID: 1, Text: "code", Col: 0, Row: 1}}
	m.handleRune('c')
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.lockID != 0 || m.typed != "" {
		t.Fatalf("tab should clear lock")
	}
}

func TestBackspaceIgnored(t *testing.T) {
	m := testModel(t)
	m.words = []Word{{ID: 1, Text: "code", Col: 0, Row: 1}}
	m.handleRune('c')
	m.handleRune('o')
	_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if m.typed != "co" {
		t.Fatalf("backspace should not change typed buffer, got %q", m.typed)
	}
}

func TestPasteIgnored(t *testing.T) {
	m := testModel(t)
	m.words = []Word{{ID: 1, Text: "code", Col: 0, Row: 1}}
	_, _ = m.Update(tea.PasteMsg{Content: "code"})
	if m.lockID != 0 || m.score != 0 {
		t.Fatalf("paste should be ignored")
	}
}

func TestClampWordsToLeftEdge(t *testing.T) {
	m := testModel(t)
	m.width = 40
	m.words = []Word{{ID: 1, Text: "code", Col: 0, Row: 1}}
	m.clampWordsToWidth(m.innerWidth())
	if m.words[0].Col != minSpawnCol {
		t.Fatalf("col=%d want %d", m.words[0].Col, minSpawnCol)
	}
}

func TestClampWordsToWidth(t *testing.T) {
	m := testModel(t)
	m.width = 40
	m.words = []Word{{ID: 1, Text: "terminal", Col: 50, Row: 1}}
	m.clampWordsToWidth(m.innerWidth())
	inner := m.innerWidth()
	wlen := utf8.RuneCountInString("terminal")
	maxCol := maxSpawnCol(inner, wlen)
	if m.words[0].Col != maxCol {
		t.Fatalf("col=%d want %d", m.words[0].Col, maxCol)
	}
}

func TestSpawnWordRespectsColumnBounds(t *testing.T) {
	pool := NewWordPool([]string{"go", "code", "terminal"})
	rng := rand.New(rand.NewPCG(1, 2))
	w, ok := spawnWord(pool, 20, rng, 1, nil, 50, "")
	if !ok {
		t.Fatal("expected spawn ok")
	}
	wlen := utf8.RuneCountInString(w.Text)
	if w.Col < minSpawnCol || lockedSpanEnd(w.Col, wlen) >= 20 {
		t.Fatalf("col=%d text=%q exceeds inner width with brackets", w.Col, w.Text)
	}
}

func TestSelectLockCandidatePrefersHighestRow(t *testing.T) {
	words := []Word{
		{ID: 1, Text: "code", Row: 2},
		{ID: 2, Text: "cold", Row: 5},
		{ID: 3, Text: "cat", Row: 1},
	}
	id, ok := selectLockCandidate(words, 'c', "")
	if !ok || id != 2 {
		t.Fatalf("got id=%d ok=%v want id=2", id, ok)
	}
}
