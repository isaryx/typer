package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"typer/internal/analytics"
	"typer/internal/model"
	"typer/internal/session"
	"typer/internal/storage"
	"typer/internal/text"
	"typer/internal/version"
)

const noHistoryMessage = "No history yet. Run `typer start` first."

const resetProgressFlag = "--reset-progress"

const creditsMessage = `Data Credits
------------
- Words mode default list (assets/words.txt): first20hours/google-10000-english
- Original corpus source: Google Web Trillion Word Corpus (via LDC), cleaned by Josh Kaufman; subsets by Peter Norvig
- Passages mode bundled list (assets/passages.txt): Thomas Preston's Dictionary of English Proverbs and Proverbial Phrases (Project Gutenberg #39281)
- Quotes mode bundled list (assets/quotes.json): curated from dwyl/quotes
- Optional remote quote API source (when --source=remote): https://type.fit/api/quotes

See README "Data Credits" for details and usage notes.
`

var newHistoryStore = storage.NewHistoryStore

// extractPresenceFlags removes bare strict / indefinite tokens from args
// (presence only). Names with '=' or followed by a boolean-looking token
// are rejected.
func extractPresenceFlags(args []string) (rest []string, strict, indefinite bool, err error) {
	strictNames := map[string]struct{}{
		"--strict": {}, "-strict": {}, "-s": {},
	}
	indefNames := map[string]struct{}{
		"--indefinite": {}, "-indefinite": {}, "-i": {},
	}
	for i := 0; i < len(args); i++ {
		a := args[i]
		name, _, foundEq := strings.Cut(a, "=")
		if foundEq {
			if _, ok := strictNames[name]; ok {
				return nil, false, false, fmt.Errorf("invalid %s: use bare --strict, -strict, or -s for strict mode, or omit for non-strict", a)
			}
			if _, ok := indefNames[name]; ok {
				return nil, false, false, fmt.Errorf("invalid %s: use bare --indefinite or -i for indefinite mode, or omit", a)
			}
			rest = append(rest, a)
			continue
		}
		if _, ok := strictNames[a]; ok {
			if err := rejectBoolLiteralAfterPresenceFlag(args, i, a); err != nil {
				return nil, false, false, err
			}
			strict = true
			continue
		}
		if _, ok := indefNames[a]; ok {
			if err := rejectBoolLiteralAfterPresenceFlag(args, i, a); err != nil {
				return nil, false, false, err
			}
			indefinite = true
			continue
		}
		rest = append(rest, a)
	}
	return rest, strict, indefinite, nil
}

func rejectBoolLiteralAfterPresenceFlag(args []string, i int, flagToken string) error {
	if i+1 >= len(args) {
		return nil
	}
	next := args[i+1]
	if strings.HasPrefix(next, "-") {
		return nil
	}
	if _, perr := strconv.ParseBool(next); perr == nil {
		return fmt.Errorf("invalid: do not pass %q after %s; use the flag alone, or omit", next, flagToken)
	}
	return nil
}

func Execute(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printHelp(stdout)
		return nil
	}

	switch args[0] {
	case "start":
		return runStart(ctx, args[1:], stdin, stdout)
	case "set":
		return runSet(args[1:], stdout)
	case "history":
		return runHistory(args[1:], stdout)
	case "stats":
		return runStats(args[1:], stdout)
	case "credits":
		return runCredits(stdout)
	case "version", "--version", "-v":
		return runVersion(stdout)
	case resetProgressFlag:
		return runResetProgress(args[1:], stdin, stdout, stderr)
	case "help", "--help", "-h":
		printHelp(stdout)
		return nil
	default:
		// Shorthand: `typer -m w -w 5` is the same as `typer start -m w -w 5`.
		if strings.HasPrefix(args[0], "-") {
			return runStart(ctx, args, stdin, stdout)
		}
		printHelp(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "typer %s\n", version.Version)
	return err
}

func runCredits(out io.Writer) error {
	_, err := fmt.Fprint(out, creditsMessage)
	return err
}

func readResetConfirm(stdin io.Reader) (ok bool, err error) {
	r := bufio.NewReader(stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && line == "" {
			return false, nil
		}
		if errors.Is(err, io.EOF) {
			// If we got a partial line without a trailing newline, still interpret it.
			return parseResetYes(strings.TrimSpace(line)), nil
		}
		return false, err
	}
	return parseResetYes(strings.TrimSpace(line)), nil
}

func parseResetYes(s string) bool {
	switch strings.ToLower(s) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

func runResetProgress(args []string, stdin io.Reader, out, outErr io.Writer) error {
	if len(args) > 0 {
		return fmt.Errorf("%s does not take additional arguments", resetProgressFlag)
	}
	fmt.Fprint(outErr, "Reset all local session history? [y/N]: ")
	ok, err := readResetConfirm(stdin)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(outErr, "Reset cancelled.")
		return nil
	}
	store, err := newHistoryStore()
	if err != nil {
		return err
	}
	if err := store.Reset(); err != nil {
		return err
	}
	fmt.Fprintln(out, "Progress reset. Local session history is now empty.")
	return nil
}

func runStart(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var mode string
	fs.StringVar(&mode, "mode", "quotes", "Text mode: passages|p, words|w, or quotes|q.")
	fs.StringVar(&mode, "m", "quotes", "Shorthand for --mode (same values as --mode).")
	var words int
	fs.IntVar(&words, "words", 15, "Number of words per prompt (words mode only).")
	fs.IntVar(&words, "w", 15, "Shorthand for --words.")
	var source string
	fs.StringVar(&source, "source", "seed", "Quotes mode fetch policy: auto|remote|cache|seed.")

	rest, strict, indefinite, err := extractPresenceFlags(args)
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

	canonMode, err := text.CanonicalMode(strings.TrimSpace(mode))
	if err != nil {
		return err
	}

	sourceProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "source" {
			sourceProvided = true
		}
	})
	if sourceProvided && canonMode != model.ModeQuote {
		return fmt.Errorf("--source is only valid with --mode quotes")
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
	provider, err := text.NewProvider(canonMode, cache, settings.WordsFile, settings.PassagesFile)
	if err != nil {
		return err
	}

	runner := session.NewRunner(provider)
	historyStore, err := storage.NewHistoryStore()
	if err != nil {
		return err
	}

	opts := model.SessionOptions{
		Mode:       canonMode,
		Words:      words,
		Source:     source,
		Strict:     strict,
		Indefinite: indefinite,
	}

	var results []model.SessionResult
	round := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		result, err := runner.Run(ctx, opts, stdin, stdout)
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
		fmt.Fprintln(out)
		fmt.Fprintln(out, "No completed sessions.")
		return
	}
	if len(finished) == 1 {
		printOneSessionResult(out, finished[0])
		return
	}
	var sumGross, sumNet, sumAcc, sumCons float64
	var sumElapsed int64
	sumErr := 0
	for _, r := range finished {
		sumGross += r.Metrics.GrossWPM
		sumNet += r.Metrics.NetWPM
		sumAcc += r.Metrics.Accuracy
		sumCons += r.Metrics.Consistency
		sumElapsed += r.ElapsedMS
		sumErr += r.Metrics.Errors
	}
	n := float64(len(finished))
	avgGross := sumGross / n
	avgNet := sumNet / n
	avgAcc := sumAcc / n
	avgCons := sumCons / n
	printMetricsTable(out, fmt.Sprintf("Summary (%d completed sessions)", len(finished)), avgGross, avgNet, avgAcc, avgCons, sumErr, sumElapsed, true)
}

func printOneSessionResult(out io.Writer, r model.SessionResult) {
	m := r.Metrics
	printMetricsTable(out, "Session result", m.GrossWPM, m.NetWPM, m.Accuracy, m.Consistency, m.Errors, r.ElapsedMS, false)
}

func printMetricsTable(out io.Writer, heading string, gross, net, acc, cons float64, errCount int, elapsedMS int64, summary bool) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  · "+heading)
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	printRow := func(label, value string) {
		_, _ = fmt.Fprintln(tw, "  "+label+"\t"+value)
	}
	if summary {
		printRow("Avg gross WPM", fmt.Sprintf("%.2f", gross))
		printRow("Avg net WPM", fmt.Sprintf("%.2f", net))
		printRow("Avg accuracy", fmt.Sprintf("%.2f%%", acc))
		printRow("Avg consistency", fmt.Sprintf("%.2f", cons))
		printRow("Total errors", fmt.Sprintf("%d", errCount))
		printRow("Total time", formatElapsedMS(elapsedMS))
	} else {
		printRow("Gross WPM", fmt.Sprintf("%.2f", gross))
		printRow("Net WPM", fmt.Sprintf("%.2f", net))
		printRow("Accuracy", fmt.Sprintf("%.2f%%", acc))
		if cons > 0 {
			printRow("Consistency", fmt.Sprintf("%.2f", cons))
		}
		printRow("Errors", fmt.Sprintf("%d", errCount))
		printRow("Time", formatElapsedMS(elapsedMS))
	}
	_ = tw.Flush()
	fmt.Fprintln(out)
}

func formatElapsedMS(ms int64) string {
	if ms < 0 {
		return "0 ms"
	}
	if ms < 1000 {
		return fmt.Sprintf("%d ms", ms)
	}
	secs := ms / 1000
	if secs < 60 {
		if ms < 10_000 {
			return fmt.Sprintf("%.1f s", float64(ms)/1000)
		}
		return fmt.Sprintf("%d s", secs)
	}
	mins := secs / 60
	secs %= 60
	if mins < 60 {
		return fmt.Sprintf("%dm %02ds", mins, secs)
	}
	h := mins / 60
	mins %= 60
	return fmt.Sprintf("%dh %02dm %02ds", h, mins, secs)
}

func runSet(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("set", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	wordsFile := fs.String("words-file", "", "Path to a newline-separated custom word list.")
	passagesFile := fs.String("passages-file", "", "Path to a blank-line-separated custom passages file.")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printSetHelp(stdout)
			return nil
		}
		return err
	}

	wf := strings.TrimSpace(*wordsFile)
	pf := strings.TrimSpace(*passagesFile)
	if wf == "" && pf == "" {
		return errors.New("set requires --words-file and/or --passages-file")
	}

	validateFile := func(label, path string) (string, error) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		info, err := os.Stat(absPath)
		if err != nil {
			return "", fmt.Errorf("invalid %s %q: %w", label, absPath, err)
		}
		if info.IsDir() {
			return "", fmt.Errorf("invalid %s %q: expected file, got directory", label, absPath)
		}
		return absPath, nil
	}

	settingsStore, err := storage.NewSettingsStore()
	if err != nil {
		return err
	}
	settings, err := settingsStore.Load()
	if err != nil {
		return err
	}

	if wf != "" {
		absPath, err := validateFile("words file", wf)
		if err != nil {
			return err
		}
		settings.WordsFile = absPath
	}
	if pf != "" {
		absPath, err := validateFile("passages file", pf)
		if err != nil {
			return err
		}
		settings.PassagesFile = absPath
	}

	if err := settingsStore.Save(settings); err != nil {
		return err
	}

	if wf != "" {
		fmt.Fprintf(stdout, "Custom words file set to %s\n", settings.WordsFile)
	}
	if pf != "" {
		fmt.Fprintf(stdout, "Custom passages file set to %s\n", settings.PassagesFile)
	}
	return nil
}

func runHistory(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("history", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	last := fs.Int("last", 20, "Maximum number of recent sessions to list.")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printHistoryHelp(stdout)
			return nil
		}
		return err
	}

	store, err := storage.NewHistoryStore()
	if err != nil {
		return err
	}
	results, err := store.List(*last)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		fmt.Fprintln(stdout, noHistoryMessage)
		return nil
	}

	fmt.Fprintf(stdout, "Showing %d session(s).\n", len(results))
	for i, r := range results {
		fmt.Fprintf(
			stdout,
			"%2d) %s  mode=%s net=%.2f acc=%.2f%% errors=%d source=%s\n",
			i+1,
			r.StartedAt.Local().Format("2006-01-02 15:04:05"),
			r.Prompt.Mode,
			r.Metrics.NetWPM,
			r.Metrics.Accuracy,
			r.Metrics.Errors,
			r.Prompt.Source,
		)
	}
	return nil
}

func runStats(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("stats", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	last := fs.Int("last", 20, "Maximum number of recent sessions to analyze.")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printStatsHelp(stdout)
			return nil
		}
		return err
	}

	store, err := storage.NewHistoryStore()
	if err != nil {
		return err
	}
	results, err := store.List(*last)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		fmt.Fprintln(stdout, noHistoryMessage)
		return nil
	}

	sum := analytics.BuildSummary(results, 5)
	fmt.Fprintf(stdout, "Stats for last %d session(s).\n", sum.Sessions)
	fmt.Fprintf(stdout, "Avg Gross WPM     : %.2f\n", sum.AvgGrossWPM)
	fmt.Fprintf(stdout, "Avg Net WPM       : %.2f\n", sum.AvgNetWPM)
	fmt.Fprintf(stdout, "Avg Accuracy      : %.2f%%\n", sum.AvgAccuracy)
	fmt.Fprintf(stdout, "Avg Errors        : %.2f\n", sum.AvgErrors)
	fmt.Fprintf(stdout, "Consistency Trend : %.2f\n", sum.ConsistencyTrend)
	printCharCounts(stdout, "Top Mistyped Target Characters", sum.TopErrorChars)
	printCharCounts(stdout, "Top Missing Target Characters", sum.TopMissingChars)
	printCharCounts(stdout, "Top Unexpected Typed Characters", sum.TopUnexpectedChar)
	return nil
}

func printCharCounts(out io.Writer, title string, counts []analytics.CharCount) {
	fmt.Fprintf(out, "%s:\n", title)
	if len(counts) == 0 {
		fmt.Fprintln(out, "  - none")
		return
	}
	for _, c := range counts {
		fmt.Fprintf(out, "  - %-12s %d\n", c.Key, c.Count)
	}
}

func printHelp(out io.Writer) {
	fmt.Fprintf(out, "typer %s — terminal typing practice\n", version.Version)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer <command> [arguments]")
	fmt.Fprintln(out, "  typer --reset-progress       # clear saved history (type y/yes to confirm)")
	fmt.Fprintln(out, "  typer [arguments]              # same as: typer start [arguments] when the first token begins with \"-\"")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Commands:")
	fmt.Fprintln(out, "  start        Run an interactive typing session.")
	fmt.Fprintln(out, "  set          Save custom words and/or passages file paths.")
	fmt.Fprintln(out, "  history      List recent sessions from local history.")
	fmt.Fprintln(out, "  stats        Summarize recent sessions.")
	fmt.Fprintln(out, "  credits      Show data source credits.")
	fmt.Fprintln(out, "  version      Print the installed typer version.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Quick reference:")
	fmt.Fprintln(out, "  typer start [--mode|-m MODE] [--words|-w N] [--source SRC]")
	fmt.Fprintln(out, "              [--strict|-s] [--indefinite|-i]")
	fmt.Fprintln(out, "  typer set [--words-file PATH] [--passages-file PATH]")
	fmt.Fprintln(out, "  typer history [--last N]")
	fmt.Fprintln(out, "  typer stats [--last N]")
	fmt.Fprintln(out, "  typer credits")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Per-command help:")
	fmt.Fprintln(out, "  typer start -h | typer start --help")
	fmt.Fprintln(out, "  typer set -h   | typer set --help")
	fmt.Fprintln(out, "  typer history -h | typer stats -h")
	fmt.Fprintln(out, "  typer --reset-progress         # type y/yes to confirm on the prompt")
}

func printStartHelp(out io.Writer) {
	fmt.Fprintln(out, "Run an interactive typing session.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer start [--mode|-m MODE] [--words|-w N] [--source SRC]")
	fmt.Fprintln(out, "              [--strict|-s] [--indefinite|-i]")
	fmt.Fprintln(out, "  typer [same flags...]            # omit the word \"start\" when the first argument is a flag")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Options:")
	fmt.Fprintln(out, "  -m, --mode string    Text mode: passages|p, words|w, or quotes|q (default quotes).")
	fmt.Fprintln(out, "  -w, --words int      Words per prompt in words mode (default 15).")
	fmt.Fprintln(out, "      --source string  Quotes mode only: auto|remote|cache|seed (default seed).")
	fmt.Fprintln(out, "  -s, --strict         Strict matching: wrong input does not advance (bare flag; omit for non-strict).")
	fmt.Fprintln(out, "  -i, --indefinite     After each finished session, start another until Ctrl+C or Esc (bare flag).")
}

func printSetHelp(out io.Writer) {
	fmt.Fprintln(out, "Save paths to custom corpus files (at least one flag is required).")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer set [--words-file PATH] [--passages-file PATH]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Options:")
	fmt.Fprintln(out, "      --words-file PATH     Newline-separated word list.")
	fmt.Fprintln(out, "      --passages-file PATH  Blank-line-separated passage blocks.")
}

func printHistoryHelp(out io.Writer) {
	fmt.Fprintln(out, "List recent typing sessions from local history.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer history [--last N]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Options:")
	fmt.Fprintln(out, "      --last int   Maximum sessions to show (default 20).")
}

func printStatsHelp(out io.Writer) {
	fmt.Fprintln(out, "Print aggregate statistics over recent sessions.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer stats [--last N]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Options:")
	fmt.Fprintln(out, "      --last int   Maximum sessions to analyze (default 20).")
}
