package cli

import (
	"context"
	"io"
	"strings"

	"typer/internal/storage"
)

var newHistoryStore = storage.NewHistoryStore

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
	case "replay":
		return runReplay(ctx, args[1:], stdin, stdout)
	case "credits":
		return runCredits(stdout)
	case "key-press":
		return runKeyPress(ctx, args[1:], stdin, stdout)
	case "play":
		return runPlay(ctx, args[1:], stdin, stdout)
	case "train":
		return runTrain(ctx, args[1:], stdin, stdout)
	case "--version", "-v":
		return runVersion(stdout)
	case resetProgressFlag:
		return runResetProgress(args[1:], stdin, stdout, stderr)
	case "--help", "-h":
		printHelp(stdout)
		return nil
	default:
		printHelp(stderr)
		return usageErrf("unknown command %q", args[0])
	}
}

func rejectExtraArgs(command string, extras []string) error {
	if len(extras) == 0 {
		return nil
	}
	return usageErrf("%s does not take positional arguments: %s", command, strings.Join(extras, " "))
}
