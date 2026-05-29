package hangman

import (
	"reflect"
	"strings"
	"testing"
)

func TestLinesStage0IsPlatformOnly(t *testing.T) {
	got := Lines(0)
	if got[6] != "=======" {
		t.Fatalf("stage 0 ground: %q", got[6])
	}
	for i := 0; i < 6; i++ {
		if strings.TrimSpace(got[i]) != "" {
			t.Fatalf("stage 0 line %d should be blank, got %q", i, got[i])
		}
	}
}

func TestLinesStage1DrawsFullPole(t *testing.T) {
	got := Lines(1)
	for i := 0; i < 6; i++ {
		if got[i] != poleSeg {
			t.Fatalf("stage 1 line %d = %q, want full pole %q", i, got[i], poleSeg)
		}
	}
}

func TestLinesStageProgression(t *testing.T) {
	if !strings.Contains(Lines(2)[0], "+") {
		t.Fatal("stage 2 should show overhead beam")
	}
	for i := 1; i < 6; i++ {
		if Lines(2)[i] != poleSeg {
			t.Fatalf("stage 2 line %d should redraw full pole", i)
		}
	}
	if strings.TrimSpace(Lines(3)[1]) == "" {
		t.Fatal("stage 3 should show rope row")
	}
	for i := 2; i < 6; i++ {
		if Lines(3)[i] != poleSeg {
			t.Fatalf("stage 3 line %d should redraw full pole", i)
		}
	}
	if !strings.Contains(Lines(4)[2], "O") {
		t.Fatal("stage 4 should show head")
	}
	if !strings.Contains(Lines(5)[3], "/") {
		t.Fatal("stage 5 should show body")
	}
}

func TestLinesStage6MatchesCanonical(t *testing.T) {
	got := Lines(DefaultMaxStrikes)
	want := canonicalStage6[:]
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("stage 6 art:\n got %q\nwant %q", got, want)
	}
}

func TestLinesClampsNegativeAndHigh(t *testing.T) {
	if got := Lines(-1); !reflect.DeepEqual(got, Lines(0)) {
		t.Fatalf("negative stage should clamp to 0")
	}
	if got := Lines(99); !reflect.DeepEqual(got, Lines(DefaultMaxStrikes)) {
		t.Fatalf("high stage should clamp to max")
	}
}

func TestCenteredLinesWidth(t *testing.T) {
	lines := CenteredLines(3, 20)
	for i, line := range lines {
		if len(line) < artWidth {
			t.Fatalf("line %d too short: %q", i, line)
		}
	}
}
