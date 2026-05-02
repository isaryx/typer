package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

const resetProgressFlag = "--reset-progress"

func readResetConfirm(stdin io.Reader) (ok bool, err error) {
	r := bufio.NewReader(stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && line == "" {
			return false, nil
		}
		if errors.Is(err, io.EOF) {
			// If we got a partial line without a trailing newline, still interpret it.
			return parseResetYes(strings.TrimSpace(line)), nil
		}
		return false, err
	}
	return parseResetYes(strings.TrimSpace(line)), nil
}

func parseResetYes(s string) bool {
	switch strings.ToLower(s) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

func runResetProgress(args []string, stdin io.Reader, out, outErr io.Writer) error {
	if len(args) > 0 {
		return fmt.Errorf("%s does not take additional arguments", resetProgressFlag)
	}
	fmt.Fprint(outErr, "Reset all local session history? [y/N]: ")
	ok, err := readResetConfirm(stdin)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(outErr, "Reset cancelled.")
		return nil
	}
	store, err := newHistoryStore()
	if err != nil {
		return err
	}
	if err := store.Reset(); err != nil {
		return err
	}
	fmt.Fprintln(out, "Progress reset. Local session history is now empty.")
	return nil
}
