package cli

import (
	"context"
	"io"

	"typer/internal/model"
	"typer/internal/session"
	"typer/internal/storage"
)

// applySessionDisplayFromSettings sets HideHint and InputPlacement from saved app settings.
func applySessionDisplayFromSettings(opts *model.SessionOptions, s storage.AppSettings) {
	opts.HideHint = !s.HintVisible()
	opts.InputPlacement = s.InputPlacement()
}

// runSessionAndPersist runs a typing session and appends the result to history when it completed (not aborted).
func runSessionAndPersist(
	ctx context.Context,
	runner *session.Runner,
	opts model.SessionOptions,
	stdin io.Reader,
	stdout io.Writer,
	replayBaseline *model.SessionResult,
	history *storage.HistoryStore,
) (model.SessionResult, error) {
	result, err := runner.Run(ctx, opts, stdin, stdout, replayBaseline)
	if err != nil {
		return model.SessionResult{}, err
	}
	if !result.Aborted {
		if err := history.Append(result); err != nil {
			return model.SessionResult{}, err
		}
	}
	return result, nil
}
