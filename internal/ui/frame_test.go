package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestTopMiddleWidth(t *testing.T) {
	if got, want := TopMiddleWidth(60), 62; got != want {
		t.Fatalf("TopMiddleWidth(60) = %d, want %d", got, want)
	}
}

func TestFrameBodyInnerWidth(t *testing.T) {
	if got, want := FrameBodyInnerWidth(88), 84; got != want {
		t.Fatalf("FrameBodyInnerWidth(88) = %d, want %d", got, want)
	}
	if got, want := FrameBodyInnerWidth(24), 20; got != want {
		t.Fatalf("FrameBodyInnerWidth(24) = %d, want %d", got, want)
	}
	if got, want := FrameBodyInnerWidth(3), 1; got != want {
		t.Fatalf("FrameBodyInnerWidth(3) = %d, want %d", got, want)
	}
}

func TestTruncateTopCaption(t *testing.T) {
	if got := TruncateTopCaption("Session Stats", 80); got != "Session Stats" {
		t.Fatalf("got %q", got)
	}
	long := strings.Repeat("x", 100)
	got := TruncateTopCaption(long, 12)
	if lipgloss.Width(" "+got+" ") > 12 {
		t.Fatalf("truncated prefix still too wide: %q width %d", got, lipgloss.Width(" "+got+" "))
	}
}

func TestCenterBorderCaretXMatchesLayout(t *testing.T) {
	border := lipgloss.NewStyle()
	const middleCells = 40
	center := border.Render("|") + " hi " + border.Render("|")
	prefix := border.Render("|") + " hi"
	x := CenterBorderCaretX(true, "", border, middleCells, center, prefix)
	leftW := lipgloss.Width(center)
	rem := middleCells - leftW
	if rem < 0 {
		rem = 0
	}
	left := rem / 2
	leftSeg := border.Render("╭" + strings.Repeat("─", left))
	wantX := lipgloss.Width(leftSeg) + lipgloss.Width(prefix)
	if x != wantX {
		t.Fatalf("CenterBorderCaretX=%d want %d", x, wantX)
	}
}

func TestCenterBorderCaretXLeftMatchesLayout(t *testing.T) {
	border := lipgloss.NewStyle()
	const middleCells = 40
	center := border.Render("|") + " hi " + border.Render("|")
	prefix := border.Render("|") + " hi"
	const leftDashes = 3
	x := CenterBorderCaretXLeft(true, "", border, middleCells, center, prefix, leftDashes)
	line := RenderRoundedTopCenterBorderLeft("", border, center, middleCells, leftDashes)
	if lipgloss.Width(line) != lipgloss.Width(RenderRoundedTop("", border, lipgloss.NewStyle(), "x", middleCells)) {
		t.Fatalf("offset top width mismatch")
	}
	leftSeg := border.Render("╭" + strings.Repeat("─", leftDashes))
	wantX := lipgloss.Width(leftSeg) + lipgloss.Width(prefix)
	if x != wantX {
		t.Fatalf("CenterBorderCaretXLeft=%d want %d", x, wantX)
	}
}

func TestRenderRoundedTopHalvesWidth(t *testing.T) {
	border := lipgloss.NewStyle()
	cap := lipgloss.NewStyle()
	const middleCells = 40
	line := RenderRoundedTopHalves("", border, cap, "Left hand", "Right hand", middleCells)
	ref := RenderRoundedTop("", border, cap, "x", middleCells)
	if lipgloss.Width(line) != lipgloss.Width(ref) {
		t.Fatalf("halves width %d != single-caption width %d", lipgloss.Width(line), lipgloss.Width(ref))
	}
}

func TestRenderRoundedBottomHalvesWidth(t *testing.T) {
	border := lipgloss.NewStyle()
	cap := lipgloss.NewStyle()
	const middleCells = 40
	line := RenderRoundedBottomHalves("", border, cap, "Left", "Right", middleCells)
	ref := RenderRoundedBottomPlain("", border, middleCells)
	if lipgloss.Width(line) != lipgloss.Width(ref) {
		t.Fatalf("bottom halves width %d != plain bottom width %d", lipgloss.Width(line), lipgloss.Width(ref))
	}
}

func TestRenderRoundedTopCenterBorderWidth(t *testing.T) {
	border := lipgloss.NewStyle()
	cap := lipgloss.NewStyle().Bold(true)
	const middleCells = 40
	styled := cap.Render("> hello")
	line := RenderRoundedTopCenterBorder("", border, styled, middleCells)
	ref := RenderRoundedTop("", border, lipgloss.NewStyle(), "x", middleCells)
	if lipgloss.Width(line) != lipgloss.Width(ref) {
		t.Fatalf("center top width %d != ref width %d", lipgloss.Width(line), lipgloss.Width(ref))
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
