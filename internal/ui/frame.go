package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TopMiddleWidth returns the cell count between ╭ and ╮ for a frame whose body
// text width (between │ gutters) is contentInner.
func TopMiddleWidth(contentInner int) int {
	return contentInner + 2
}

// TruncateTopCaption trims label rune-wise until " "+result+" " fits in middleCells.
func TruncateTopCaption(label string, middleCells int) string {
	head := strings.TrimSpace(label)
	for {
		if lipgloss.Width(" "+head+" ") <= middleCells || head == "" {
			return head
		}
		r := []rune(head)
		if len(r) <= 1 {
			return ""
		}
		head = string(r[:len(r)-1])
	}
}

// RenderRoundedTop draws ╭ + styled caption + horizontal rule + ╮. marginLeft is
// prepended to the row (e.g. "  " for indented metrics, "" for passage).
func RenderRoundedTop(marginLeft string, border, caption lipgloss.Style, label string, middleCells int) string {
	trunc := TruncateTopCaption(label, middleCells)
	prefix := " " + trunc + " "
	fill := middleCells - lipgloss.Width(prefix)
	if fill < 0 {
		fill = 0
	}
	core := lipgloss.JoinHorizontal(lipgloss.Top,
		border.Render("╭"),
		caption.Render(prefix),
		border.Render(strings.Repeat("─", fill)+"╮"),
	)
	if marginLeft == "" {
		return core
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Render(marginLeft), core)
}

// RenderRoundedSide draws │ + padded inner + │.
func RenderRoundedSide(marginLeft string, border lipgloss.Style, innerWidth int, inner string) string {
	padded := lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Left).Render(inner)
	core := lipgloss.JoinHorizontal(lipgloss.Top,
		border.Render("│"),
		" ",
		padded,
		" ",
		border.Render("│"),
	)
	if marginLeft == "" {
		return core
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Render(marginLeft), core)
}

// RenderRoundedBottomPlain draws ╰ + horizontal rule + ╯ between corners (middleCells wide).
func RenderRoundedBottomPlain(marginLeft string, border lipgloss.Style, middleCells int) string {
	core := border.Render("╰" + strings.Repeat("─", middleCells) + "╯")
	if marginLeft == "" {
		return core
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Render(marginLeft), core)
}

// FitBottomCaption fits caption (no leading space) for the bottom border row; returns
// plain text for meta.Render (includes leading space and optional …) and left dash count.
func FitBottomCaption(caption string, middleCells, rightDashes int) (capPlain string, leftDashes int) {
	d2 := rightDashes
	maxCap := middleCells - d2
	core := caption
	shortened := false
	for {
		capTry := " " + core
		if shortened {
			capTry += "…"
		}
		if lipgloss.Width(capTry) <= maxCap || len([]rune(core)) <= 1 {
			break
		}
		rs := []rune(core)
		core = string(rs[:len(rs)-1])
		shortened = true
	}
	capPlain = " " + core
	if shortened {
		capPlain += "…"
	}
	d1 := middleCells - lipgloss.Width(capPlain) - d2
	if d1 < 0 {
		d1 = 0
	}
	return capPlain, d1
}

// RenderRoundedBottomCaption draws ╰ + dashes + styled caption + dashes + ╯.
func RenderRoundedBottomCaption(marginLeft string, border, caption lipgloss.Style, leftDashes int, capPlain string, rightDashes int) string {
	core := lipgloss.JoinHorizontal(lipgloss.Top,
		border.Render("╰"+strings.Repeat("─", leftDashes)),
		caption.Render(capPlain),
		border.Render(strings.Repeat("─", rightDashes)+"╯"),
	)
	if marginLeft == "" {
		return core
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Render(marginLeft), core)
}

// BuildRoundedTopPlain returns one top frame line with optional indent before ╭ (e.g. "  ").
func BuildRoundedTopPlain(indent string, label string, contentInner int) string {
	mw := TopMiddleWidth(contentInner)
	head := TruncateTopCaption(label, mw)
	prefix := " " + head + " "
	fill := mw - lipgloss.Width(prefix)
	if fill < 0 {
		fill = 0
	}
	return fmt.Sprintf("%s╭%s%s╮\n", indent, prefix, strings.Repeat("─", fill))
}
