package defense

import "time"

const (
	DefaultLives         = 6
	DefaultSpawnSeconds  = 1.2
	DefaultBaseFallSpeed = 0.5 // rows per second
	MaxConcurrentWords      = 8
	PlayfieldRows           = 19
	MinTerminalWidth        = 60
	MinTerminalHeight       = 26
	MinWordLen              = 3
	MaxWordLen              = 12
	TickInterval            = 16 * time.Millisecond

	fallRampSeconds   = 120.0
	fallRampBonus     = 0.5
	maxFallMultiplier = 2.0

	spawnRampInterval  = 60.0
	spawnRampBonus    = 0.05
	maxSpawnMultiplier = 1.5

	startMaxWordLen   = 4 // 3–4 letter words at game start
	wordsPerLenTier   = 10 // +1 max length every N words destroyed
)

// Config holds defense game parameters.
type Config struct {
	Lives            int
	BaseSpawnSeconds float64
	BaseFallSpeed    float64
	NoAudible        bool
}

// DefaultConfig returns standard defense settings.
func DefaultConfig() Config {
	return Config{
		Lives:            DefaultLives,
		BaseSpawnSeconds: DefaultSpawnSeconds,
		BaseFallSpeed:    DefaultBaseFallSpeed,
	}
}
