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
typer set --words-file ./path-to-words.txt
typer set --passages-file ./path-to-passages.txt
typer set --words-file ./words.txt --passages-file ./passages.txt
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

After a run, the CLI prints the same numbers that are stored in history (see `internal/scoring` for formulas).

- **Gross WPM** — Keystroke-based words per minute: (keystrokes / 5) / minutes elapsed.
- **Net WPM** — Gross WPM minus uncorrected character errors per minute.
- **Accuracy** — Share of keystrokes that matched the expected character at the time of the keypress.
- **Errors** — Count of character-level mismatches still present when a word is submitted (and similar effects from length skew).
- **Consistency** — Score from **0 to 100** for how *steady* your *speed* was: after each completed word, a running gross WPM is sampled, and the spread of those samples is turned into a score (**100** = very stable, lower = more up-and-down pace). It is *not* accuracy; you can be consistently slow or fast.
- **Time** — Wall-clock duration of the session (shown in a readable form in the result table).

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
