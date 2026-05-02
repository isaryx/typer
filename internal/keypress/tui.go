package keypress

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const titleText = "Key Viewer"

const (
	colorMeta   = "245"
	colorLabel  = "87"
	colorBorder = "238"
	colorChipBG = "236"
	colorChipFG = "252"
	colorLastBG = "237"
	colorLastFG = "87"
)

const historyMax = 10

// historyNarrowTw stacks one chip per row at or below this width.
const historyNarrowTw = 48

// chipWrap renders a chip label without truncation; if wider than maxCols, text wraps (lipgloss Width).
func chipWrap(sty lipgloss.Style, text string, maxCols int) string {
	if maxCols < 1 {
		maxCols = 1
	}
	raw := sty.Render(text)
	if lipgloss.Width(raw) <= maxCols {
		return raw
	}
	return sty.Width(maxCols).Render(text)
}

// heroCard returns the bordered hero panel style with a blended rainbow outline (see lipgloss BorderForegroundBlend).
func heroCard(cardW int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForegroundBlend(
			lipgloss.Color("#FF6B6B"),
			lipgloss.Color("#FFA94D"),
			lipgloss.Color("#FFE066"),
			lipgloss.Color("#69DB7C"),
			lipgloss.Color("#4DABF7"),
			lipgloss.Color("#9775FA"),
			lipgloss.Color("#FF6B6B"),
		).
		Padding(1, 2).
		Width(cardW)
}

func randomTitleHex() string {
	return fmt.Sprintf("#%06x", rand.Uint32()&0xffffff)
}

type keyPressModel struct {
	width     int
	height    int
	lastLabel string
	titleFG   string // random hex; updates on key/paste
	history   []string
	title     lipgloss.Style
	meta      lipgloss.Style
	place     lipgloss.Style
	hero      lipgloss.Style
	chip      lipgloss.Style
	chipLast  lipgloss.Style
	footer    lipgloss.Style
}

func newKeyPressModel() keyPressModel {
	return keyPressModel{
		width:   80,
		height:  24,
		titleFG: randomTitleHex(),
		history: make([]string, 0, historyMax),
		title: lipgloss.NewStyle().
			Bold(true),
		meta: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMeta)),
		place: lipgloss.NewStyle().
			Faint(true).
			Foreground(lipgloss.Color(colorMeta)),
		hero: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorLabel)),
		chip: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color(colorChipBG)).
			Foreground(lipgloss.Color(colorChipFG)),
		chipLast: lipgloss.NewStyle().
			Padding(0, 1).
			Background(lipgloss.Color(colorLastBG)).
			Foreground(lipgloss.Color(colorLastFG)).
			Bold(true),
		footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMeta)),
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
	case tea.PasteMsg:
		m.titleFG = randomTitleHex()
		label := truncateRunes(msg.Content, maxDisplayLabelRunes)
		m.lastLabel = label
		m.history = AppendHistory(m.history, label, historyMax)
		return m, nil
	case tea.KeyPressMsg:
		label := DisplayLabel(msg)
		m.lastLabel = label
		m.history = AppendHistory(m.history, label, historyMax)
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		m.titleFG = randomTitleHex()
		return m, nil
	default:
		return m, nil
	}
}

func (m keyPressModel) View() tea.View {
	tw := m.width
	th := m.height

	title := m.title.
		Foreground(lipgloss.Color(m.titleFG)).
		Width(tw).
		Align(lipgloss.Center).
		Render(titleText)

	labelText := m.lastLabel
	heroSty := m.hero
	if labelText == "" {
		labelText = "(press a key)"
		heroSty = m.place
	}

	cardW := tw - 4
	if cardW < 16 {
		cardW = tw - 2
	}
	if cardW < 12 {
		cardW = tw
	}

	card := heroCard(cardW)

	innerW := cardW - 6 // rounded border + horizontal padding
	if innerW < 1 {
		innerW = 1
	}
	heroLine := heroSty.Width(innerW).Align(lipgloss.Center).Render(labelText)
	bigBox := card.Align(lipgloss.Center).Render(heroLine)

	mid := lipgloss.JoinVertical(lipgloss.Center, title, "", bigBox)

	if len(m.history) > 0 {
		recent := layoutHistoryChips(m.history, tw, m.chip, m.chipLast, historyNarrowTw)
		mid = lipgloss.JoinVertical(lipgloss.Center, mid, "", recent)
	}

	foot := m.footer.
		Width(tw).
		Align(lipgloss.Center).
		BorderTop(true).
		BorderForeground(lipgloss.Color(colorBorder)).
		Padding(1, 0).
		Render("Press Ctrl+C to quit.")

	stack := lipgloss.JoinVertical(lipgloss.Center, mid, "", foot)

	v := tea.NewView(lipgloss.Place(tw, th, lipgloss.Center, lipgloss.Center, stack))
	v.AltScreen = true
	return v
}

// layoutHistoryChips renders history as pills; packs horizontally with wrapping, or one column when narrow.
func layoutHistoryChips(history []string, tw int, chip, chipLast lipgloss.Style, narrowTw int) string {
	if len(history) == 0 {
		return ""
	}
	n := len(history)
	if tw <= narrowTw {
		inner := tw - 4
		if inner < 1 {
			inner = 1
		}
		lines := make([]string, 0, n)
		for i, text := range history {
			sty := chip
			if i == n-1 {
				sty = chipLast
			}
			line := chipWrap(sty, text, inner)
			lines = append(lines, lipgloss.NewStyle().Width(tw).Align(lipgloss.Center).Render(line))
		}
		return lipgloss.JoinVertical(lipgloss.Center, lines...)
	}

	const gap = 1
	var rows []string
	var row []string
	rowW := 0
	for i, text := range history {
		sty := chip
		if i == n-1 {
			sty = chipLast
		}
		chunk := chipWrap(sty, text, tw)
		w := lipgloss.Width(chunk)
		need := w
		if len(row) > 0 {
			need += gap
		}
		if len(row) > 0 && rowW+need > tw {
			rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, row...))
			row = nil
			rowW = 0
			need = w
		}
		if len(row) > 0 {
			row = append(row, strings.Repeat(" ", gap))
			rowW += gap
		}
		row = append(row, chunk)
		rowW += w
	}
	if len(row) > 0 {
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, row...))
	}
	joined := lipgloss.JoinVertical(lipgloss.Center, rows...)
	return lipgloss.NewStyle().Width(tw).Align(lipgloss.Center).Render(joined)
}

// RunKeyPress runs the full-screen key display until Ctrl+C or ctx cancellation.
func RunKeyPress(ctx context.Context, input io.Reader, output io.Writer) error {
	p := tea.NewProgram(
		newKeyPressModel(),
		tea.WithContext(ctx),
		tea.WithInput(input),
		tea.WithOutput(output),
	)
	_, err := p.Run()
	return err
}
