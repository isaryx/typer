package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"typer/internal/model"
	"typer/internal/session"
	"typer/internal/storage"
	"typer/internal/text"
	"typer/internal/ui"
)

func runStart(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var mode string
	fs.StringVar(&mode, "mode", "quotes", "passages|p, words|w, quotes|q.")
	fs.StringVar(&mode, "m", "quotes", "Shorthand for --mode.")
	var words int
	fs.IntVar(&words, "words", 15, "Words per prompt (words mode).")
	fs.IntVar(&words, "w", 15, "Shorthand for --words.")
	var source string
	fs.StringVar(&source, "source", "remote", "Quotes: remote|cache|seed.")
	var quoteSourcesRaw string
	fs.StringVar(&quoteSourcesRaw, "quote-sources", "", "Quotes only: comma-separated remote IDs for this session (e.g. zenquotes,typefit). Overrides saved toggles.")

	var noGhost bool
	fs.BoolVar(&noGhost, "no-ghost", false, "Skip ghost overlay from best prior run.")
	var noInput bool
	fs.BoolVar(&noInput, "no-input", false, "Hide input line.")
	var noAudible bool
	fs.BoolVar(&noAudible, "no-audible", false, "Disable terminal bell on mistakes.")

	rest, strict, indefinite, fingerHint, err := extractPresenceFlags(args)
	if err != nil {
		return err
	}
	if err := fs.Parse(rest); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printStartHelp(stdout)
			return nil
		}
		return err
	}
	if err := rejectExtraArgs("start", fs.Args()); err != nil {
		return err
	}

	canonMode, err := text.CanonicalMode(strings.TrimSpace(mode))
	if err != nil {
		return err
	}

	sourceProvided := false
	quoteSourcesProvided := false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "source":
			sourceProvided = true
		case "quote-sources":
			quoteSourcesProvided = true
		}
	})
	if sourceProvided && canonMode != model.ModeQuote {
		return fmt.Errorf("--source is only valid with --mode quotes")
	}
	if quoteSourcesProvided && canonMode != model.ModeQuote {
		return fmt.Errorf("--quote-sources is only valid with --mode quotes")
	}

	opts := model.SessionOptions{
		Mode:       canonMode,
		Words:      words,
		Source:     source,
		Strict:     strict,
		Indefinite: indefinite,
		FingerHint: fingerHint,
		NoInput:    noInput,
		NoAudible:  noAudible,
	}
	if err := model.ValidateSessionOptions(opts); err != nil {
		return err
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
	opts.HideHint = !settings.HintVisible()
	opts.InputPlacement = settings.InputPlacement()

	var quoteCfg text.QuoteProviderConfig
	if canonMode == model.ModeQuote {
		allowlist, err := parseCommaQuoteSources(quoteSourcesRaw)
		if err != nil {
			return err
		}
		quoteCfg.EnabledRemoteIDs = text.ResolveEnabledQuoteRemotes(settings.QuoteRemoteEnabled, allowlist)
		opts.RemoteQuoteFetchSplash = text.QuoteModeMayBlockOnNetwork(canonMode, source) &&
			len(quoteCfg.EnabledRemoteIDs) > 0
	}

	provider, err := text.NewProvider(canonMode, cache, settings.WordsFile, settings.PassagesFile, quoteCfg)
	if err != nil {
		return err
	}

	runner := session.NewRunner(provider)
	historyStore, err := storage.NewHistoryStore()
	if err != nil {
		return err
	}
	if !noGhost {
		hst := historyStore
		runner.GhostBaseline = func(ctx context.Context, p model.Prompt) (*model.SessionResult, error) {
			h := model.PromptContentHash(p.Content)
			best, err := hst.BestSessionForGhost(h)
			if err != nil {
				if errors.Is(err, storage.ErrNoGhostCandidate) {
					return nil, nil
				}
				return nil, err
			}
			b := best
			return &b, nil
		}
	}

	var results []model.SessionResult
	round := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		result, err := runner.Run(ctx, opts, stdin, stdout, nil)
		if err != nil {
			return err
		}
		if !result.Aborted {
			if err := historyStore.Append(result); err != nil {
				return err
			}
		}
		round++
		results = append(results, result)
		if !indefinite {
			break
		}
		if result.Aborted {
			break
		}
		fmt.Fprintf(stdout, "\n── Round %d complete — next session ──\n\n", round)
	}

	printStartResults(stdout, results)
	return nil
}

func printStartResults(out io.Writer, results []model.SessionResult) {
	var finished []model.SessionResult
	for _, r := range results {
		if !r.Aborted {
			finished = append(finished, r)
		}
	}
	if len(finished) == 0 {
		fmt.Fprintln(out, "\nSession aborted.")
		return
	}
	if len(finished) == 1 {
		printOneSessionResult(out, finished[0])
		if len(results) > 0 && results[len(results)-1].Aborted {
			fmt.Fprintln(out, "\nSession aborted.")
		}
		return
	}
	var sumGross, sumNet, sumAdjusted, sumAcc, sumCons float64
	var sumElapsed int64
	sumErr := 0
	for _, r := range finished {
		sumGross += r.Metrics.GrossWPM
		sumNet += r.Metrics.NetWPM
		sumAdjusted += r.Metrics.AdjustedWPM
		sumAcc += r.Metrics.Accuracy
		sumCons += r.Metrics.Consistency
		sumElapsed += r.ElapsedMS
		sumErr += r.Metrics.Errors
	}
	n := float64(len(finished))
	avgGross := sumGross / n
	avgNet := sumNet / n
	avgAdjusted := sumAdjusted / n
	avgAcc := sumAcc / n
	avgCons := sumCons / n
	printMetricsTable(out, fmt.Sprintf("Summary (%d completed sessions)", len(finished)), avgGross, avgNet, avgAdjusted, avgAcc, avgCons, sumErr, sumElapsed, true)
	if len(results) > 0 && results[len(results)-1].Aborted {
		fmt.Fprintln(out, "\nSession aborted.")
	}
}

func printOneSessionResult(out io.Writer, r model.SessionResult) {
	m := r.Metrics
	printMetricsTable(out, "Session Stats", m.GrossWPM, m.NetWPM, m.AdjustedWPM, m.Accuracy, m.Consistency, m.Errors, r.ElapsedMS, false)
}

func printMetricsTable(out io.Writer, heading string, gross, net, adjusted, acc, cons float64, errCount int, elapsedMS int64, summary bool) {
	ui.PrintMetricsTable(out, heading, gross, net, adjusted, acc, cons, errCount, elapsedMS, summary)
}

// parseCommaQuoteSources splits a comma-separated list of remote registry IDs.
// Empty input returns (nil, nil) meaning "use settings only" (no session allowlist).
func parseCommaQuoteSources(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		id := strings.ToLower(strings.TrimSpace(p))
		if id == "" {
			continue
		}
		if !text.IsKnownQuoteRemoteID(id) {
			return nil, fmt.Errorf("unknown quote remote %q (known: %s)", id, strings.Join(text.KnownQuoteRemoteIDs(), ", "))
		}
		out = append(out, id)
	}
	return out, nil
}
