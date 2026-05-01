package session

import "unicode"

// US QWERTY touch-typing finger zones (for expectedKeystroke runes) follow common
// teaching layouts; spot-check with a reference such as
// https://www.rapidtyping.com/typing-instructions.html

// Finger identifies a finger on a standard US QWERTY keyboard for touch typing.
// Slots in the hands UI (left-to-right) map to:
// Lp Lr Lm Li Lt | Rt Ri Rm Rr Rp.
type Finger int

const (
	FingerUnknown Finger = iota
	FingerLeftPinky
	FingerLeftRing
	FingerLeftMiddle
	FingerLeftIndex
	FingerBothThumbs // space / enter — highlight Lt and Rt labels
	FingerRightIndex
	FingerRightMiddle
	FingerRightRing
	FingerRightPinky
)

// FingerForRune maps a keystroke to a finger using conventional US QWERTY zones
// (letters case-insensitive; punctuation matches unshifted/shifted pairs on the US layout).
func FingerForRune(r rune) Finger {
	if r > unicode.MaxASCII {
		return FingerUnknown
	}
	r = unicode.ToLower(r)
	switch r {
	case ' ', '\t':
		return FingerBothThumbs
	case '\n', '\r':
		return FingerBothThumbs
	case '`', '~':
		return FingerLeftPinky
	case '1', '!':
		return FingerLeftPinky
	case '2', '@':
		return FingerLeftRing
	case '3', '#':
		return FingerLeftMiddle
	case '4', '$', '5', '%':
		return FingerLeftIndex
	case '6', '^', '7', '&':
		return FingerRightIndex
	case '8', '*':
		return FingerRightMiddle
	case '9', '(':
		return FingerRightRing
	case '0', ')', '-', '_', '=', '+':
		return FingerRightPinky
	case 'q', 'a', 'z':
		return FingerLeftPinky
	case 'w', 's', 'x':
		return FingerLeftRing
	case 'e', 'd', 'c':
		return FingerLeftMiddle
	case 'r', 'f', 'v', 't', 'g', 'b':
		return FingerLeftIndex
	case 'y', 'u', 'h', 'j', 'n', 'm':
		return FingerRightIndex
	case 'i', 'k', ',', '<':
		return FingerRightMiddle
	case 'o', 'l', '.', '>':
		return FingerRightRing
	case 'p', ';', ':', '\'', '"', '[', '{', ']', '}', '\\', '|', '/', '?':
		return FingerRightPinky
	default:
		return FingerUnknown
	}
}
