package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"typer/internal/storage"
)

func runSet(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("set", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	wordsFile := fs.String("words-file", "", "Path to a newline-separated custom word list.")
	passagesFile := fs.String("passages-file", "", "Path to a blank-line-separated custom passages file.")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printSetHelp(stdout)
			return nil
		}
		return err
	}
	if err := rejectExtraArgs("set", fs.Args()); err != nil {
		return err
	}

	wf := strings.TrimSpace(*wordsFile)
	pf := strings.TrimSpace(*passagesFile)
	if wf == "" && pf == "" {
		return errors.New("set requires --words-file and/or --passages-file")
	}

	validateFile := func(label, path string) (string, error) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		info, err := os.Stat(absPath)
		if err != nil {
			return "", fmt.Errorf("invalid %s %q: %w", label, absPath, err)
		}
		if info.IsDir() {
			return "", fmt.Errorf("invalid %s %q: expected file, got directory", label, absPath)
		}
		return absPath, nil
	}

	settingsStore, err := storage.NewSettingsStore()
	if err != nil {
		return err
	}
	settings, err := settingsStore.Load()
	if err != nil {
		return err
	}

	if wf != "" {
		absPath, err := validateFile("words file", wf)
		if err != nil {
			return err
		}
		settings.WordsFile = absPath
	}
	if pf != "" {
		absPath, err := validateFile("passages file", pf)
		if err != nil {
			return err
		}
		settings.PassagesFile = absPath
	}

	if err := settingsStore.Save(settings); err != nil {
		return err
	}

	if wf != "" {
		fmt.Fprintf(stdout, "Custom words file set to %s\n", settings.WordsFile)
	}
	if pf != "" {
		fmt.Fprintf(stdout, "Custom passages file set to %s\n", settings.PassagesFile)
	}
	return nil
}
