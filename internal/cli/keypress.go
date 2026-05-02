package cli

import (
	"context"
	"fmt"
	"io"

	"typer/internal/keypress"
)

func printKeyPressHelp(out io.Writer) {
	fmt.Fprintln(out, "Show each key as centered text; the last 10 keys appear on the line below.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer key-press")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Press Ctrl+C to quit.")
}

func runKeyPress(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	for _, a := range args {
		if a == "-h" || a == "--help" {
			printKeyPressHelp(stdout)
			return nil
		}
	}
	if err := rejectExtraArgs("key-press", args); err != nil {
		return err
	}
	return keypress.RunKeyPress(ctx, stdin, stdout)
}
