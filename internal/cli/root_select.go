package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
)

func selectRootCommand(stdin io.Reader, stdout io.Writer) (string, error) {
	fmt.Fprintln(stdout, "Choose a command:")
	fmt.Fprintln(stdout, "  1) start     Free practice session")
	fmt.Fprintln(stdout, "  2) train     Structured training with lessons")
	fmt.Fprintln(stdout, "  3) play      Typing games")
	fmt.Fprintln(stdout, "  4) history   List recent sessions")
	fmt.Fprintln(stdout, "  5) stats     View aggregated statistics")
	fmt.Fprint(stdout, "> ")

	line, err := readPlayChoiceLine(stdin)
	if err != nil {
		return "", err
	}
	choice := strings.TrimSpace(line)
	if choice == "" {
		return "", nil
	}
	cmd, ok := parseRootChoice(choice)
	if !ok {
		return "", usageErrf("invalid choice %q (enter 1–5 or a command name)", choice)
	}
	return cmd, nil
}

func parseRootChoice(s string) (command string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "start":
		return "start", true
	case "2", "train":
		return "train", true
	case "3", "play":
		return "play", true
	case "4", "history":
		return "history", true
	case "5", "stats":
		return "stats", true
	default:
		return "", false
	}
}

func runRootInteractive(ctx context.Context, stdin io.Reader, stdout io.Writer) error {
	cmd, err := selectRootCommand(stdin, stdout)
	if err != nil {
		return err
	}
	if cmd == "" {
		return nil
	}
	switch cmd {
	case "start":
		return runStart(ctx, nil, stdin, stdout)
	case "train":
		return runTrain(ctx, nil, stdin, stdout)
	case "play":
		return runPlay(ctx, nil, stdin, stdout)
	case "history":
		return runHistory(nil, stdout)
	case "stats":
		return runStats(nil, stdout)
	default:
		return usageErrf("unknown command %q", cmd)
	}
}
