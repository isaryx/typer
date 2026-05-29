package cli

import (
	"context"
	"fmt"
	"io"
)

func runPlay(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		game, err := selectPlayGame(stdin, stdout)
		if err != nil {
			return err
		}
		if game == "" {
			return nil
		}
		return runPlayGame(ctx, game, nil, stdin, stdout)
	}
	switch args[0] {
	case "--help", "-h":
		printPlayHelp(stdout)
		return nil
	default:
		return runPlayGame(ctx, args[0], args[1:], stdin, stdout)
	}
}

func runPlayGame(ctx context.Context, game string, args []string, stdin io.Reader, stdout io.Writer) error {
	switch game {
	case "hangman":
		return runPlayHangman(ctx, args, stdin, stdout)
	case "defense":
		return runPlayDefense(ctx, args, stdin, stdout)
	default:
		return usageErrf("unknown play subcommand %q (try: hangman, defense)", game)
	}
}

func printPlayHelp(out io.Writer) {
	fmt.Fprintln(out, "Solo typing games. Run without a subcommand to choose interactively.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer play")
	fmt.Fprintln(out, "  typer play hangman ...")
	fmt.Fprintln(out, "  typer play defense ...")
	fmt.Fprintln(out)
	printPlayHangmanHelp(out)
	fmt.Fprintln(out)
	printPlayDefenseHelp(out)
}
