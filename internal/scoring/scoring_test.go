package scoring

import (
	"testing"
	"time"
)

func TestCompute_PerfectTyping(t *testing.T) {
	target := "hello world"
	typed := "hello world"
	metrics := Compute(target, typed, 30*time.Second, Keystrokes{
		Total:             11,
		Correct:           11,
		UncorrectedErrors: 0,
	})

	if metrics.Errors != 0 {
		t.Fatalf("expected 0 errors, got %d", metrics.Errors)
	}
	if metrics.Accuracy != 100 {
		t.Fatalf("expected 100 accuracy, got %.2f", metrics.Accuracy)
	}
	if metrics.NetWPM <= 0 {
		t.Fatalf("expected positive net wpm, got %.2f", metrics.NetWPM)
	}
	if metrics.NetWPM != metrics.GrossWPM {
		t.Fatalf("expected net == gross when no errors, got gross=%.2f net=%.2f", metrics.GrossWPM, metrics.NetWPM)
	}
}

func TestCompute_WithUncorrectedErrors(t *testing.T) {
	target := "the quick brown fox"
	typed := "the quack brown fix"
	metrics := Compute(target, typed, 1*time.Minute, Keystrokes{
		Total:             19,
		Correct:           17,
		UncorrectedErrors: 2,
	})

	if metrics.Errors != 2 {
		t.Fatalf("expected 2 uncorrected errors, got %d", metrics.Errors)
	}
	if metrics.Accuracy >= 100 {
		t.Fatalf("expected accuracy under 100, got %.2f", metrics.Accuracy)
	}
	if metrics.NetWPM > metrics.GrossWPM {
		t.Fatalf("net wpm should not exceed gross wpm")
	}
}

func TestCompute_KeystrokeAccuracy(t *testing.T) {
	metrics := Compute("hello", "hxllo", time.Minute, Keystrokes{
		Total:             5,
		Correct:           4,
		UncorrectedErrors: 1,
	})
	if metrics.Accuracy != 80 {
		t.Fatalf("expected 80%% accuracy, got %.2f", metrics.Accuracy)
	}
}

func TestCompute_FallbackWithoutKeystrokes(t *testing.T) {
	metrics := Compute("hello", "helXo", time.Minute, Keystrokes{})
	if metrics.Accuracy == 100 {
		t.Fatalf("expected accuracy to drop without full match, got %.2f", metrics.Accuracy)
	}
	if metrics.Errors == 0 {
		t.Fatalf("expected fallback to detect errors")
	}
}

func TestConsistencyFromSamples(t *testing.T) {
	stable := ConsistencyFromSamples([]float64{60, 60, 60, 60})
	unstable := ConsistencyFromSamples([]float64{20, 80, 30, 90})

	if stable < unstable {
		t.Fatalf("expected stable consistency > unstable consistency")
	}
}
