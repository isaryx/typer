package model

import "fmt"

const (
	// MaxWordsPerPrompt caps words-mode prompts to avoid accidental huge sessions (e.g. typos).
	MaxWordsPerPrompt = 999
	// MaxRetainedHistorySessions caps history.json size (storage) and the largest sensible --last query.
	MaxRetainedHistorySessions = 5000
)

// ValidateSessionOptions checks bounds that apply before starting a session (CLI or programmatic).
// Words==0 is allowed (providers treat non-positive values as their default).
func ValidateSessionOptions(o SessionOptions) error {
	if o.Words < 0 {
		return fmt.Errorf("words per prompt cannot be negative (got %d)", o.Words)
	}
	if o.Words > MaxWordsPerPrompt {
		return fmt.Errorf("words per prompt cannot exceed %d (got %d)", MaxWordsPerPrompt, o.Words)
	}
	if o.DurationMS < 0 {
		return fmt.Errorf("duration cannot be negative (got %d)", o.DurationMS)
	}
	return nil
}

// ValidateHistoryLast checks --last for history and stats commands.
func ValidateHistoryLast(n int) error {
	if n < 1 {
		return fmt.Errorf("--last must be at least 1")
	}
	if n > MaxRetainedHistorySessions {
		return fmt.Errorf("--last cannot exceed %d", MaxRetainedHistorySessions)
	}
	return nil
}
