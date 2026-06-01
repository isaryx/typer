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
	"typer/internal/text"
)

// quoteSourceFlagList accumulates repeated --quote-source values.
type quoteSourceFlagList []string

func (q *quoteSourceFlagList) String() string { return "" }

func (q *quoteSourceFlagList) Set(s string) error {
	*q = append(*q, s)
	return nil
}

func runSet(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("set", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	wordsFile := fs.String("words-file", "", "Words mode word list file.")
	passagesFile := fs.String("passages-file", "", "Passages mode passages file.")
	showHint := fs.String("show-hint", "", `Hint visibility: on|off|yes|no|true|false|1|0 (omit = unchanged).`)
	inputPosition := fs.String("input-position", "", `Input line placement (default otd; e.g. top-left, bc, on-top / ot, on-top-dynamic / otd).`)
	var quoteToggles quoteSourceFlagList
	fs.Var(&quoteToggles, "quote-source", "Remote quote API: ID=on|off (repeat). IDs: "+strings.Join(text.KnownQuoteRemoteIDs(), ", ")+".")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printSetHelp(stdout)
			return nil
		}
		return usageErrf("%v", err)
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
		return usageErrf("set: --show-hint must be one of on, off, yes, no, true, false, 1, 0, got %q", strings.TrimSpace(*showHint))
	}
	inputPosArg := strings.TrimSpace(*inputPosition)
	var inputPlacement *model.InputPlacement
	if inputPosArg != "" {
		p, err := model.ParseInputPosition(inputPosArg)
		if err != nil {
			return usageErrf("set: %v", err)
		}
		inputPlacement = &p
	}
	if wf == "" && pf == "" && showHintPtr == nil && inputPlacement == nil && len(quoteToggles) == 0 {
		return usageErrf("set requires --words-file, --passages-file, --show-hint, --input-position, and/or --quote-source")
	}

	validateFile := func(label, path string) (string, error) {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", usageErrf("invalid %s %q: %v", label, path, err)
		}
		info, err := os.Stat(absPath)
		if err != nil {
			return "", usageErrf("invalid %s %q: %v", label, absPath, err)
		}
		if info.IsDir() {
			return "", usageErrf("invalid %s %q: expected file, got directory", label, absPath)
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
	for _, pair := range quoteToggles {
		id, on, err := parseQuoteSourceToggle(pair)
		if err != nil {
			return usageErrf("set: %v", err)
		}
		if settings.QuoteRemoteEnabled == nil {
			settings.QuoteRemoteEnabled = map[string]bool{}
		}
		settings.QuoteRemoteEnabled[id] = on
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
	if len(quoteToggles) > 0 {
		var parts []string
		for _, id := range text.KnownQuoteRemoteIDs() {
			state := "on"
			if !text.QuoteRemoteEffectiveEnabled(settings.QuoteRemoteEnabled, id) {
				state = "off"
			}
			parts = append(parts, id+"="+state)
		}
		fmt.Fprintf(stdout, "Remote quote APIs: %s\n", strings.Join(parts, " "))
	}
	return nil
}

func parseQuoteSourceToggle(s string) (id string, on bool, err error) {
	s = strings.TrimSpace(s)
	idx := strings.IndexByte(s, '=')
	if idx <= 0 || idx == len(s)-1 {
		return "", false, usageErrf("--quote-source expects ID=on|off, got %q", s)
	}
	id = strings.ToLower(strings.TrimSpace(s[:idx]))
	val := strings.ToLower(strings.TrimSpace(s[idx+1:]))
	switch val {
	case "on", "true", "1", "yes":
		on = true
	case "off", "false", "0", "no":
		on = false
	default:
		return "", false, usageErrf("--quote-source value must be on|off, got %q", val)
	}
	if !text.IsKnownQuoteRemoteID(id) {
		return "", false, usageErrf("unknown quote remote ID %q (known: %s)", id, strings.Join(text.KnownQuoteRemoteIDs(), ", "))
	}
	return id, on, nil
}
