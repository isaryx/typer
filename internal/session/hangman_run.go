package session

import (
	"context"
	"io"
	"time"

	"typer/internal/game/hangman"
	"typer/internal/model"
	"typer/internal/scoring"
	"typer/internal/version"
)

// HangmanRunOpts configures a hangman typing session.
type HangmanRunOpts struct {
	Hangman        hangman.Config
	NoInput        bool
	HideHint       bool
	InputPlacement model.InputPlacement
	NoAudible      bool
}

// HangmanOutcome summarizes a hangman session for callers outside the session package.
type HangmanOutcome struct {
	Result string // "win", "lose", or "" when aborted
	Stage  int
}

// RunHangman runs a quote-typing session with the hangman gallows overlay.
func RunHangman(ctx context.Context, input io.Reader, output io.Writer, prompt model.Prompt, o HangmanRunOpts, now func() time.Time) (model.SessionResult, HangmanOutcome, error) {
	if err := hangman.ValidateConfig(o.Hangman); err != nil {
		return model.SessionResult{}, HangmanOutcome{}, err
	}
	if now == nil {
		now = time.Now
	}

	prompt.Mode = model.ModeHangman
	hm := hangman.NewState(o.Hangman)

	tuiResult, err := runTypingSession(ctx, input, output, prompt, typingSessionRunOpts{
		strict:         false,
		now:            now,
		noInput:        o.NoInput,
		hideHint:       o.HideHint,
		inputPlacement: o.InputPlacement,
		noAudible:      o.NoAudible,
		hangman:        hm,
	})
	if err != nil {
		return model.SessionResult{}, HangmanOutcome{}, err
	}

	metrics := scoring.Compute(prompt.Content, tuiResult.TypedText, tuiResult.EndedAt.Sub(tuiResult.StartedAt), scoring.Keystrokes{
		Total:             tuiResult.TotalKeystrokes,
		Correct:           tuiResult.CorrectKeystrokes,
		UncorrectedErrors: tuiResult.UncorrectedErrors,
	})
	metrics.Consistency = scoring.ConsistencyFromSamples(tuiResult.WPMSamples)

	result := model.SessionResult{
		ID:          newSessionID(tuiResult.StartedAt),
		ContentHash: model.PromptContentHash(prompt.Content),
		StartedAt:   tuiResult.StartedAt.UTC(),
		EndedAt:     tuiResult.EndedAt.UTC(),
		ElapsedMS:   tuiResult.EndedAt.Sub(tuiResult.StartedAt).Milliseconds(),
		Prompt:      prompt,
		TypedText:   tuiResult.TypedText,
		Metrics:     metrics,
		Aborted:     tuiResult.Aborted,

		ResultSchema:      model.SessionResultSchema,
		TyperVersion:      version.Version,
		Options:           model.OptionsSnapshot(model.SessionOptions{Mode: model.ModeHangman}),
		TotalKeystrokes:   tuiResult.TotalKeystrokes,
		CorrectKeystrokes: tuiResult.CorrectKeystrokes,
		WPMSamples:        append([]float64(nil), tuiResult.WPMSamples...),
		TypingTrace:       append([]model.ReplayEvent(nil), tuiResult.TypingTrace...),
	}
	outcome := HangmanOutcome{
		Result: tuiResult.HangmanOutcome,
		Stage:  tuiResult.HangmanStage,
	}
	return result, outcome, nil
}
