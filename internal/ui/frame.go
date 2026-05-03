package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// TopMiddleWidth returns the cell count between ╭ and ╮ for a frame whose body
// text width (between │ gutters) is contentInner.
func TopMiddleWidth(contentInner int) int {
	return contentInner + 2
}

// FrameBodyInnerWidth is the padded text width between "│ " and " │" in RenderRoundedSide
// when the frame's outer width (╭ through ╮ on the top row, │ through │ on body rows) is layoutWidth.
// Session typing UI uses the same value as promptInnerWidth(wrapWidth).
func FrameBodyInnerWidth(layoutWidth int) int {
	w := layoutWidth - 4
	if w < 1 {
		return 1
	}
	return w
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
// RenderRoundedTopHalves draws ╭─╮ with leftLabel left-aligned in the left half of the rule and
// rightLabel right-aligned in the right half (only ─ between segments; no mid junction glyph).
// middleCells is the cell count strictly between the corner characters, matching RenderRoundedTop.
func RenderRoundedTopHalves(marginLeft string, border, caption lipgloss.Style, leftLabel, rightLabel string, middleCells int) string {
	if middleCells < 4 {
		return RenderRoundedTop(marginLeft, border, caption, leftLabel+" · "+rightLabel, middleCells)
	}
	leftHalf := middleCells / 2
	rightHalf := middleCells - leftHalf
	leftSeg := leftCaptionSegment(leftLabel, leftHalf, border, caption)
	rightSeg := rightCaptionSegment(rightLabel, rightHalf, border, caption)
	core := lipgloss.JoinHorizontal(lipgloss.Top,
		border.Render("╭"),
		leftSeg,
		rightSeg,
		border.Render("╮"),
	)
	if marginLeft == "" {
		return core
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Render(marginLeft), core)
}

func leftCaptionSegment(label string, segmentCells int, border, caption lipgloss.Style) string {
	if segmentCells <= 0 {
		return ""
	}
	trunc := strings.TrimSpace(label)
	for {
		prefix := " " + trunc
		if lipgloss.Width(caption.Render(prefix)) <= segmentCells || len([]rune(trunc)) == 0 {
			break
		}
		trunc = string([]rune(trunc)[:len([]rune(trunc))-1])
	}
	prefix := " " + trunc
	rendered := caption.Render(prefix)
	w := lipgloss.Width(rendered)
	dash := segmentCells - w
	if dash < 0 {
		dash = 0
	}
	return rendered + border.Render(strings.Repeat("─", dash))
}

func rightCaptionSegment(label string, segmentCells int, border, caption lipgloss.Style) string {
	if segmentCells <= 0 {
		return ""
	}
	// One cell for trailing ─ before ╮, matching bottom caption row (one ─ before ╯).
	if segmentCells == 1 {
		return border.Render("─")
	}
	inner := segmentCells - 1
	trunc := strings.TrimSpace(label)
	for {
		core := " " + trunc
		if lipgloss.Width(caption.Render(core)) <= inner || len([]rune(trunc)) == 0 {
			break
		}
		trunc = string([]rune(trunc)[:len([]rune(trunc))-1])
	}
	core := " " + trunc
	rendered := caption.Render(core)
	w := lipgloss.Width(rendered)
	dash := inner - w
	if dash < 0 {
		dash = 0
	}
	return border.Render(strings.Repeat("─", dash)) + rendered + border.Render("─")
}

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

// RenderRoundedTopCenterBorder draws ╭ + dashes + centerStyled + dashes + ╮; centerStyled must already be rendered (may contain ANSI).
func RenderRoundedTopCenterBorder(marginLeft string, border lipgloss.Style, centerStyled string, middleCells int) string {
	w := lipgloss.Width(centerStyled)
	rem := middleCells - w
	if rem < 0 {
		rem = 0
	}
	return RenderRoundedTopCenterBorderLeft(marginLeft, border, centerStyled, middleCells, rem/2)
}

// RenderRoundedTopCenterBorderLeft draws the same row as RenderRoundedTopCenterBorder but places exactly leftDashes cells of ─ after ╭ (clamped into the slack around centerStyled).
func RenderRoundedTopCenterBorderLeft(marginLeft string, border lipgloss.Style, centerStyled string, middleCells, leftDashes int) string {
	w := lipgloss.Width(centerStyled)
	rem := middleCells - w
	if rem < 0 {
		rem = 0
	}
	left := leftDashes
	if left < 0 {
		left = 0
	}
	if left > rem {
		left = rem
	}
	right := rem - left
	core := lipgloss.JoinHorizontal(lipgloss.Top,
		border.Render("╭"+strings.Repeat("─", left)),
		centerStyled,
		border.Render(strings.Repeat("─", right)+"╮"),
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

// RenderRoundedBottomHalves draws ╰ + left caption half + right caption half + ╯ (same segment split as RenderRoundedTopHalves).
func RenderRoundedBottomHalves(marginLeft string, border, caption lipgloss.Style, leftLabel, rightLabel string, middleCells int) string {
	if middleCells < 4 {
		combined := strings.TrimSpace(strings.TrimSpace(leftLabel) + " · " + strings.TrimSpace(rightLabel))
		combined = strings.TrimSpace(combined)
		if combined == "" {
			return RenderRoundedBottomPlain(marginLeft, border, middleCells)
		}
		capPlain, d1 := FitBottomCaption(combined, middleCells, 1)
		return RenderRoundedBottomCaption(marginLeft, border, caption, d1, capPlain, 1)
	}
	leftHalf := middleCells / 2
	rightHalf := middleCells - leftHalf
	leftSeg := leftCaptionSegment(leftLabel, leftHalf, border, caption)
	rightSeg := rightCaptionSegment(rightLabel, rightHalf, border, caption)
	core := lipgloss.JoinHorizontal(lipgloss.Top,
		border.Render("╰"),
		leftSeg,
		rightSeg,
		border.Render("╯"),
	)
	if marginLeft == "" {
		return core
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Render(marginLeft), core)
}

// RenderRoundedBottomCenterBorder draws ╰ + dashes + centerStyled + dashes + ╯.
func RenderRoundedBottomCenterBorder(marginLeft string, border lipgloss.Style, centerStyled string, middleCells int) string {
	w := lipgloss.Width(centerStyled)
	rem := middleCells - w
	if rem < 0 {
		rem = 0
	}
	return RenderRoundedBottomCenterBorderLeft(marginLeft, border, centerStyled, middleCells, rem/2)
}

// RenderRoundedBottomCenterBorderLeft draws the same row as RenderRoundedBottomCenterBorder but places exactly leftDashes cells of ─ after ╰.
func RenderRoundedBottomCenterBorderLeft(marginLeft string, border lipgloss.Style, centerStyled string, middleCells, leftDashes int) string {
	w := lipgloss.Width(centerStyled)
	rem := middleCells - w
	if rem < 0 {
		rem = 0
	}
	left := leftDashes
	if left < 0 {
		left = 0
	}
	if left > rem {
		left = rem
	}
	right := rem - left
	core := lipgloss.JoinHorizontal(lipgloss.Top,
		border.Render("╰"+strings.Repeat("─", left)),
		centerStyled,
		border.Render(strings.Repeat("─", right)+"╯"),
	)
	if marginLeft == "" {
		return core
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Render(marginLeft), core)
}

// CenterBorderCaretX returns the 0-based column for the terminal insertion caret immediately after
// insertionPrefix inside the centered segment (same geometry as RenderRoundedTopCenterBorder /
// RenderRoundedBottomCenterBorder). topRow selects ╭ vs ╰ for the left corner run.
func CenterBorderCaretX(topRow bool, marginLeft string, border lipgloss.Style, middleCells int, centerStyled, insertionPrefix string) int {
	w := lipgloss.Width(centerStyled)
	rem := middleCells - w
	if rem < 0 {
		rem = 0
	}
	return CenterBorderCaretXLeft(topRow, marginLeft, border, middleCells, centerStyled, insertionPrefix, rem/2)
}

// CenterBorderCaretXLeft matches RenderRoundedTopCenterBorderLeft / RenderRoundedBottomCenterBorderLeft with the same leftDashes.
func CenterBorderCaretXLeft(topRow bool, marginLeft string, border lipgloss.Style, middleCells int, centerStyled, insertionPrefix string, leftDashes int) int {
	w := lipgloss.Width(centerStyled)
	rem := middleCells - w
	if rem < 0 {
		rem = 0
	}
	left := leftDashes
	if left < 0 {
		left = 0
	}
	if left > rem {
		left = rem
	}
	corner := "╭"
	if !topRow {
		corner = "╰"
	}
	leftSeg := border.Render(corner + strings.Repeat("─", left))
	marginW := 0
	if marginLeft != "" {
		marginW = lipgloss.Width(lipgloss.NewStyle().Render(marginLeft))
	}
	return marginW + lipgloss.Width(leftSeg) + lipgloss.Width(insertionPrefix)
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
