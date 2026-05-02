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

// Shorthand: tl, tc, tr, bl, bc, br (top/bottom × left/center/right).
var inputPositionShorthand = map[string]InputPlacement{
	"tl": {V: InputVerticalTop, H: InputHorizontalLeft},
	"tc": {V: InputVerticalTop, H: InputHorizontalCenter},
	"tr": {V: InputVerticalTop, H: InputHorizontalRight},
	"bl": {V: InputVerticalBottom, H: InputHorizontalLeft},
	"bc": {V: InputVerticalBottom, H: InputHorizontalCenter},
	"br": {V: InputVerticalBottom, H: InputHorizontalRight},
}

// ParseInputPosition parses values like "top-left", "bottom-center", or shorthand "tl", "bc".
func ParseInputPosition(s string) (InputPlacement, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return InputPlacement{}, fmt.Errorf(`input position: empty (expected e.g. "bottom-left", "bc", "top-center")`)
	}
	if len(s) == 2 {
		if p, ok := inputPositionShorthand[s]; ok {
			return p, nil
		}
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
