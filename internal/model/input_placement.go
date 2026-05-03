package model

import (
	"fmt"
	"strings"
)

// Vertical placement of the typed input line relative to the passage frame.
type InputVertical uint8

const (
	InputVerticalBottom InputVertical = iota
	InputVerticalTop
	// InputVerticalOnTopBorder places the typed input centered in the passage frame top border (╭─╮).
	InputVerticalOnTopBorder
	// InputVerticalOnBottomBorder places the typed input centered in the passage frame bottom border (╰─╯).
	InputVerticalOnBottomBorder
	// InputVerticalOnTopBorderDynamic is like OnTopBorder but shifts the input segment horizontally to track the active word.
	InputVerticalOnTopBorderDynamic
	// InputVerticalOnBottomBorderDynamic is like OnBottomBorder but shifts the input segment horizontally to track the active word.
	InputVerticalOnBottomBorderDynamic
)

// Horizontal alignment of the typed input line within the content width.
type InputHorizontal uint8

const (
	InputHorizontalLeft InputHorizontal = iota
	InputHorizontalCenter
	InputHorizontalRight
)

// InputPlacement configures where the "> …" input row appears in the typing TUI.
// The Go zero value is bottom-left (both enums zero); settings omitting input_position use DefaultInputPlacement.
// Horizontal alignment (H) applies to top/bottom rows inside the layout; border placements (on-top, on-bottom, dynamic variants) ignore H and center the segment in the frame rule.
type InputPlacement struct {
	V InputVertical
	H InputHorizontal
}

// DefaultInputPlacement is used when stored settings leave input_position empty or invalid.
func DefaultInputPlacement() InputPlacement {
	return InputPlacement{V: InputVerticalTop, H: InputHorizontalLeft}
}

// CanonicalString returns a stable form for JSON storage and CLI display (e.g. "bottom-left").
func (p InputPlacement) CanonicalString() string {
	switch p.V {
	case InputVerticalOnTopBorder:
		return "on-top"
	case InputVerticalOnBottomBorder:
		return "on-bottom"
	case InputVerticalOnTopBorderDynamic:
		return "on-top-dynamic"
	case InputVerticalOnBottomBorderDynamic:
		return "on-bottom-dynamic"
	}
	var vert string
	switch p.V {
	case InputVerticalTop:
		vert = "top"
	default:
		vert = "bottom"
	}
	var horiz string
	switch p.H {
	case InputHorizontalCenter:
		horiz = "center"
	case InputHorizontalRight:
		horiz = "right"
	default:
		horiz = "left"
	}
	return vert + "-" + horiz
}

// Shorthand: tl, tc, tr, bl, bc, br (top/bottom × left/center/right); ot, ob (on top/bottom border).
var inputPositionShorthand = map[string]InputPlacement{
	"tl": {V: InputVerticalTop, H: InputHorizontalLeft},
	"tc": {V: InputVerticalTop, H: InputHorizontalCenter},
	"tr": {V: InputVerticalTop, H: InputHorizontalRight},
	"bl": {V: InputVerticalBottom, H: InputHorizontalLeft},
	"bc": {V: InputVerticalBottom, H: InputHorizontalCenter},
	"br": {V: InputVerticalBottom, H: InputHorizontalRight},
	"ot":  {V: InputVerticalOnTopBorder, H: InputHorizontalCenter},
	"ob":  {V: InputVerticalOnBottomBorder, H: InputHorizontalCenter},
	"otd": {V: InputVerticalOnTopBorderDynamic, H: InputHorizontalCenter},
	"obd": {V: InputVerticalOnBottomBorderDynamic, H: InputHorizontalCenter},
}

// ParseInputPosition parses values like "top-left", "bottom-center", "on-top", "on-top-dynamic", or shorthand "tl", "bc", "ot", "otd".
func ParseInputPosition(s string) (InputPlacement, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return InputPlacement{}, fmt.Errorf(`input position: empty (expected e.g. "bottom-left", "bc", "top-center")`)
	}
	switch s {
	case "on-top":
		return InputPlacement{V: InputVerticalOnTopBorder, H: InputHorizontalCenter}, nil
	case "on-bottom":
		return InputPlacement{V: InputVerticalOnBottomBorder, H: InputHorizontalCenter}, nil
	case "on-top-dynamic":
		return InputPlacement{V: InputVerticalOnTopBorderDynamic, H: InputHorizontalCenter}, nil
	case "on-bottom-dynamic":
		return InputPlacement{V: InputVerticalOnBottomBorderDynamic, H: InputHorizontalCenter}, nil
	}
	if p, ok := inputPositionShorthand[s]; ok {
		return p, nil
	}
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return InputPlacement{}, fmt.Errorf(`input position: %q must be shorthand (tl, bc, …) or "top-left" form`, s)
	}
	var p InputPlacement
	switch parts[0] {
	case "top":
		p.V = InputVerticalTop
	case "bottom":
		p.V = InputVerticalBottom
	default:
		return InputPlacement{}, fmt.Errorf(`input position: unknown vertical %q (want top or bottom)`, parts[0])
	}
	switch parts[1] {
	case "left":
		p.H = InputHorizontalLeft
	case "center":
		p.H = InputHorizontalCenter
	case "right":
		p.H = InputHorizontalRight
	default:
		return InputPlacement{}, fmt.Errorf(`input position: unknown horizontal %q (want left, center, or right)`, parts[1])
	}
	return p, nil
}
