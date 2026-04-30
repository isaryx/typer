package session

import (
	"crypto/rand"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/oklog/ulid/v2"
)

func TestNewSessionIDIsULIDAlignedWithStartedAt(t *testing.T) {
	start := time.Date(2024, 3, 15, 9, 30, 0, 123456789, time.UTC)
	id := newSessionID(start)
	u, err := ulid.ParseStrict(id)
	if err != nil {
		t.Fatalf("newSessionID: %v", err)
	}
	if u.Timestamp().UnixMilli() != start.UTC().UnixMilli() {
		t.Fatalf("ULID ms %d != startedAt ms %d", u.Timestamp().UnixMilli(), start.UTC().UnixMilli())
	}
}

func TestFormatSessionIDForDisplay(t *testing.T) {
	t.Setenv("TZ", "UTC")
	if got, want := formatSessionIDForDisplay("2026-04-30T12:04:05.123456789Z"), "2026-04-30 12:04:05"; got != want {
		t.Fatalf("RFC3339Nano: got %q want %q", got, want)
	}
	if got, want := formatSessionIDForDisplay("2026-04-30T12:04:05Z"), "2026-04-30 12:04:05"; got != want {
		t.Fatalf("RFC3339: got %q want %q", got, want)
	}
	long := strings.Repeat("x", 40)
	if got := formatSessionIDForDisplay(long); !strings.HasSuffix(got, "…") || utf8.RuneCountInString(got) != 34 {
		t.Fatalf("truncate: %q (runes=%d)", got, utf8.RuneCountInString(got))
	}
	ms := ulid.Timestamp(time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC))
	id, err := ulid.New(ms, ulid.Monotonic(rand.Reader, 0))
	if err != nil {
		t.Fatal(err)
	}
	if got := formatSessionIDForDisplay(id.String()); got != "2026-05-01 10:00:00" {
		t.Fatalf("ULID id display: got %q", got)
	}
}
