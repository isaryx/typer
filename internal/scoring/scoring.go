package scoring

import (
	"math"
	"time"

	"typer/internal/model"
)

// Keystrokes captures a session's raw and correctness counts.
// Total represents every character the typist pressed (including ones later corrected).
// Correct represents keystrokes that matched the expected character at time of press.
// UncorrectedErrors represents final mismatches still present when a word was committed.
type Keystrokes struct {
	Total             int
	Correct           int
	UncorrectedErrors int
}

// Compute returns metrics using the industry-standard formulas:
//   - Gross WPM = (totalKeystrokes / 5) / minutes
//   - Net WPM   = Gross - (uncorrectedErrors / minutes)
//   - Accuracy  = correctKeystrokes / totalKeystrokes * 100
//
// If ks.Total is zero, the function falls back to the typed text string, which keeps
// simpler tests and historic call sites functional.
func Compute(target, typed string, elapsed time.Duration, ks Keystrokes) model.SessionMetrics {
	typedRunes := []rune(typed)

	minutes := elapsed.Minutes()
	if minutes <= 0 {
		minutes = 1.0 / 60.0
	}

	total := ks.Total
	correct := ks.Correct
	uncorrected := ks.UncorrectedErrors

	if total == 0 {
		total = len(typedRunes)
		correct, uncorrected = compareRunes([]rune(target), typedRunes)
	}

	gross := (float64(total) / 5.0) / minutes
	net := gross - (float64(uncorrected) / minutes)
	if net < 0 {
		net = 0
	}

	accuracy := 0.0
	if total > 0 {
		accuracy = (float64(correct) / float64(total)) * 100.0
	}

	return model.SessionMetrics{
		GrossWPM:    round2(gross),
		NetWPM:      round2(net),
		Accuracy:    round2(accuracy),
		Consistency: 0,
		Errors:      uncorrected,
	}
}

func compareRunes(target, typed []rune) (int, int) {
	n := min(len(target), len(typed))
	correct := 0
	for i := 0; i < n; i++ {
		if target[i] == typed[i] {
			correct++
		}
	}
	uncorrected := (len(typed) - correct) + max(0, len(target)-len(typed))
	return correct, uncorrected
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
