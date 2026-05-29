package defense

// Difficulty ramps on two axes: fall speed and spawn rate scale with elapsed survival
// time (see EffectiveFallSpeed / EffectiveSpawnInterval); max word length scales with
// words destroyed (see MaxWordLenForScore).

import "time"

// EffectiveFallMultiplier returns the speed multiplier from elapsed survival time.
func EffectiveFallMultiplier(elapsed time.Duration) float64 {
	sec := elapsed.Seconds()
	ramp := 1.0 + (sec/fallRampSeconds)*fallRampBonus
	if ramp > maxFallMultiplier {
		return maxFallMultiplier
	}
	return ramp
}

// EffectiveFallSpeed returns rows per second after time scaling.
func EffectiveFallSpeed(base float64, elapsed time.Duration) float64 {
	return base * EffectiveFallMultiplier(elapsed)
}

// EffectiveSpawnMultiplier returns spawn-rate multiplier (lower interval = faster spawns).
func EffectiveSpawnMultiplier(elapsed time.Duration) float64 {
	sec := elapsed.Seconds()
	ramp := 1.0 + (sec/spawnRampInterval)*spawnRampBonus
	if ramp > maxSpawnMultiplier {
		return maxSpawnMultiplier
	}
	return ramp
}

// EffectiveSpawnInterval returns average seconds between spawns after time scaling.
func EffectiveSpawnInterval(baseSeconds float64, elapsed time.Duration) float64 {
	return baseSeconds / EffectiveSpawnMultiplier(elapsed)
}
