package session

import (
	"crypto/rand"
	"time"
	"unicode/utf8"

	"github.com/oklog/ulid/v2"
)

// newSessionID returns a sortable unique id; the ULID time field matches startedAt.
// On entropy or ULID construction failure, falls back to RFC3339Nano (legacy).
func newSessionID(startedAt time.Time) string {
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}
	id, err := ulid.New(ulid.Timestamp(startedAt), ulid.Monotonic(rand.Reader, 0))
	if err != nil {
		return startedAt.UTC().Format(time.RFC3339Nano)
	}
	return id.String()
}

// formatSessionIDForDisplay formats a session id for UI when StartedAt is missing
// (e.g. partial data). Recognizes legacy RFC3339 ids and ULID strings.
func formatSessionIDForDisplay(id string) string {
	if id == "" {
		return ""
	}
	if t, err := time.Parse(time.RFC3339Nano, id); err == nil {
		return t.Local().Format("2006-01-02 15:04:05")
	}
	if t, err := time.Parse(time.RFC3339, id); err == nil {
		return t.Local().Format("2006-01-02 15:04:05")
	}
	if u, err := ulid.ParseStrict(id); err == nil {
		return u.Timestamp().Local().Format("2006-01-02 15:04:05")
	}
	if utf8.RuneCountInString(id) > 36 {
		r := []rune(id)
		return string(r[:33]) + "…"
	}
	return id
}
