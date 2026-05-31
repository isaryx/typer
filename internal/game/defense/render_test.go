package defense

import (
	"strings"
	"testing"

	"typer/internal/session"
)

func TestComposeWordRowShowsBothWords(t *testing.T) {
	styles := session.DefaultStyles()
	segs := []rowSegment{
		{col: 2, rendered: session.RenderUpcomingText(styles, "alpha")},
		{col: 14, rendered: session.RenderUpcomingText(styles, "beta")},
	}
	line := composeWordRow(40, segs)
	plain := stripANSI(line)
	if !strings.Contains(plain, "alpha") || !strings.Contains(plain, "beta") {
		t.Fatalf("expected both words on row, got %q", plain)
	}
}

func TestLockedWordKeepsTextColumn(t *testing.T) {
	styles := session.DefaultStyles()
	w := Word{Text: "code", Col: 10}
	unlocked := composeWordRow(40, []rowSegment{
		{col: w.Col, rendered: renderFallingWord(styles, w, false, "")},
	})
	locked := composeWordRow(40, []rowSegment{
		{col: lockedSpanStart(w.Col), rendered: renderFallingWord(styles, w, true, "")},
	})
	unlockedPlain := stripANSI(unlocked)
	lockedPlain := stripANSI(locked)
	unlockedIdx := strings.Index(unlockedPlain, "code")
	lockedIdx := strings.Index(lockedPlain, "code")
	if unlockedIdx < 0 || lockedIdx < 0 {
		t.Fatalf("word missing: unlocked=%q locked=%q", unlockedPlain, lockedPlain)
	}
	if unlockedIdx != lockedIdx {
		t.Fatalf("word shifted on lock: unlocked at %d locked at %d", unlockedIdx, lockedIdx)
	}
	if lockedIdx > 0 && lockedPlain[lockedIdx-1] != '[' {
		t.Fatalf("expected [ before word, got %q", lockedPlain)
	}
}

func TestEffectiveSegmentColClampsLockedWord(t *testing.T) {
	styles := session.DefaultStyles()
	w := Word{Text: "terminal", Col: 70}
	rendered := renderFallingWord(styles, w, true, "term")
	innerWidth := 76
	col := effectiveSegmentCol(lockedSpanStart(w.Col), rendered, innerWidth)
	width := len([]rune("terminal")) + lockBracketCols
	if col+width > innerWidth {
		t.Fatalf("col=%d width=%d exceeds innerWidth=%d", col, width, innerWidth)
	}
}

func TestLockedWordAtMaxColKeepsPosition(t *testing.T) {
	styles := session.DefaultStyles()
	innerWidth := 40
	w := Word{Text: "code", Col: maxSpawnCol(innerWidth, 4)}
	rendered := renderFallingWord(styles, w, true, "")
	col := effectiveSegmentCol(lockedSpanStart(w.Col), rendered, innerWidth)
	line := composeWordRow(innerWidth, []rowSegment{{col: col, rendered: rendered}})
	plain := stripANSI(line)
	wordIdx := strings.Index(plain, "code")
	if wordIdx != w.Col {
		t.Fatalf("word at %d want %d in %q", wordIdx, w.Col, plain)
	}
	if plain[len(plain)-1] != ']' {
		t.Fatalf("closing bracket should fit at right edge, got %q", plain)
	}
}

func TestLockedRenderNearRightEdge(t *testing.T) {
	styles := session.DefaultStyles()
	w := Word{Text: "code", Col: 72}
	rendered := renderFallingWord(styles, w, true, "co")
	innerWidth := 76
	col := effectiveSegmentCol(lockedSpanStart(w.Col), rendered, innerWidth)
	segs := []rowSegment{{col: col, rendered: rendered}}
	line := composeWordRow(innerWidth, segs)
	if len([]rune(stripANSI(line))) > innerWidth+2 {
		t.Fatalf("line too wide: %d", len(stripANSI(line)))
	}
	if !strings.Contains(stripANSI(line), "code") {
		t.Fatalf("word should remain visible: %q", stripANSI(line))
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	in := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			in = true
			continue
		}
		if in {
			if s[i] == 'm' {
				in = false
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
