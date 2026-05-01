package session

import "testing"

func TestFingerForRune_homeRow(t *testing.T) {
	tests := []struct {
		r    rune
		want Finger
	}{
		{'a', FingerLeftPinky},
		{'s', FingerLeftRing},
		{'d', FingerLeftMiddle},
		{'f', FingerLeftIndex},
		{'g', FingerLeftIndex},
		{'h', FingerRightIndex},
		{'j', FingerRightIndex},
		{'k', FingerRightMiddle},
		{'l', FingerRightRing},
		{'A', FingerLeftPinky},
	}
	for _, tt := range tests {
		if got := FingerForRune(tt.r); got != tt.want {
			t.Errorf("FingerForRune(%q) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestFingerForRune_topRowLetters(t *testing.T) {
	tests := []struct {
		r    rune
		want Finger
	}{
		{'q', FingerLeftPinky},
		{'w', FingerLeftRing},
		{'e', FingerLeftMiddle},
		{'r', FingerLeftIndex},
		{'t', FingerLeftIndex},
		{'y', FingerRightIndex},
		{'u', FingerRightIndex},
		{'i', FingerRightMiddle},
		{'o', FingerRightRing},
		{'p', FingerRightPinky},
	}
	for _, tt := range tests {
		if got := FingerForRune(tt.r); got != tt.want {
			t.Errorf("FingerForRune(%q) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestFingerForRune_digits(t *testing.T) {
	tests := []struct {
		r    rune
		want Finger
	}{
		{'1', FingerLeftPinky},
		{'2', FingerLeftRing},
		{'3', FingerLeftMiddle},
		{'4', FingerLeftIndex},
		{'5', FingerLeftIndex},
		{'6', FingerRightIndex},
		{'7', FingerRightIndex},
		{'8', FingerRightMiddle},
		{'9', FingerRightRing},
		{'0', FingerRightPinky},
	}
	for _, tt := range tests {
		if got := FingerForRune(tt.r); got != tt.want {
			t.Errorf("FingerForRune(%q) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestFingerForRune_thumbsSpace(t *testing.T) {
	for _, r := range []rune{' ', '\t', '\n'} {
		if got := FingerForRune(r); got != FingerBothThumbs {
			t.Errorf("FingerForRune(%q) = %v, want FingerBothThumbs", r, got)
		}
	}
}

func TestFingerForRune_unknown(t *testing.T) {
	if got := FingerForRune('€'); got != FingerUnknown {
		t.Fatalf("FingerForRune('€') = %v, want FingerUnknown", got)
	}
}
