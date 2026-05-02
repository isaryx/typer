package session

import (
	"context"
	"io"
	"time"

	"typer/internal/model"
	"typer/internal/scoring"
	"typer/internal/text"
	"typer/internal/version"
)

type Runner struct {
	Provider text.Provider
	Now      func() time.Time
	// GhostBaseline, if set, runs after the prompt is chosen and may return a session to use as replay ghost baseline (nil = none).
	GhostBaseline func(ctx context.Context, prompt model.Prompt) (*model.SessionResult, error)
}

func NewRunner(provider text.Provider) *Runner {
	return &Runner{
		Provider: provider,
		Now:      time.Now,
	}
}

// replayBaseline, when non-nil, comes from `typer replay` and enables replay header UI plus shadow trace.
// GhostBaseline runs only when replayBaseline is nil; it supplies shadow trace without replay chrome.
func (r *Runner) Run(ctx context.Context, opts model.SessionOptions, input io.Reader, output io.Writer, replayBaseline *model.SessionResult) (model.SessionResult, error) {
	if err := model.ValidateSessionOptions(opts); err != nil {
		return model.SessionResult{}, err
	}

	prompt, err := r.Provider.Next(ctx, text.Constraints{
		Words:  opts.Words,
		Source: opts.Source,
	})
	if err != nil {
		return model.SessionResult{}, err
	}

	showReplayUI := replayBaseline != nil
	baseline := replayBaseline
	if baseline == nil && r.GhostBaseline != nil {
		b, err := r.GhostBaseline(ctx, prompt)
		if err != nil {
			return model.SessionResult{}, err
		}
		baseline = b
	}

	tuiResult, err := runTypingSession(ctx, input, output, prompt, opts.Strict, opts.Indefinite, r.Now, baseline, showReplayUI, opts.FingerHint, opts.NoInput, opts.HideHint, opts.InputPlacement)
	if err != nil {
		return model.SessionResult{}, err
	}
	startedAt := tuiResult.StartedAt
	endedAt := tuiResult.EndedAt
	typed := tuiResult.TypedText

	metrics := scoring.Compute(prompt.Content, typed, endedAt.Sub(startedAt), scoring.Keystrokes{
		Total:             tuiResult.TotalKeystrokes,
		Correct:           tuiResult.CorrectKeystrokes,
		UncorrectedErrors: tuiResult.UncorrectedErrors,
	})
	metrics.Consistency = scoring.ConsistencyFromSamples(tuiResult.WPMSamples)

	wpmSamples := append([]float64(nil), tuiResult.WPMSamples...)
	typingTrace := append([]model.ReplayEvent(nil), tuiResult.TypingTrace...)

	result := model.SessionResult{
		ID:          newSessionID(startedAt),
		ContentHash: model.PromptContentHash(prompt.Content),
		StartedAt:   startedAt.UTC(),
		EndedAt:     endedAt.UTC(),
		ElapsedMS:   endedAt.Sub(startedAt).Milliseconds(),
		Prompt:      prompt,
		TypedText:   typed,
		Metrics:     metrics,
		Aborted:     tuiResult.Aborted,

		ResultSchema:      model.SessionResultSchema,
		TyperVersion:      version.Version,
		Options:           model.OptionsSnapshot(opts),
		TotalKeystrokes:   tuiResult.TotalKeystrokes,
		CorrectKeystrokes: tuiResult.CorrectKeystrokes,
		WPMSamples:        wpmSamples,
		TypingTrace:       typingTrace,
	}
	return result, nil
}
