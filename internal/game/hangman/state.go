package hangman

import "fmt"

// Config holds hangman game parameters.
type Config struct {
	MaxStrikes        int
	MistakesPerStrike int
}

// ValidateConfig checks v1 constraints.
func ValidateConfig(c Config) error {
	if c.MaxStrikes != DefaultMaxStrikes {
		return fmt.Errorf("--strikes must be %d in this version (got %d)", DefaultMaxStrikes, c.MaxStrikes)
	}
	if c.MistakesPerStrike < 1 {
		return fmt.Errorf("--mistakes-per-strike must be at least 1 (got %d)", c.MistakesPerStrike)
	}
	return nil
}

// DefaultConfig returns standard hangman settings.
func DefaultConfig() Config {
	return Config{
		MaxStrikes:        DefaultMaxStrikes,
		MistakesPerStrike: 1,
	}
}

// State tracks mistake-driven gallows progression during a session.
type State struct {
	MaxStrikes        int
	MistakesPerStrike int
	mistakeCount      int
	stage             int
}

// NewState builds hangman progress tracking from config.
func NewState(c Config) *State {
	return &State{
		MaxStrikes:        c.MaxStrikes,
		MistakesPerStrike: c.MistakesPerStrike,
	}
}

// Stage returns the current gallows stage (0..MaxStrikes).
func (s *State) Stage() int {
	return s.stage
}

// MistakeCount returns total mistake keystrokes recorded.
func (s *State) MistakeCount() int {
	return s.mistakeCount
}

// IsLost reports whether the gallows has reached the final stage.
func (s *State) IsLost() bool {
	return s.stage >= s.MaxStrikes
}

// RecordMistake increments the mistake counter and advances the gallows when enough
// mistakes have accumulated. Returns true when the player loses.
func (s *State) RecordMistake() bool {
	s.mistakeCount++
	if s.MistakesPerStrike <= 0 {
		s.MistakesPerStrike = 1
	}
	if s.mistakeCount%s.MistakesPerStrike != 0 {
		return false
	}
	s.stage++
	return s.stage >= s.MaxStrikes
}
