package session

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"typer/internal/ui"
)

var fingerAbbrev = []string{"Lp", "Lr", "Lm", "Li", "Lt", "Rt", "Ri", "Rm", "Rr", "Rp"}

// fingerHighlightSlots returns UI slot indices (0..9) to paint with fingerHi.
func fingerHighlightSlots(f Finger) []int {
	switch f {
	case FingerUnknown:
		return nil
	case FingerBothThumbs:
		return []int{4, 5}
	case FingerLeftPinky:
		return []int{0}
	case FingerLeftRing:
		return []int{1}
	case FingerLeftMiddle:
		return []int{2}
	case FingerLeftIndex:
		return []int{3}
	case FingerRightIndex:
		return []int{6}
	case FingerRightMiddle:
		return []int{7}
	case FingerRightRing:
		return []int{8}
	case FingerRightPinky:
		return []int{9}
	default:
		return nil
	}
}

func padAbbrevInSlot(abbr string, slotCells int) string {
	if slotCells <= 0 {
		return ""
	}
	aw := lipgloss.Width(abbr)
	if aw >= slotCells {
		return truncateDisplay(abbr, slotCells)
	}
	pad := slotCells - aw
	l := pad / 2
	r := pad - l
	return strings.Repeat(" ", l) + abbr + strings.Repeat(" ", r)
}

func truncateDisplay(s string, maxCells int) string {
	for lipgloss.Width(s) > maxCells && len([]rune(s)) > 0 {
		s = string([]rune(s)[:len([]rune(s))-1])
	}
	return s
}

func (m *typingSessionModel) renderFingerHandsFrame(selected Finger) string {
	tw := m.wrapWidth()
	middleCells := tw - 2
	contentW := m.promptInnerWidth()

	top := ui.RenderRoundedTopHalves("", m.styles.border, m.styles.meta, "Left hand", "Right hand", middleCells)

	highlight := fingerHighlightSlots(selected)
	hl := make(map[int]bool)
	for _, i := range highlight {
		hl[i] = true
	}

	sep := " │ "
	sepW := lipgloss.Width(sep)
	if contentW < sepW+20 {
		sep = "│"
		sepW = lipgloss.Width(sep)
	}

	avail := contentW - sepW
	slotW := avail / 10
	if slotW < 1 {
		slotW = 1
	}
	for slotW > 0 && (10*slotW+sepW > contentW) {
		slotW--
	}

	var parts []string
	for i := 0; i < 10; i++ {
		cell := padAbbrevInSlot(fingerAbbrev[i], slotW)
		st := m.styles.fingerDim
		if hl[i] {
			st = m.styles.fingerHi
		}
		parts = append(parts, st.Render(cell))
		if i == 4 {
			parts = append(parts, m.styles.border.Render(sep))
		}
	}
	inner := lipgloss.JoinHorizontal(lipgloss.Top, parts...)

	side := ui.RenderRoundedSide("", m.styles.border, contentW, inner)
	bottom := ui.RenderRoundedBottomPlain("", m.styles.border, middleCells)

	var b strings.Builder
	b.WriteString(top)
	b.WriteString("\n")
	b.WriteString(side)
	b.WriteString("\n")
	b.WriteString(bottom)
	return b.String()
}
