package session

import "github.com/charmbracelet/lipgloss"

// Palette used by the typing TUI. Numbers are ANSI 256-color indices except
// where noted; keep them grouped here so themes are easy to swap.
const (
	colorTitle        = "70"      // muted green
	colorMeta         = "245"     // dim gray
	colorBorder       = "#ffffff" // pure white
	colorCompletedOK  = "70"      // muted green
	colorCompletedBad = "203"     // salmon red
	colorActiveFg     = "45"      // bright cyan
	colorCursorFg     = "229"     // pale yellow
	colorCursorBg     = "33"      // bright blue
	colorUpcoming     = "252"     // light gray
	colorInputFg      = "229"     // pale yellow
	colorInputBadFg   = "203"     // salmon red
	colorErrorFg      = "203"     // salmon red
)

type tuiStyles struct {
	title        lipgloss.Style
	meta         lipgloss.Style
	promptBox    lipgloss.Style
	completed    lipgloss.Style
	completedBad lipgloss.Style
	active       lipgloss.Style
	activeTyped  lipgloss.Style
	activeCursor lipgloss.Style
	upcoming     lipgloss.Style
	input        lipgloss.Style
	inputBad     lipgloss.Style
	errorMessage lipgloss.Style
}

func defaultStyles() tuiStyles {
	return tuiStyles{
		title:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorTitle)),
		meta:         lipgloss.NewStyle().Foreground(lipgloss.Color(colorMeta)),
		promptBox:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(colorBorder)),
		completed:    lipgloss.NewStyle().Foreground(lipgloss.Color(colorCompletedOK)),
		completedBad: lipgloss.NewStyle().Foreground(lipgloss.Color(colorCompletedBad)).Bold(true),
		active:       lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color(colorActiveFg)),
		activeTyped:  lipgloss.NewStyle().Foreground(lipgloss.Color(colorActiveFg)),
		activeCursor: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorCursorFg)).Background(lipgloss.Color(colorCursorBg)),
		upcoming:     lipgloss.NewStyle().Foreground(lipgloss.Color(colorUpcoming)),
		input:        lipgloss.NewStyle().Foreground(lipgloss.Color(colorInputFg)),
		inputBad:     lipgloss.NewStyle().Foreground(lipgloss.Color(colorInputBadFg)).Bold(true),
		errorMessage: lipgloss.NewStyle().Foreground(lipgloss.Color(colorErrorFg)),
	}
}
