package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"typer/internal/train"
)

type trainContinueAction int

const (
	trainContinueNext trainContinueAction = iota
	trainContinueRetry
	trainContinueStop
)

func printTrainContinuePrompt(out io.Writer, outcome train.LessonOutcome, adaptive bool) {
	if adaptive {
		fmt.Fprintln(out, "Space or Enter → next drill · Ctrl+C → stop")
		return
	}
	if outcome.Passed && outcome.NextLesson != "" {
		fmt.Fprintln(out, "Enter → next lesson · Space → retry · Ctrl+C → stop")
		return
	}
	if outcome.Passed {
		fmt.Fprintln(out, "Space or Enter → continue · Ctrl+C → stop")
		return
	}
	fmt.Fprintln(out, "Space or Enter → retry · Ctrl+C → stop")
}

func promptTrainContinue(stdin io.Reader, stdout io.Writer, outcome train.LessonOutcome, adaptive bool) (trainContinueAction, error) {
	printTrainContinuePrompt(stdout, outcome, adaptive)
	action, err := readTrainContinueKey(stdin)
	if err != nil {
		return trainContinueStop, err
	}
	if adaptive {
		return action, nil
	}
	if !outcome.Passed && action == trainContinueNext {
		return trainContinueRetry, nil
	}
	return action, nil
}

func readTrainContinueKey(stdin io.Reader) (trainContinueAction, error) {
	r := bufio.NewReader(stdin)
	for {
		b, err := r.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return trainContinueStop, nil
			}
			return trainContinueStop, err
		}
		switch b {
		case 3, 27: // ctrl+c, esc
			return trainContinueStop, nil
		case '\r', '\n':
			return trainContinueNext, nil
		case ' ':
			return trainContinueRetry, nil
		}
	}
}
