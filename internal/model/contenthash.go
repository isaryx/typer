package model

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// CanonicalPromptContent returns a stable form of prompt text for hashing:
//   - CRLF and lone CR become LF
//   - each line is trimmed and internal runs of spaces/tabs collapse to one space
//   - whole string leading/trailing whitespace removed (after line join)
//   - Unicode NFC normalization
func CanonicalPromptContent(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(squashSpaceAndTab(line))
	}
	s = strings.TrimSpace(strings.Join(lines, "\n"))
	return norm.NFC.String(s)
}

// squashSpaceAndTab replaces each run of spaces and tabs with a single space.
func squashSpaceAndTab(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inWS := false
	for _, r := range s {
		if r == ' ' || r == '\t' {
			inWS = true
			continue
		}
		if inWS && b.Len() > 0 {
			b.WriteByte(' ')
		}
		inWS = false
		b.WriteRune(r)
	}
	return b.String()
}

// PromptContentHash is SHA-256 hex (64 chars) of UTF-8 canonical prompt content.
func PromptContentHash(raw string) string {
	sum := sha256.Sum256([]byte(CanonicalPromptContent(raw)))
	return hex.EncodeToString(sum[:])
}

// SessionContentHashKey returns persisted ContentHash when set; otherwise
// recomputes from Prompt.Content so legacy sessions match by text.
func SessionContentHashKey(r SessionResult) string {
	if r.ContentHash != "" {
		return r.ContentHash
	}
	return PromptContentHash(r.Prompt.Content)
}
