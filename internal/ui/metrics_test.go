package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatElapsedMS(t *testing.T) {
	tests := []struct {
		ms   int64
		want string
	}{
		{-1, "0 ms"},
		{0, "0 ms"},
		{500, "500 ms"},
		{999, "999 ms"},
		{1000, "1.0 s"},
		{5500, "5.5 s"},
		{10_000, "10 s"},
		{59_000, "59 s"},
		{60_000, "1m 00s"},
		{61_000, "1m 01s"},
		{3599_000, "59m 59s"},
		{3600_000, "1h 00m 00s"},
	}
	for _, tc := range tests {
		if got := FormatElapsedMS(tc.ms); got != tc.want {
			t.Errorf("FormatElapsedMS(%d) = %q, want %q", tc.ms, got, tc.want)
		}
	}
}

func TestPrintMetricsTablePlainSession(t *testing.T) {
	var buf bytes.Buffer
	PrintMetricsTable(&buf, "Session Stats", 72, 68, 64, 88.9, 91.2, 3, 27_000, false)
	got := buf.String()
	for _, needle := range []string{
		"Session Stats",
		"Gross WPM",
		"Net WPM",
		"Adjusted WPM",
		"88.90%",
		"27 s",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("missing %q in:\n%s", needle, got)
		}
	}
}

func TestPrintMetricsTablePlainSummary(t *testing.T) {
	var buf bytes.Buffer
	PrintMetricsTable(&buf, "Summary (2 completed sessions)", 70, 66, 62, 90, 85, 5, 120_000, true)
	got := buf.String()
	for _, needle := range []string{
		"Summary (2 completed sessions)",
		"Avg gross WPM",
		"Total time:",
		"2m 00s",
	} {
		if !strings.Contains(got, needle) {
			t.Fatalf("missing %q in:\n%s", needle, got)
		}
	}
}

func TestMetricsInnerWidthMatchesFrameBody(t *testing.T) {
	const w = 100
	if got := metricsInnerWidth(w); got != FrameBodyInnerWidth(w) {
		t.Fatalf("metricsInnerWidth(%d) = %d, FrameBodyInnerWidth = %d", w, got, FrameBodyInnerWidth(w))
	}
}
