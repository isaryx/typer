package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"

	"typer/internal/model"
	"typer/internal/session"
	"typer/internal/storage"
	"typer/internal/text"
	"typer/internal/ui"
)

// Replay comparison line colors (shared TUI palette).
var (
	replayDeltaGood = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorTitle))
	replayDeltaBad  = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorCompletedBad))
)

func runReplay(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("replay", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.String("id", "", "Session id from typer history.")
	nth := fs.Int("nth", 0, "Replay Nth newest (1 = newest).")
	var last bool
	fs.BoolVar(&last, "last", false, "Newest session.")
	fs.BoolVar(&last, "l", false, "Shorthand for --last.")
	var noInput bool
	fs.BoolVar(&noInput, "no-input", false, "Hide input line.")
	var noAudible bool
	fs.BoolVar(&noAudible, "no-audible", false, "Disable terminal bell on mistakes.")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printReplayHelp(stdout)
			return nil
		}
		return usageErrf("%v", err)
	}
	if err := rejectExtraArgs("replay", fs.Args()); err != nil {
		return err
	}

	var useID string
	var nSelect int
	switch {
	case last:
		nSelect = 1
	case *nth > 0:
		nSelect = *nth
	case strings.TrimSpace(*id) != "":
		useID = strings.TrimSpace(*id)
	default:
		return usageErrf("replay requires -l/--last, --nth N, or --id ID (see: typer replay -h)")
	}

	store, err := newHistoryStore()
	if err != nil {
		return err
	}
	settingsStore, err := storage.NewSettingsStore()
	if err != nil {
		return err
	}
	settings, err := settingsStore.Load()
	if err != nil {
		return err
	}
	var baseline model.SessionResult
	if useID != "" {
		baseline, err = store.GetByID(useID)
	} else {
		baseline, err = store.NthNewest(nSelect)
	}
	if err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) || errors.Is(err, storage.ErrInsufficientHistory) || errors.Is(err, storage.ErrNthOutOfRange) {
			return usageErrf("%v", err)
		}
		return err
	}
	if len(strings.Fields(baseline.Prompt.Content)) == 0 {
		return errors.New("cannot replay: stored prompt has no words")
	}

	if msg := typingTraceUserNote(baseline); msg != "" {
		fmt.Fprintln(stdout, msg)
	}

	runner := session.NewRunner(text.NewStaticProvider(baseline.Prompt))
	opts := model.SessionOptionsForReplay(baseline)
	opts.NoInput = noInput
	opts.NoAudible = noAudible
	applySessionDisplayFromSettings(&opts, settings)

	result, err := runSessionAndPersist(ctx, runner, opts, stdin, stdout, &baseline, store)
	if err != nil {
		return err
	}
	if result.Aborted {
		fmt.Fprintln(stdout, "\nReplay aborted.")
		return nil
	}
	printStartResults(stdout, []model.SessionResult{result})
	printReplayComparison(stdout, baseline, result)
	return nil
}

func printReplayComparison(out io.Writer, previous, current model.SessionResult) {
	if current.Aborted {
		return
	}
	dNet := current.Metrics.NetWPM - previous.Metrics.NetWPM
	dAcc := current.Metrics.Accuracy - previous.Metrics.Accuracy
	dErr := current.Metrics.Errors - previous.Metrics.Errors
	deltaMS := current.ElapsedMS - previous.ElapsedMS
	fmt.Fprintln(out, "Comparison:")
	fmt.Fprintf(out, "  Net WPM:  %s (%.2f →  %.2f)\n",
		styleHigherBetterDelta(dNet).Render(fmt.Sprintf("%+.2f", dNet)),
		previous.Metrics.NetWPM, current.Metrics.NetWPM)
	fmt.Fprintf(out, "  Accuracy: %s (%.2f%% →  %.2f%%)\n",
		styleHigherBetterDelta(dAcc).Render(fmt.Sprintf("%+.2f%%", dAcc)),
		previous.Metrics.Accuracy, current.Metrics.Accuracy)
	fmt.Fprintf(out, "  Errors:   %s (%d →  %d)\n",
		styleFewerErrorsBetterDelta(dErr).Render(fmt.Sprintf("%+d", dErr)),
		previous.Metrics.Errors, current.Metrics.Errors)
	switch {
	case deltaMS < 0:
		timePhrase := replayDeltaGood.Render(fmt.Sprintf("%s faster", ui.FormatElapsedMS(-deltaMS)))
		fmt.Fprintf(out, "  Time:     %s  (%s →  %s)\n", timePhrase, ui.FormatElapsedMS(previous.ElapsedMS), ui.FormatElapsedMS(current.ElapsedMS))
	case deltaMS > 0:
		timePhrase := replayDeltaBad.Render(fmt.Sprintf("%s slower", ui.FormatElapsedMS(deltaMS)))
		fmt.Fprintf(out, "  Time:     %s  (%s →  %s)\n", timePhrase, ui.FormatElapsedMS(previous.ElapsedMS), ui.FormatElapsedMS(current.ElapsedMS))
	default:
		fmt.Fprintf(out, "  Time:     same  (%s)\n", ui.FormatElapsedMS(current.ElapsedMS))
	}
}

func styleHigherBetterDelta(delta float64) lipgloss.Style {
	switch {
	case delta > 0:
		return replayDeltaGood
	case delta < 0:
		return replayDeltaBad
	default:
		return lipgloss.NewStyle()
	}
}

func styleFewerErrorsBetterDelta(delta int) lipgloss.Style {
	switch {
	case delta < 0:
		return replayDeltaGood
	case delta > 0:
		return replayDeltaBad
	default:
		return lipgloss.NewStyle()
	}
}

// typingTraceUserNote is a one-line notice when the baseline has no stored keystroke trace
// (older sessions or future options to omit traces), so the TUI will not show a shadow.
func typingTraceUserNote(baseline model.SessionResult) string {
	if len(baseline.TypingTrace) == 0 {
		return "Note: This session has no typing trace; shadow replay is unavailable."
	}
	return ""
}
