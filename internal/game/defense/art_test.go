package defense

import (
	"strings"
	"testing"
)

func TestCenteredBaseLines(t *testing.T) {
	lines := CenteredBaseLines(40)
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
