package scoring

import (
	"math"
	"strings"
	"time"

	"typer/internal/model"
)

// Keystrokes captures a session's raw counts for gross WPM and net penalty.
// Total is every character typed (including corrections); Correct is recorded for diagnostics
// but Compute derives Accuracy from final target vs typed text.
// UncorrectedErrors is the sum of per-word CompareRunes mismatch weights when words are committed.
type Keystrokes struct {
	Total             int
	Correct           int
	UncorrectedErrors int
}

// Compute returns metrics using the industry-standard formulas:
//   - Gross WPM = (totalKeystrokes / 5) / minutes
//   - Net WPM   = Gross - (uncorrectedErrors / minutes)
//   - Accuracy  = 100 × textCorrect / (textCorrect + textErr), comparing typed text to
//     the prompt normalized like the typing UI (strings.Fields joined by spaces). Raw
//     newlines or multiple spaces in target otherwise disagree with joined typed words.
//
// If ks.Total is zero, gross WPM and uncorrectedErrors for net WPM fall back to comparing
// full target vs typed text so callers without keystroke counts still get sensible metrics.
func Compute(target, typed string, elapsed time.Duration, ks Keystrokes) model.SessionMetrics {
	typedRunes := []rune(typed)
	normTarget := strings.Join(strings.Fields(target), " ")
	normTargetRunes := []rune(normTarget)

	minutes := elapsed.Minutes()
	if minutes <= 0 {
		minutes = 1.0 / 60.0
	}

	totalK := ks.Total
	uncorrected := ks.UncorrectedErrors

	if totalK == 0 {
		totalK = len(typedRunes)
		_, uncorrected = CompareRunes(normTargetRunes, typedRunes)
	}

	gross := (float64(totalK) / 5.0) / minutes
	net := gross - (float64(uncorrected) / minutes)
	if net < 0 {
		net = 0
	}

	textCorrect, textErr := CompareRunes(normTargetRunes, typedRunes)
	accuracy := 100.0
	if d := textCorrect + textErr; d > 0 {
		accuracy = 100 * (float64(textCorrect) / float64(d))
	}
	adjusted := gross * (accuracy / 100.0)

	return model.SessionMetrics{
		GrossWPM:    Round2(gross),
		NetWPM:      Round2(net),
		AdjustedWPM: Round2(adjusted),
		Accuracy:    Round2(accuracy),
		Consistency: 0,
		Errors:      uncorrected,
	}
}

// CompareRunes returns (correctCount, uncorrectedErrorCount) when comparing
// typed runes to the target. Uncorrected errors include substitutions, extras,
// and missing trailing runes.
func CompareRunes(target, typed []rune) (int, int) {
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

// Round2 rounds to two decimals (banker-agnostic, uses math.Round).
func Round2(v float64) float64 {
	return math.Round(v*100) / 100
}
