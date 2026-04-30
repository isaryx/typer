package scoring

import (
	"strings"
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
	if metrics.AdjustedWPM != metrics.GrossWPM {
		t.Fatalf("expected adjusted == gross at 100%% accuracy, got gross=%.2f adjusted=%.2f", metrics.GrossWPM, metrics.AdjustedWPM)
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
	if metrics.AdjustedWPM >= metrics.GrossWPM {
		t.Fatalf("expected adjusted wpm below gross when accuracy < 100, got gross=%.2f adjusted=%.2f", metrics.GrossWPM, metrics.AdjustedWPM)
	}
}

func TestCompute_KeystrokeAccuracy(t *testing.T) {
	metrics := Compute("hello", "hxllo", time.Minute, Keystrokes{
		Total:             5,
		Correct:           4,
		UncorrectedErrors: 1,
	})
	if metrics.Accuracy != 80 {
		t.Fatalf("expected 80%% text accuracy, got %.2f", metrics.Accuracy)
	}
}

func TestCompute_TextAccuracyIncompleteDoesNotShow100(t *testing.T) {
	// Final text shorter than target: every typed key matched at cursor, but passage incomplete.
	m := Compute("hello", "hell", time.Minute, Keystrokes{
		Total:             4,
		Correct:           4,
		UncorrectedErrors: 1,
	})
	if m.Accuracy != 80 {
		t.Fatalf("Accuracy = %.2f, want 80 (not 100 despite keystroke hits)", m.Accuracy)
	}
}

func TestCompute_AccuracyMatchesFieldsJoinWhenPromptHasNewlines(t *testing.T) {
	// TUI uses strings.Fields; typed text is strings.Join(typedWords, " ").
	// Newlines in the stored prompt must not make accuracy < 100% when errors are 0.
	target := "Your vision will become clear only when you can look into your own heart.\n\nWho looks outside, dreams; who looks inside, awakes."
	typed := strings.Join(strings.Fields(target), " ")
	m := Compute(target, typed, 30*time.Second, Keystrokes{
		Total:             len([]rune(typed)),
		Correct:           len([]rune(typed)),
		UncorrectedErrors: 0,
	})
	if m.Accuracy != 100 {
		t.Fatalf("Accuracy = %.2f, want 100 when normalized text matches typed", m.Accuracy)
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
