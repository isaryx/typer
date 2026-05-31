package defense

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/lipgloss/v2"

	"typer/internal/session"
	"typer/internal/ui"
)

type rowSegment struct {
	col      int
	rendered string
}

func (m *defenseModel) layoutWidth() int {
	tw := m.width
	if tw > ui.MaxContentWidth {
		return ui.MaxContentWidth
	}
	if tw < 1 {
		return 80
	}
	return tw
}

func (m *defenseModel) innerWidth() int {
	return ui.FrameBodyInnerWidth(m.layoutWidth())
}

func (m *defenseModel) renderView() string {
	if m.tooSmall {
		return session.RenderErrorText(m.styles, fmt.Sprintf(
			"Terminal too small for defense mode (need at least %d×%d). Press Ctrl+C to quit.",
			MinTerminalWidth, MinTerminalHeight,
		))
	}

	inner := m.innerWidth()
	middle := ui.TopMiddleWidth(inner)

	var b strings.Builder
	tw := m.layoutWidth()
	b.WriteString(m.styles.MetaStyle().Width(tw).Render("typer · defense"))
	b.WriteByte('\n')

	left, right := formatCaptionHalves(m)
	b.WriteString(ui.RenderRoundedTopHalves("", m.styles.BorderStyle(), m.styles.MetaStyle(), left, right, middle))
	b.WriteByte('\n')

	rows := m.buildPlayfieldRows(inner)
	for _, row := range rows {
		b.WriteString(ui.RenderRoundedSide("", m.styles.BorderStyle(), inner, row))
		b.WriteByte('\n')
	}

	return b.String()
}

func (m *defenseModel) buildPlayfieldRows(innerWidth int) []string {
	rows := m.initPlayfieldRows(innerWidth)
	shield := ShieldRow()
	segsByRow := make(map[int][]rowSegment)
	for _, w := range m.words {
		row := int(w.Row)
		if row < 0 || row >= shield {
			continue
		}
		locked := w.ID == m.lockID
		rendered := renderFallingWord(m.styles, w, locked, m.typed)
		col := w.Col
		if locked {
			// Brackets wrap the word without shifting its text: "[" sits one column left.
			col = effectiveSegmentCol(lockedSpanStart(col), rendered, innerWidth)
		}
		segsByRow[row] = append(segsByRow[row], rowSegment{col: col, rendered: rendered})
	}
	for row, segs := range segsByRow {
		rows[row] = composeWordRow(innerWidth, segs)
	}
	return rows
}

func (m *defenseModel) initPlayfieldRows(innerWidth int) []string {
	rows := make([]string, PlayfieldRows)
	shield := ShieldRow()
	baseStart := BaseArtStartRow()
	baseLines := RenderBaseLines(innerWidth, m.styles, m.baseHitFlashing(time.Now()))
	for i := 0; i < PlayfieldRows; i++ {
		if i == shield {
			rows[i] = session.RenderErrorText(m.styles, strings.Repeat("─", innerWidth))
			continue
		}
		if i >= baseStart && i < baseStart+BaseArtRows {
			rows[i] = baseLines[i-baseStart]
			continue
		}
		rows[i] = strings.Repeat(" ", innerWidth)
	}
	return rows
}

func effectiveSegmentCol(col int, rendered string, innerWidth int) int {
	width := lipgloss.Width(rendered)
	if width <= 0 {
		return col
	}
	if col+width <= innerWidth {
		return col
	}
	shifted := innerWidth - width
	if shifted < 0 {
		return 0
	}
	return shifted
}

func composeWordRow(innerWidth int, segments []rowSegment) string {
	if len(segments) == 0 {
		return strings.Repeat(" ", innerWidth)
	}
	sort.Slice(segments, func(i, j int) bool {
		if segments[i].col == segments[j].col {
			return segments[i].rendered < segments[j].rendered
		}
		return segments[i].col < segments[j].col
	})

	parts := make([]string, 0, len(segments)*2)
	pos := 0
	for _, seg := range segments {
		if seg.col >= innerWidth || seg.rendered == "" {
			continue
		}
		segW := lipgloss.Width(seg.rendered)
		if segW <= 0 {
			continue
		}
		if seg.col > pos {
			parts = append(parts, strings.Repeat(" ", seg.col-pos))
			pos = seg.col
		}
		if seg.col < pos {
			trim := pos - seg.col
			trimmed := trimStyledLeft(seg.rendered, trim)
			if trimmed == "" || lipgloss.Width(trimmed) <= 0 {
				continue
			}
			parts = append(parts, trimmed)
			pos += lipgloss.Width(trimmed)
			continue
		}
		parts = append(parts, seg.rendered)
		pos = seg.col + segW
	}
	if len(parts) == 0 {
		return strings.Repeat(" ", innerWidth)
	}
	line := lipgloss.JoinHorizontal(lipgloss.Top, parts...)
	w := lipgloss.Width(line)
	if w < innerWidth {
		line += strings.Repeat(" ", innerWidth-w)
	} else if w > innerWidth {
		line = truncateStyledWidth(line, innerWidth)
	}
	return line
}

func trimStyledLeft(s string, cells int) string {
	if cells <= 0 {
		return s
	}
	for cells > 0 && s != "" {
		if lipgloss.Width(s) <= cells {
			return ""
		}
		advance := styledPrefixByWidth(s, 1)
		if advance == 0 {
			break
		}
		s = s[advance:]
		cells--
	}
	return s
}

func styledPrefixByWidth(s string, width int) int {
	if width <= 0 {
		return 0
	}
	for i := 1; i <= len(s); i++ {
		if lipgloss.Width(s[:i]) >= width {
			return i
		}
	}
	return len(s)
}

func truncateStyledWidth(s string, max int) string {
	for lipgloss.Width(s) > max && len(s) > 0 {
		_, n := utf8.DecodeLastRuneInString(s)
		if n <= 0 {
			break
		}
		s = s[:len(s)-n]
	}
	return s
}

func renderFallingWord(styles session.Styles, w Word, locked bool, typed string) string {
	if !locked {
		return session.RenderUpcomingText(styles, w.Text)
	}
	open := session.RenderMetaText(styles, "[")
	close := session.RenderMetaText(styles, "]")
	body := session.RenderActiveWordText(styles, w.Text, typed)
	return open + body + close
}
