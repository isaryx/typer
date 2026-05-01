package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestTopMiddleWidth(t *testing.T) {
	if got, want := TopMiddleWidth(60), 62; got != want {
		t.Fatalf("TopMiddleWidth(60) = %d, want %d", got, want)
	}
}

func TestTruncateTopCaption(t *testing.T) {
	if got := TruncateTopCaption("Session stats", 80); got != "Session stats" {
		t.Fatalf("got %q", got)
	}
	long := strings.Repeat("x", 100)
	got := TruncateTopCaption(long, 12)
	if lipgloss.Width(" "+got+" ") > 12 {
		t.Fatalf("truncated prefix still too wide: %q width %d", got, lipgloss.Width(" "+got+" "))
	}
}

func TestFitBottomCaption(t *testing.T) {
	cap, d1 := FitBottomCaption("by Henry Ford", 40, 1)
	if cap == "" {
		t.Fatal("empty cap")
	}
	if d1 < 0 {
		t.Fatalf("d1 = %d", d1)
	}
	if lipgloss.Width(cap)+d1+1 > 40 {
		t.Fatalf("bottom row wider than middle: capW=%d d1=%d", lipgloss.Width(cap), d1)
	}
}
