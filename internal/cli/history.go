package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"typer/internal/analytics"
	"typer/internal/model"
	"typer/internal/storage"
)

const noHistoryMessage = "No history yet. Run `typer start` first."

func runHistory(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("history", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	last := fs.Int("last", 20, "How many recent sessions to list (default 20).")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printHistoryHelp(stdout)
			return nil
		}
		return usageErrf("%v", err)
	}
	if err := rejectExtraArgs("history", fs.Args()); err != nil {
		return err
	}
	if err := model.ValidateHistoryLast(*last); err != nil {
		return usageErrf("%v", err)
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
			"%2d) %s  mode=%s net=%.2f acc=%.2f%% errors=%d source=%s  id=%s\n",
			i+1,
			r.StartedAt.Local().Format("2006-01-02 15:04:05"),
			r.Prompt.Mode,
			r.Metrics.NetWPM,
			r.Metrics.Accuracy,
			r.Metrics.Errors,
			r.Prompt.Source,
			r.ID,
		)
	}
	return nil
}

func runStats(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("stats", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	last := fs.Int("last", 20, "How many sessions to include (default 20).")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printStatsHelp(stdout)
			return nil
		}
		return usageErrf("%v", err)
	}
	if err := rejectExtraArgs("stats", fs.Args()); err != nil {
		return err
	}
	if err := model.ValidateHistoryLast(*last); err != nil {
		return usageErrf("%v", err)
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
	fmt.Fprintf(stdout, "Avg Adjusted WPM  : %.2f\n", sum.AvgAdjustedWPM)
	fmt.Fprintf(stdout, "Avg Accuracy      : %.2f%%\n", sum.AvgAccuracy)
	fmt.Fprintf(stdout, "Avg Errors        : %.2f\n", sum.AvgErrors)
	fmt.Fprintf(stdout, "Net WPM stability (across sessions): %.2f\n", sum.NetWPMStability)
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
