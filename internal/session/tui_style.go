package session

import "github.com/charmbracelet/lipgloss"

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
		title:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("70")),
		meta:      lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		promptBox: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#ffffff")),
		completed: lipgloss.NewStyle().Foreground(lipgloss.Color("70")),
		completedBad: lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true),
		active:       lipgloss.NewStyle().Bold(true).Underline(true).Foreground(lipgloss.Color("45")),
		activeTyped:  lipgloss.NewStyle().Foreground(lipgloss.Color("45")),
		activeCursor: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Background(lipgloss.Color("33")),
		upcoming:     lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		input:        lipgloss.NewStyle().Foreground(lipgloss.Color("229")),
		inputBad:     lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true),
		errorMessage: lipgloss.NewStyle().Foreground(lipgloss.Color("203")),
	}
}
