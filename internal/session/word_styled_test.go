package session

import (
	"strings"
	"testing"
)

func TestRenderActiveWordTextMatchesSessionSegments(t *testing.T) {
	styles := DefaultStyles()
	got := RenderActiveWordText(styles, "python", "py")
	if got == "" {
		t.Fatal("empty render")
	}
	// Styled output should contain the plain letters.
	if !strings.Contains(stripANSI(got), "python") {
		t.Fatalf("missing letters in %q", got)
	}
	if !strings.Contains(stripANSI(got), "py") {
		t.Fatalf("missing typed prefix in %q", got)
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	in := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			in = true
			continue
		}
		if in {
			if s[i] == 'm' {
				in = false
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
