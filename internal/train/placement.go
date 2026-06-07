package train

import (
	"typer/internal/model"
	"typer/internal/scoring"
)

// PlacementInput aggregates metrics from placement test segments.
type PlacementInput struct {
	Segments []model.SessionResult
}

// AssignPlacement computes tier and starting lesson from combined placement metrics.
func AssignPlacement(in PlacementInput) PlacementResult {
	if len(in.Segments) == 0 {
		return PlacementResult{
			AssignedTier:   TierFoundation,
			AssignedLesson: "1.1",
		}
	}

	var sumNet, sumAcc float64
	for _, s := range in.Segments {
		sumNet += s.Metrics.NetWPM
		sumAcc += SegmentAccuracy(s)
	}
	n := float64(len(in.Segments))
	netWPM := scoring.Round2(sumNet / n)
	accuracy := scoring.Round2(sumAcc / n)

	tier, lesson := placementTierLesson(netWPM, accuracy)
	return PlacementResult{
		NetWPM:         netWPM,
		Accuracy:       accuracy,
		AssignedTier:   tier,
		AssignedLesson: lesson,
	}
}

func placementTierLesson(netWPM, accuracy float64) (tier, lesson string) {
	switch {
	case netWPM < 20 || accuracy < 85:
		return TierFoundation, "1.1"
	case netWPM < 35 || accuracy < 92:
		if netWPM >= 30 && accuracy >= 90 {
			return TierBuilding, "2.1"
		}
		return TierFoundation, "1.5"
	case netWPM < 50 || accuracy < 95:
		return TierBuilding, "2.4"
	default:
		return TierFluent, "3.1"
	}
}

// SegmentAccuracy uses keystroke accuracy for strict placement segments, else final text accuracy.
func SegmentAccuracy(s model.SessionResult) float64 {
	if s.Options.Strict {
		if ks := KeystrokeAccuracy(s); ks >= 0 {
			return ks
		}
	}
	return s.Metrics.Accuracy
}
