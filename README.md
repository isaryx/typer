# Typer

`typer` is a cross-platform typing trainer CLI written in Go.

## Features

- `start` a typing session from passages, words, or quotes mode
- See gross/net WPM, accuracy, errors, and elapsed time
- Save session history locally as JSON
- Offline-first with bundled text and quote seed data

## Requirements

- Go 1.24+ (same as the `go` version in `go.mod`)

## Run

```bash
go run ./cmd/typer start --mode passages
```

```bash
go run ./cmd/typer start --mode words --words 50
```

```bash
go run ./cmd/typer start --mode quotes
```

```bash
go run ./cmd/typer history --last 20
```

```bash
go run ./cmd/typer stats --last 20
```

```bash
go run ./cmd/typer version
```

```bash
go run ./cmd/typer credits
```

```bash
typer set --words-file ./path-to-words.txt
typer set --passages-file ./path-to-passages.txt
typer set --words-file ./words.txt --passages-file ./passages.txt
```

```bash
typer --reset-progress
# at the prompt, type y (or yes) to clear saved history
```

## Build

Local build:

```bash
go build -o typer ./cmd/typer
```

Cross-platform examples:

```bash
GOOS=linux GOARCH=amd64 go build -o dist/typer-linux-amd64 ./cmd/typer
GOOS=darwin GOARCH=arm64 go build -o dist/typer-darwin-arm64 ./cmd/typer
GOOS=windows GOARCH=amd64 go build -o dist/typer-windows-amd64.exe ./cmd/typer
```

Embed version when building manually:

```bash
go build -ldflags "-X typer/internal/version.Version=0.01" -o typer ./cmd/typer
```

## Install

Download the latest archive from this repository's GitHub Releases page, then extract and place the binary on your `PATH`.

- macOS/Linux: extract `.tar.gz` and move `typer` to somewhere like `/usr/local/bin`.
- Windows: extract `.zip` and add the folder containing `typer.exe` to `PATH`.

## Open-source libraries

The typing session UI is built with:

- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** — TUI framework ([MIT License](https://github.com/charmbracelet/bubbletea/blob/main/LICENSE))
- **[Lip Gloss](https://github.com/charmbracelet/lipgloss)** — terminal styles and layout ([MIT License](https://github.com/charmbracelet/lipgloss/blob/main/LICENSE))

Other entries in `go.sum` are transitive dependencies (Charm’s terminal stack, `golang.org/x/sys`, `golang.org/x/text`, and related packages). See each module’s repository for its license.

## Notes

- Quote mode remote source currently uses `https://type.fit/api/quotes`.
- Fallback policy for quote mode:
  - `remote` (default), `auto`: try remote API, then cache, then seed
  - `cache`: cache, then seed
  - `seed`: seed only
- Session UX:
  - Current target word is highlighted.
  - Press `space` to submit current word and advance to next word.
  - Pass `--strict` or `-s` to block advance on mismatch (strict mode). Omit both for non-strict (wrong input can still advance; mismatches count as errors).

## Session metrics

After a run, the CLI prints the same numbers that are stored in history.

### Scoring contract (current behavior)

`typer` uses the common typing-test convention of **5 characters = 1 word**.

Formulas:

- **Gross WPM** = `(totalKeystrokes / 5) / minutesElapsed`
- **Net WPM** = `GrossWPM - (uncorrectedErrors / minutesElapsed)` (floored at `0`)
- **Adjusted WPM** = `GrossWPM * (Accuracy / 100)`
- **Accuracy (%)** = `(correctKeystrokes / totalKeystrokes) * 100`

These formulas are implemented in `internal/scoring/scoring.go`.

### Counting rules

- **`totalKeystrokes`** counts rune keypresses typed into the active word.
- **`correctKeystrokes`** counts rune keypresses that match the expected rune at the time of press.
- **Backspace, Space, and Enter** are not counted as keystrokes for WPM/accuracy.
- **`uncorrectedErrors`** is measured when a word is submitted and includes:
  - wrong characters at aligned positions,
  - extra typed characters,
  - missing trailing target characters.

### Strict vs non-strict

- **Strict mode (`--strict`, `-s`)**: you cannot advance unless the current word matches exactly, so submitted-word errors are typically `0`.
- **Non-strict mode (default)**: you can advance with mismatches; those mismatches contribute to `uncorrectedErrors`.

### Edge cases and display

- If elapsed time is zero or extremely small, scoring uses a minimum of **1 second** (`1/60` minute) to avoid divide-by-zero.
- All displayed/stored metric values are rounded to 2 decimals.
- **Consistency** is a separate `0-100` steadiness score derived from per-word gross-WPM samples (not an accuracy metric).
- **Time** starts on the first typed rune and ends when the session completes (or aborts).

## Data Credits

- Words mode default list (`assets/words.txt`) is sourced from [first20hours/google-10000-english](https://github.com/first20hours/google-10000-english).
- You can override words mode list at runtime with `typer set --words-file <path>`, and passage mode corpus with `typer set --passages-file <path>` (same blank-line-separated block format as the bundle).
- Original corpus source: Google Web Trillion Word Corpus (via LDC), with cleanup by Josh Kaufman and subsets by Peter Norvig.
- Usage note from upstream license: educational and personal/research use is permitted; commercial use may require separate LDC licensing.
- Passage mode bundled list (`assets/passages.txt`) contains 200 public-domain English proverb passages from
  Thomas Preston’s *Dictionary of English Proverbs and Proverbial Phrases* (Gutenberg eBook #39281). The
  generator bundles consecutive short lines into **15–50 word** passages (see `assets/PASSAGES_CREDITS.md`
  and `scripts/build_passages.py`).
- Quote API source: [type.fit](https://type.fit/api/quotes).
