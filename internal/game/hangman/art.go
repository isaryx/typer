package hangman

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"

	"typer/internal/ui"
)

const (
	// DefaultMaxStrikes is the only supported strike count in v1 (stages 0..6).
	DefaultMaxStrikes = 6
	artLineCount      = 7
	artWidth          = 7

	// poleSeg is one full row of the vertical stick (right-hand post).
	poleSeg = "      |"
)

// Canonical final figure (stage 6) for tests.
var canonicalStage6 = [artLineCount]string{
	"  +---+",
	"  |   |",
	"  O   |",
	" /|\\  |",
	" / \\  |",
	poleSeg,
	"=======",
}

// stageArt redraws the entire gallows at each stage (not incremental fragments).
// Build order across 6 mistake steps (stages 0..6):
//  0 platform · 1 full pole · 2 pole+beam · 3 +rope · 4 +head · 5 +body · 6 +legs (lose).
var stageArt = [DefaultMaxStrikes + 1][artLineCount]string{
	0: {"       ", "       ", "       ", "       ", "       ", "       ", "======="},
	1: {poleSeg, poleSeg, poleSeg, poleSeg, poleSeg, poleSeg, "======="},
	2: {"  +---+", poleSeg, poleSeg, poleSeg, poleSeg, poleSeg, "======="},
	3: {"  +---+", "  |   |", poleSeg, poleSeg, poleSeg, poleSeg, "======="},
	4: {"  +---+", "  |   |", "  O   |", poleSeg, poleSeg, poleSeg, "======="},
	5: {"  +---+", "  |   |", "  O   |", " /|\\  |", poleSeg, poleSeg, "======="},
	6: canonicalStage6,
}

// Lines returns the art lines for stage (clamped to 0..DefaultMaxStrikes).
func Lines(stage int) []string {
	if stage < 0 {
		stage = 0
	}
	if stage > DefaultMaxStrikes {
		stage = DefaultMaxStrikes
	}
	out := make([]string, artLineCount)
	copy(out, stageArt[stage][:])
	return out
}

// CenteredLines pads each art line so the block is centered in innerWidth cells.
func CenteredLines(stage, innerWidth int) []string {
	raw := Lines(stage)
	if innerWidth < artWidth {
		innerWidth = artWidth
	}
	pad := (innerWidth - artWidth) / 2
	if pad < 0 {
		pad = 0
	}
	prefix := strings.Repeat(" ", pad)
	out := make([]string, artLineCount)
	for i, line := range raw {
		out[i] = prefix + line
	}
	return out
}

func strikeCaption(stage, maxStrikes, mistakeCount int) string {
	label := strconv.Itoa(stage) + "/" + strconv.Itoa(maxStrikes) + " strikes"
	if mistakeCount > 0 {
		label += " · " + strconv.Itoa(mistakeCount) + " mistakes"
	}
	return label
}

// RenderBoxWithStats draws the rounded hangman frame for the current game state.
func RenderBoxWithStats(s *State, layoutWidth int, border, caption lipgloss.Style) string {
	inner := ui.FrameBodyInnerWidth(layoutWidth)
	middle := ui.TopMiddleWidth(inner)
	label := strikeCaption(s.Stage(), s.MaxStrikes, s.MistakeCount())
	var b strings.Builder
	b.WriteString(ui.RenderRoundedTop("", border, caption, "hangman · "+label, middle))
	b.WriteByte('\n')
	for _, line := range CenteredLines(s.Stage(), inner) {
		b.WriteString(ui.RenderRoundedSide("", border, inner, line))
		b.WriteByte('\n')
	}
	b.WriteString(ui.RenderRoundedBottomPlain("", border, middle))
	return b.String()
}
