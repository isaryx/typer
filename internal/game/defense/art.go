package defense

import (
	"strings"

	"charm.land/lipgloss/v2"

	"typer/internal/session"
)

const BaseArtRows = 2

var baseArtLines = [BaseArtRows][2]string{
	{"( ^_^ )", "( T_T )"},
	{" \\___/ ", " \\___/ "},
}

// RenderBaseLines returns the robot defender art centered for innerWidth.
// When flash is true the face turns sad and the art is styled red briefly.
func RenderBaseLines(innerWidth int, styles session.Styles, flash bool) []string {
	face := 0
	if flash {
		face = 1
	}
	styleLine := func(line string) string { return line }
	if flash {
		styleLine = func(line string) string { return session.RenderErrorText(styles, line) }
	}
	out := make([]string, BaseArtRows)
	for i := range baseArtLines {
		line := baseArtLines[i][face]
		w := lipgloss.Width(line)
		if w > innerWidth {
			out[i] = strings.Repeat(" ", innerWidth)
			continue
		}
		pad := (innerWidth - w) / 2
		if pad < 0 {
			pad = 0
		}
		out[i] = strings.Repeat(" ", pad) + styleLine(line)
	}
	return out
}
