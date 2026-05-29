package defense

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const BaseArtRows = 2

var baseArtLines = [BaseArtRows]string{
	"( ^_^ )",
	" \\___/ ",
}

// CenteredBaseLines returns the robot defender art centered for innerWidth.
func CenteredBaseLines(innerWidth int) []string {
	out := make([]string, BaseArtRows)
	for i, line := range baseArtLines {
		w := lipgloss.Width(line)
		if w > innerWidth {
			out[i] = strings.Repeat(" ", innerWidth)
			continue
		}
		pad := (innerWidth - w) / 2
		if pad < 0 {
			pad = 0
		}
		out[i] = strings.Repeat(" ", pad) + line
	}
	return out
}
