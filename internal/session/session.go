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
}

func NewRunner(provider text.Provider) *Runner {
	return &Runner{
		Provider: provider,
		Now:      time.Now,
	}
}

// replayBaseline, when non-nil, enables replay UI (previous run metrics) for the same prompt.
func (r *Runner) Run(ctx context.Context, opts model.SessionOptions, input io.Reader, output io.Writer, replayBaseline *model.SessionResult) (model.SessionResult, error) {
	prompt, err := r.Provider.Next(ctx, text.Constraints{
		Words:  opts.Words,
		Source: opts.Source,
	})
	if err != nil {
		return model.SessionResult{}, err
	}

	tuiResult, err := runTypingSession(ctx, input, output, prompt, opts.Strict, opts.Indefinite, r.Now, replayBaseline)
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
		ID:        newSessionID(startedAt),
		StartedAt: startedAt.UTC(),
		EndedAt:   endedAt.UTC(),
		ElapsedMS: endedAt.Sub(startedAt).Milliseconds(),
		Prompt:    prompt,
		TypedText: typed,
		Metrics:   metrics,
		Aborted:   tuiResult.Aborted,

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
