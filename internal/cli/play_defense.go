package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"typer/internal/game/defense"
	"typer/internal/storage"
)

func runPlayDefense(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("play defense", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var lives int
	fs.IntVar(&lives, "lives", defense.DefaultLives, "Starting lives.")
	var wordsFile string
	fs.StringVar(&wordsFile, "words-file", "", "Word list path (default from settings).")
	var spawnRate float64
	fs.Float64Var(&spawnRate, "spawn-rate", defense.DefaultSpawnSeconds, "Average seconds between spawns.")
	var fallSpeed float64
	fs.Float64Var(&fallSpeed, "fall-speed", defense.DefaultBaseFallSpeed, "Starting rows per second.")
	var noAudible bool
	fs.BoolVar(&noAudible, "no-audible", false, "Disable terminal bell on mistakes.")
	var seed uint64
	fs.Uint64Var(&seed, "seed", 0, "RNG seed (0 = random each run).")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			printPlayDefenseHelp(stdout)
			return nil
		}
		return usageErrf("%v", err)
	}
	if err := rejectExtraArgs("play defense", fs.Args()); err != nil {
		return err
	}

	cfg := defense.Config{
		Lives:            lives,
		BaseSpawnSeconds: spawnRate,
		BaseFallSpeed:    fallSpeed,
		NoAudible:        noAudible,
	}
	if err := defense.ValidateConfig(cfg); err != nil {
		return usageErrf("%v", err)
	}

	settingsStore, err := storage.NewSettingsStore()
	if err != nil {
		return err
	}
	settings, err := settingsStore.Load()
	if err != nil {
		return err
	}
	if wordsFile == "" {
		wordsFile = settings.WordsFile
	}

	pool, err := defense.LoadWordPool(wordsFile)
	if err != nil {
		return err
	}

	result, err := defense.Run(ctx, stdin, stdout, pool, cfg, seed)
	if err != nil {
		return err
	}

	if !result.Aborted {
		historyStore, err := storage.NewHistoryStore()
		if err != nil {
			return err
		}
		if err := historyStore.Append(defense.BuildSessionResult(result)); err != nil {
			return err
		}
	}

	defense.PrintOutcome(stdout, result)
	return nil
}

func printPlayDefenseHelp(out io.Writer) {
	fmt.Fprintln(out, "Defense — destroy falling words by typing them (strict mode).")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  typer play defense [--lives 6] [--words-file PATH]")
	fmt.Fprintln(out, "                     [--spawn-rate 1.2] [--fall-speed 0.5]")
	fmt.Fprintln(out, "                     [--seed N] [--no-audible]")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Flags:")
	fmt.Fprintln(out, "  --lives int          Starting lives (default 6).")
	fmt.Fprintln(out, "  --words-file string  Word list (default from settings / bundled).")
	fmt.Fprintln(out, "  --spawn-rate float   Average seconds between spawns (default 1.2).")
	fmt.Fprintln(out, "  --fall-speed float   Starting fall speed in rows/sec (default 0.5).")
	fmt.Fprintln(out, "  --seed uint          RNG seed for reproducible runs (0 = random).")
	fmt.Fprintln(out, "  --no-audible         Disable terminal bell on wrong keys.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Controls:")
	fmt.Fprintln(out, "  Tab / Esc            Clear word lock.")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintln(out, "  typer play defense")
	fmt.Fprintln(out, "  typer play defense --lives 8 --fall-speed 1.5 --seed 42")
}
