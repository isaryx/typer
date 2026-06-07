package cli

import (
	"fmt"
	"io"
	"text/tabwriter"
)

func printTrainHelp(out io.Writer) {
	fmt.Fprintln(out, "Structured typing training with lessons and placement test.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer train [-e|--evaluate] [--list] [--lesson ID]")
	fmt.Fprintln(out, "  typer train status")
	fmt.Fprintln(out, "  typer train reset")
	fmt.Fprintln(out)

	writeHelpTabwriter(out, func(w *tabwriter.Writer) {
		fmt.Fprintln(w, "Flags:")
		fmt.Fprintf(w, "  %s\t%s\n", "-e, --evaluate", "Run placement test and create/replace profile (prompts if one exists).")
		fmt.Fprintf(w, "  %s\t%s\n", "--list", "Show curriculum and completion status.")
		fmt.Fprintf(w, "  %s\t%s\n", "--lesson string", "Replay an unlocked lesson without changing saved progress.")
	})
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Subcommands:")
	fmt.Fprintln(out, "  status    Show training profile summary.")
	fmt.Fprintln(out, "  reset     Clear training profile (history is kept).")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintln(out, "  typer train -e")
	fmt.Fprintln(out, "  typer train          # lesson loop until stop (Ctrl+C or continue prompt)")
	fmt.Fprintln(out, "  typer train --lesson 2.4")
	fmt.Fprintln(out, "  typer train status")
}
