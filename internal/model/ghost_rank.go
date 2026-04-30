package model

// BetterGhostCandidate reports whether a should rank above b as a ghost replay source.
// Ordering: typing trace preferred; non-aborted over aborted; then higher net WPM,
// higher accuracy, fewer errors, shorter elapsed time, newer StartedAt.
func BetterGhostCandidate(a, b SessionResult) bool {
	aTrace := len(a.TypingTrace) > 0
	bTrace := len(b.TypingTrace) > 0
	if aTrace != bTrace {
		return aTrace
	}
	if a.Aborted != b.Aborted {
		return !a.Aborted && b.Aborted
	}
	if a.Metrics.NetWPM != b.Metrics.NetWPM {
		return a.Metrics.NetWPM > b.Metrics.NetWPM
	}
	if a.Metrics.Accuracy != b.Metrics.Accuracy {
		return a.Metrics.Accuracy > b.Metrics.Accuracy
	}
	if a.Metrics.Errors != b.Metrics.Errors {
		return a.Metrics.Errors < b.Metrics.Errors
	}
	if a.ElapsedMS != b.ElapsedMS {
		return a.ElapsedMS < b.ElapsedMS
	}
	return a.StartedAt.After(b.StartedAt)
}
