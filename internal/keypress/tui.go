package keypress

import (
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	colorTitle = "70"
	colorMeta  = "245"
	colorLabel = "45"
)

const historyMax = 10

type keyPressModel struct {
	width     int
	height    int
	lastLabel string
	history   []string
	title     lipgloss.Style
	meta      lipgloss.Style
	big       lipgloss.Style
}

func newKeyPressModel() keyPressModel {
	return keyPressModel{
		width:   80,
		height:  24,
		history: make([]string, 0, historyMax),
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorTitle)),
		meta: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMeta)),
		big: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorLabel)),
	}
}

func (m keyPressModel) Init() tea.Cmd {
	return nil
}

func (m keyPressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.width < 20 {
			m.width = 20
		}
		if m.height < 5 {
			m.height = 5
		}
		return m, nil
	case tea.KeyMsg:
		label := DisplayLabel(msg)
		m.lastLabel = label
		m.history = AppendHistory(m.history, label, historyMax)
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m keyPressModel) View() string {
	tw := m.width
	th := m.height

	title := m.title.Width(tw).Align(lipgloss.Center).Render("Key Press Viewer")

	labelText := m.lastLabel
	if labelText == "" {
		labelText = "(press a key)"
	}

	bigLine := m.big.Width(tw).Align(lipgloss.Center).Render(labelText)

	mid := lipgloss.JoinVertical(lipgloss.Center, title, "", bigLine)
	if len(m.history) > 0 {
		recent := m.meta.Width(tw).Align(lipgloss.Center).Render(
			strings.Join(m.history, " · "),
		)
		mid = lipgloss.JoinVertical(lipgloss.Center, mid, "", recent)
	}

	footer := m.meta.Width(tw).Align(lipgloss.Center).Render("Press Ctrl+C to quit.")

	stack := lipgloss.JoinVertical(lipgloss.Center, mid, "", footer)

	return lipgloss.Place(tw, th, lipgloss.Center, lipgloss.Center, stack)
}

// RunKeyPress runs the full-screen key display until Ctrl+C.
func RunKeyPress(input io.Reader, output io.Writer) error {
	p := tea.NewProgram(
		newKeyPressModel(),
		tea.WithInput(input),
		tea.WithOutput(output),
		tea.WithAltScreen(),
	)
	_, err := p.Run()
	return err
}
