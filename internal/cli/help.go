package cli

import (
	"fmt"
	"io"

	"typer/internal/version"
)

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
	fmt.Fprintln(out, "  replay       Re-run a history session; compare metrics (-l, --nth, --id).")
	fmt.Fprintln(out, "  stats        Summarize recent sessions.")
	fmt.Fprintln(out, "  credits      Show data source credits.")
	fmt.Fprintln(out, "  key-press    Show key presses.")
	fmt.Fprintln(out, "  version      Print the installed typer version.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Quick reference:")
	fmt.Fprintln(out, "  typer start [--mode|-m MODE] [--words|-w N] [--source SRC]")
	fmt.Fprintln(out, "              [--strict|-s] [--indefinite|-i] [--finger-hint|-f] [--no-ghost]")
	fmt.Fprintln(out, "  typer set [--words-file PATH] [--passages-file PATH]")
	fmt.Fprintln(out, "  typer history [--last N]")
	fmt.Fprintln(out, "  typer replay -l | --nth N | --id ID")
	fmt.Fprintln(out, "  typer stats [--last N]")
	fmt.Fprintln(out, "  typer credits")
	fmt.Fprintln(out, "  typer key-press")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Per-command help:")
	fmt.Fprintln(out, "  typer start -h | typer start --help")
	fmt.Fprintln(out, "  typer set -h   | typer set --help")
	fmt.Fprintln(out, "  typer history -h | typer stats -h | typer replay -h")
	fmt.Fprintln(out, "  typer key-press -h | typer key-press --help")
	fmt.Fprintln(out, "  typer --reset-progress         # type y/yes to confirm on the prompt")
}

func printStartHelp(out io.Writer) {
	fmt.Fprintln(out, "Run an interactive typing session.")
	fmt.Fprintln(out, "When local history has a best prior run of the same text with a typing trace, it is used as the ghost overlay unless --no-ghost.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer start [--mode|-m MODE] [--words|-w N] [--source SRC]")
	fmt.Fprintln(out, "              [--strict|-s] [--indefinite|-i] [--finger-hint|-f] [--no-ghost]")
	fmt.Fprintln(out, "  typer [same flags...]            # omit the word \"start\" when the first argument is a flag")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Options:")
	fmt.Fprintln(out, "  -m, --mode string    Text mode: passages|p, words|w, or quotes|q (default quotes).")
	fmt.Fprintln(out, "  -w, --words int      Words per prompt in words mode (default 15).")
	fmt.Fprintln(out, "      --source string  Quotes mode only: auto|remote|cache|seed (default seed).")
	fmt.Fprintln(out, "  -s, --strict         Strict matching: wrong input does not advance (bare flag; omit for non-strict).")
	fmt.Fprintln(out, "  -i, --indefinite     After each finished session, start another until Ctrl+C or Esc (bare flag).")
	fmt.Fprintln(out, "  -f, --finger-hint    Show US QWERTY finger hints for the next key (bare flag; omit for no hints).")
	fmt.Fprintln(out, "      --no-ghost       Skip the ghost overlay from history (by default, the best prior run of the same text is used when a typing trace exists).")
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

func printReplayHelp(out io.Writer) {
	fmt.Fprintln(out, "Re-run a session from local history (`typer history` lists ids). Your run is scored against the saved one.")
	fmt.Fprintln(out, "With a typing trace, a dim shadow replays the old keystrokes above your input; without it you get a short note.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer replay -l")
	fmt.Fprintln(out, "  typer replay --nth N   # 1 = newest, same order as typer history")
	fmt.Fprintln(out, "  typer replay --id ID")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Options:")
	fmt.Fprintln(out, "  -l, --last      Most recent session.")
	fmt.Fprintln(out, "      --nth int   N-th newest session (1 = newest).")
	fmt.Fprintln(out, "      --id string Session id from history (id=... line).")
}
