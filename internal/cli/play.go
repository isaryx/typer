package cli

import (
	"context"
	"io"
)

func runPlay(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		printPlayHelp(stdout)
		return nil
	}
	switch args[0] {
	case "hangman":
		return runPlayHangman(ctx, args[1:], stdin, stdout)
	case "--help", "-h":
		printPlayHelp(stdout)
		return nil
	default:
		return usageErrf("unknown play subcommand %q (try: hangman)", args[0])
	}
}

func printPlayHelp(out io.Writer) {
	printPlayHangmanHelp(out)
}
