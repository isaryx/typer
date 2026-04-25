package scoring

import "math"

// ConsistencyFromSamples calculates how stable speed is between time slices.
// 100 means perfectly stable; lower values indicate bigger speed swings.
func ConsistencyFromSamples(wpmSamples []float64) float64 {
	if len(wpmSamples) <= 1 {
		return 100
	}
	mean := 0.0
	for _, s := range wpmSamples {
		mean += s
	}
	mean /= float64(len(wpmSamples))
	if mean <= 0 {
		return 0
	}

	variance := 0.0
	for _, s := range wpmSamples {
		diff := s - mean
		variance += diff * diff
	}
	variance /= float64(len(wpmSamples))

	stddev := math.Sqrt(variance)
	coeffVar := stddev / mean
	score := 100 * (1 - coeffVar)
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return round2(score)
}
