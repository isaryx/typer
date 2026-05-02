package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"typer/internal/model"
	"typer/internal/storage"
)

func runSet(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("set", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	wordsFile := fs.String("words-file", "", "Words mode word list file.")
	passagesFile := fs.String("passages-file", "", "Passages mode passages file.")
	showHint := fs.String("show-hint", "", `Hint visibility: on|off|yes|no|true|false|1|0 (omit = unchanged).`)
	inputPosition := fs.String("input-position", "", `Input line placement (e.g. top-left, bc).`)
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
	showHintArg := strings.TrimSpace(strings.ToLower(*showHint))
	var showHintPtr *bool
	switch showHintArg {
	case "":
		break
	case "on", "true", "1", "yes":
		v := true
		showHintPtr = &v
	case "off", "false", "0", "no":
		v := false
		showHintPtr = &v
	default:
		return fmt.Errorf("set: --show-hint must be one of on, off, yes, no, true, false, 1, 0, got %q", strings.TrimSpace(*showHint))
	}
	inputPosArg := strings.TrimSpace(*inputPosition)
	var inputPlacement *model.InputPlacement
	if inputPosArg != "" {
		p, err := model.ParseInputPosition(inputPosArg)
		if err != nil {
			return fmt.Errorf("set: %w", err)
		}
		inputPlacement = &p
	}
	if wf == "" && pf == "" && showHintPtr == nil && inputPlacement == nil {
		return errors.New("set requires --words-file, --passages-file, --show-hint, and/or --input-position")
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
	if showHintPtr != nil {
		settings.ShowHint = showHintPtr
	}
	if inputPlacement != nil {
		settings.InputPosition = inputPlacement.CanonicalString()
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
	if showHintPtr != nil {
		if *showHintPtr {
			fmt.Fprintln(stdout, "Typing hint: on")
		} else {
			fmt.Fprintln(stdout, "Typing hint: off")
		}
	}
	if inputPlacement != nil {
		fmt.Fprintf(stdout, "Input position: %s\n", settings.InputPosition)
	}
	return nil
}
