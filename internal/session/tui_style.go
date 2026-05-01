package session

import (
	"github.com/charmbracelet/lipgloss"

	"typer/internal/ui"
)

type tuiStyles struct {
	title        lipgloss.Style
	meta         lipgloss.Style
	promptPrefix lipgloss.Style
	completed    lipgloss.Style
	completedBad lipgloss.Style
	active       lipgloss.Style
	activePlain  lipgloss.Style // same hue as active word, no underline (caret uses combining mark)
	activeTyped  lipgloss.Style
	upcoming     lipgloss.Style
	input        lipgloss.Style
	inputBad     lipgloss.Style
	errorMessage lipgloss.Style
	border       lipgloss.Style // passage frame horizontal rules
	fingerDim    lipgloss.Style // finger abbrev inactive
	fingerHi     lipgloss.Style // finger abbrev suggested (bg fill)
}

func defaultStyles() tuiStyles {
	return tuiStyles{
		title:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ui.ColorTitle)),
		meta:         lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorMeta)),
		promptPrefix: lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorMeta)),
		completed:    lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorCompletedOK)),
		completedBad: lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorCompletedBad)).Bold(true),
		active:       lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color(ui.ColorActiveFg)),
		activePlain:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(ui.ColorActiveFg)),
		activeTyped:  lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorActiveFg)),
		upcoming:     lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorUpcoming)),
		input:        lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorInputFg)),
		inputBad:     lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorInputBadFg)).Bold(true),
		errorMessage: lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorErrorFg)),
		border:       lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorBorderHex)),
		fingerDim:    lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorFingerDimFg)),
		fingerHi: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ColorFingerHighlightFg)).
			Background(lipgloss.Color(ui.ColorFingerHighlightBg)).
			Bold(true),
	}
}
