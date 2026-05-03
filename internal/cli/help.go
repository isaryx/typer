package cli

import (
	"fmt"
	"io"
	"text/tabwriter"

	"typer/internal/model"
	"typer/internal/version"
)

func writeHelpTabwriter(out io.Writer, write func(*tabwriter.Writer)) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	write(w)
	w.Flush()
}

func printHelp(out io.Writer) {
	fmt.Fprintf(out, "typer %s — terminal typing practice\n\n", version.Version)

	writeHelpTabwriter(out, func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "Usage:")
		fmt.Fprintln(w, "  typer <command> [arguments]")
		fmt.Fprintf(w, "  %s\t%s\n", "typer --reset-progress", "Clear history (confirm y/yes).")
		fmt.Fprintf(w, "  %s\t%s\n", "typer [flags...]", "Same as typer start when the first arg is a flag.")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Commands:")
		cmds := []struct {
			name string
			desc string
		}{
			{"start", "Interactive typing session."},
			{"set", "Save corpus, hint, input layout (see typer set -h)."},
			{"history", "List recent sessions."},
			{"replay", "Replay a saved session; compare metrics."},
			{"stats", "Stats over recent sessions."},
			{"credits", "Data sources and libraries."},
			{"key-press", "Echo keys (demo)."},
			{"version", "Print version."},
		}
		for _, c := range cmds {
			fmt.Fprintf(w, "  %s\t%s\n", c.name, c.desc)
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Subcommand help:")
		fmt.Fprintln(w, "  typer <command> -h    Full flags for that command (same as --help).")
	})
}

func printStartHelp(out io.Writer) {
	fmt.Fprintln(out, "Interactive session. Ghost uses your best prior run of the same text unless --no-ghost.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer start [--mode|-m MODE] [--words|-w N] [--source SRC] [--quote-sources IDS]")
	fmt.Fprintln(out, "              [--strict|-s] [--indefinite|-i] [--finger-hint|-f] [--no-ghost] [--no-input] [--no-audible]")
	fmt.Fprintln(out, "  typer [same flags...]            # omit \"start\" when the first argument is a flag")
	fmt.Fprintln(out)

	writeHelpTabwriter(out, func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "Options:")
		fmt.Fprintf(w, "  %s\t%s\n", "-m, --mode string", "passages|p, words|w, quotes|q (default quotes).")
		fmt.Fprintf(w, "  %s\t%s\n", "-w, --words int", fmt.Sprintf("Words per prompt (words mode), 1..%d (default 15).", model.MaxWordsPerPrompt))
		fmt.Fprintf(w, "  %s\t%s\n", "--source string", "Quotes only: remote|cache|seed (default remote).")
		fmt.Fprintf(w, "  %s\t%s\n", "--quote-sources string", "Quotes only: comma-separated remote IDs this session (zenquotes, typefit); overrides saved toggles.")
		fmt.Fprintf(w, "  %s\t%s\n", "-s, --strict", "Wrong character does not advance (bare flag).")
		fmt.Fprintf(w, "  %s\t%s\n", "-i, --indefinite", "Run another session after each finish until Ctrl+C (bare flag).")
		fmt.Fprintf(w, "  %s\t%s\n", "-f, --finger-hint", "QWERTY finger diagram for next key (bare flag).")
		fmt.Fprintf(w, "  %s\t%s\n", "--no-ghost", "No ghost overlay from history.")
		fmt.Fprintf(w, "  %s\t%s\n", "--no-input", "Hide input line.")
		fmt.Fprintf(w, "  %s\t%s\n", "--no-audible", "Disable terminal bell on mistakes.")
	})
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Environment:")
	fmt.Fprintln(out, "  TYPER_REDUCE_MOTION  If 1, true, yes, or on: no ANSI blink on the inline input caret (█).")
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintln(out, "  typer start --quote-sources typefit")
	fmt.Fprintln(out, "  typer start --quote-sources zenquotes,typefit")
}

func printSetHelp(out io.Writer) {
	fmt.Fprintln(out, "Merge into saved config; omitted flags unchanged. One flag minimum.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer set [options]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "  typer set [--words-file PATH] [--passages-file PATH]")
	fmt.Fprintln(out, "          [--show-hint on|off] [--input-position PLACE]")
	fmt.Fprintln(out, "          [--quote-source ID=on|off] ...")
	fmt.Fprintln(out)

	writeHelpTabwriter(out, func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "Corpus:")
		fmt.Fprintf(w, "  %s\t%s\n", "--words-file PATH", "Words mode list (newline-separated).")
		fmt.Fprintf(w, "  %s\t%s\n", "--passages-file PATH", "Passages mode (blocks separated by blank lines).")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Session UI:")
		fmt.Fprintf(w, "  %s\t%s\n", "--show-hint on|off", "Show or hide hint (default on).")
		fmt.Fprintf(w, "  %s\t%s\n", "--input-position PLACE", "Input row (default on-top-dynamic / otd): top|bottom + left|center|right; shorthand tl|tc|tr|bl|bc|br; border on-top|on-bottom (ot|ob); dynamic otd|obd (on-top-dynamic|on-bottom-dynamic) shifts the border input horizontally to follow the active word.")
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Quote Source:")
		fmt.Fprintf(w, "  %s\t%s\n", "--quote-source ID=on|off", "Enable/disable a remote quote source.")
	})
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintln(out, "  typer set --quote-source zenquotes=on --quote-source typefit=on")
	fmt.Fprintln(out, "  typer set --quote-source zenquotes=off --quote-source typefit=on")
}

func printHistoryHelp(out io.Writer) {
	fmt.Fprintln(out, "List recent sessions from local history.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer history [--last N]")
	fmt.Fprintln(out)

	writeHelpTabwriter(out, func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "Options:")
		fmt.Fprintf(w, "  %s\t%s\n", "--last int", fmt.Sprintf("How many to list, 1..%d (default 20).", model.MaxRetainedHistorySessions))
	})
}

func printStatsHelp(out io.Writer) {
	fmt.Fprintln(out, "Aggregate stats over recent sessions.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer stats [--last N]")
	fmt.Fprintln(out)

	writeHelpTabwriter(out, func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "Options:")
		fmt.Fprintf(w, "  %s\t%s\n", "--last int", fmt.Sprintf("How many sessions, 1..%d (default 20).", model.MaxRetainedHistorySessions))
	})
}

func printReplayHelp(out io.Writer) {
	fmt.Fprintln(out, "Replay a saved session and compare to it. With a typing trace, a shadow replays the old keystrokes.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer replay -l")
	fmt.Fprintln(out, "  typer replay --nth N   # 1 = newest (same order as typer history)")
	fmt.Fprintln(out, "  typer replay --id ID")
	fmt.Fprintln(out)

	writeHelpTabwriter(out, func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "Options:")
		fmt.Fprintf(w, "  %s\t%s\n", "-l, --last", "Newest session.")
		fmt.Fprintf(w, "  %s\t%s\n", "--nth int", "Nth newest (1 = newest).")
		fmt.Fprintf(w, "  %s\t%s\n", "--id string", "Session id from typer history.")
		fmt.Fprintf(w, "  %s\t%s\n", "--no-input", "Hide input line.")
		fmt.Fprintf(w, "  %s\t%s\n", "--no-audible", "Disable terminal bell on mistakes.")
	})
}
