package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

func selectPlayGame(stdin io.Reader, stdout io.Writer) (string, error) {
	fmt.Fprintln(stdout, "Choose a game:")
	fmt.Fprintln(stdout, "  1) hangman   Type a quote; mistakes draw the gallows")
	fmt.Fprintln(stdout, "  2) defense   Type falling words before they reach the shield")
	fmt.Fprint(stdout, "> ")

	line, err := readPlayChoiceLine(stdin)
	if err != nil {
		return "", err
	}
	choice := strings.TrimSpace(line)
	if choice == "" {
		return "", nil
	}
	game, ok := parsePlayChoice(choice)
	if !ok {
		return "", usageErrf("invalid choice %q (enter 1, 2, hangman, or defense)", choice)
	}
	return game, nil
}

func readPlayChoiceLine(stdin io.Reader) (string, error) {
	r := bufio.NewReader(stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && line == "" {
			return "", nil
		}
		if errors.Is(err, io.EOF) {
			return strings.TrimSpace(line), nil
		}
		return "", err
	}
	return line, nil
}

func parsePlayChoice(s string) (game string, ok bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "hangman":
		return "hangman", true
	case "2", "defense":
		return "defense", true
	default:
		return "", false
	}
}
