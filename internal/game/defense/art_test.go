package defense

import (
	"strings"
	"testing"

	"typer/internal/session"
)

func TestRenderBaseLines(t *testing.T) {
	lines := RenderBaseLines(40, session.DefaultStyles(), false)
	if len(lines) != BaseArtRows {
		t.Fatalf("len=%d want %d", len(lines), BaseArtRows)
	}
	if !strings.Contains(lines[0], "( ^_^ )") {
		t.Fatalf("row0=%q", lines[0])
	}
	if !strings.Contains(lines[1], `\___/`) {
		t.Fatalf("row1=%q", lines[1])
	}
}

func TestRenderBaseLinesFlashUsesErrorStyle(t *testing.T) {
	plain := RenderBaseLines(40, session.DefaultStyles(), false)
	flash := RenderBaseLines(40, session.DefaultStyles(), true)
	if plain[0] == flash[0] {
		t.Fatalf("flash should change styling, plain=%q flash=%q", plain[0], flash[0])
	}
	if !strings.Contains(flash[0], "( T_T )") {
		t.Fatalf("flash should show sad face: %q", flash[0])
	}
	if strings.Contains(flash[0], "( ^_^ )") {
		t.Fatalf("flash should not show happy face: %q", flash[0])
	}
	if !strings.Contains(flash[0], "\x1b[") {
		t.Fatalf("flash should include ANSI styling: %q", flash[0])
	}
}

func TestShieldAndBaseRows(t *testing.T) {
	if ShieldRow() != PlayfieldRows-BaseArtRows-1 {
		t.Fatalf("shield row layout mismatch")
	}
	if BaseArtStartRow() != ShieldRow()+1 {
		t.Fatalf("base should follow shield")
	}
	if BaseArtStartRow()+BaseArtRows != PlayfieldRows {
		t.Fatalf("base art should fill bottom rows")
	}
}

func TestWordHitTriggersBaseFlash(t *testing.T) {
	m := testModel(t)
	m.words = []Word{{ID: 1, Text: "code", Col: minSpawnCol, Row: float64(ShieldRow())}}
	now := m.startedAt.Add(TickInterval)
	m.applyTick(now)
	if m.baseHitFlashUntil.IsZero() {
		t.Fatal("expected base hit flash after word reaches shield")
	}
	if !m.baseHitFlashing(now) {
		t.Fatal("base should flash immediately after hit")
	}
	if m.baseHitFlashing(now.Add(BaseHitFlashDuration)) {
		t.Fatal("base flash should end after duration")
	}
}
