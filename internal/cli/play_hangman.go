package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"typer/internal/game/hangman"
	"typer/internal/model"
	"typer/internal/session"
	"typer/internal/storage"
	"typer/internal/text"
	"typer/internal/ui"
)

func runPlayHangman(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("play hangman", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var source string
	fs.StringVar(&source, "source", "remote", "Quotes: remote|cache|seed.")
	var quoteSourcesRaw string
	fs.StringVar(&quoteSourcesRaw, "quote-sources", "", "Comma-separated remote IDs (zenquotes,typefit).")

	var strikes int
	fs.IntVar(&strikes, "strikes", hangman.DefaultMaxStrikes, "Gallows stages before loss (must be 6).")
	var mistakesPerStrike int
	fs.IntVar(&mistakesPerStrike, "mistakes-per-strike", 1, "Mistake keystrokes per gallows stage.")

	var noAudible bool
	fs.BoolVar(&noAudible, "no-audible", false, "Disable terminal bell on mistakes.")
	var noInput bool
	fs.BoolVar(&noInput, "no-input", false, "Hide input line.")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printPlayHangmanHelp(stdout)
			return nil
		}
		return usageErrf("%v", err)
	}
	if err := rejectExtraArgs("play hangman", fs.Args()); err != nil {
		return err
	}

	hmCfg := hangman.Config{
		MaxStrikes:        strikes,
		MistakesPerStrike: mistakesPerStrike,
	}
	if err := hangman.ValidateConfig(hmCfg); err != nil {
		return usageErrf("%v", err)
	}

	cache, err := storage.NewQuoteCacheStore()
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

	allowlist, err := parseCommaQuoteSources(quoteSourcesRaw)
	if err != nil {
		return err
	}
	quoteCfg := text.QuoteProviderConfig{
		EnabledRemoteIDs: text.ResolveEnabledQuoteRemotes(settings.QuoteRemoteEnabled, allowlist),
	}
	provider, err := text.NewProvider(model.ModeQuote, cache, settings.WordsFile, settings.PassagesFile, quoteCfg)
	if err != nil {
		return err
	}

	remoteSplash := text.QuoteModeMayBlockOnNetwork(model.ModeQuote, source) && len(quoteCfg.EnabledRemoteIDs) > 0
	prompt, err := session.NextQuotePrompt(ctx, stdout, func() (model.Prompt, error) {
		return provider.Next(ctx, text.Constraints{Source: source})
	}, remoteSplash)
	if err != nil {
		return err
	}

	display := model.SessionOptions{}
	applySessionDisplayFromSettings(&display, settings)

	result, outcome, err := session.RunHangman(ctx, stdin, stdout, prompt, session.HangmanRunOpts{
		Hangman:        hmCfg,
		NoInput:        noInput,
		HideHint:       display.HideHint,
		InputPlacement: display.InputPlacement,
		NoAudible:      noAudible,
	}, nil)
	if err != nil {
		return err
	}

	if !result.Aborted {
		historyStore, err := storage.NewHistoryStore()
		if err != nil {
			return err
		}
		if err := historyStore.Append(result); err != nil {
			return err
		}
	}

	printHangmanOutcome(stdout, result, outcome)
	return nil
}

func printHangmanOutcome(out io.Writer, r model.SessionResult, outcome session.HangmanOutcome) {
	fmt.Fprintln(out)
	switch outcome.Result {
	case "win":
		m := r.Metrics
		ui.PrintMetricsTable(out, "Session Stats", m.GrossWPM, m.NetWPM, m.AdjustedWPM, m.Accuracy, m.Consistency, m.Errors, r.ElapsedMS, false)
		fmt.Fprintln(out, "\nYou survived!")
	case "lose":
		fmt.Fprintf(out, "Hangman complete — game over (stage %d/%d).\n", outcome.Stage, hangman.DefaultMaxStrikes)
	default:
		if r.Aborted {
			fmt.Fprintln(out, "Game aborted.")
		}
	}
}

func printPlayHangmanHelp(out io.Writer) {
	fmt.Fprintln(out, "Hangman — type a quote; mistake keystrokes draw the gallows.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer play hangman [--source SRC] [--quote-sources IDS]")
	fmt.Fprintln(out, "                     [--strikes 6] [--mistakes-per-strike N]")
	fmt.Fprintln(out, "                     [--no-input] [--no-audible]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Flags:")
	fmt.Fprintln(out, "  --source string           remote|cache|seed (default remote).")
	fmt.Fprintln(out, "  --quote-sources string    Comma-separated remote IDs (zenquotes, typefit).")
	fmt.Fprintln(out, "  --strikes int             Gallows stages before loss (must be 6).")
	fmt.Fprintln(out, "  --mistakes-per-strike int Mistake keys per stage (default 1).")
	fmt.Fprintln(out, "  --no-input                Hide the input line.")
	fmt.Fprintln(out, "  --no-audible              Disable terminal bell on mistakes.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintln(out, "  typer play hangman")
	fmt.Fprintln(out, "  typer play hangman --mistakes-per-strike 3 --source seed")
}
