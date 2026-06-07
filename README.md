# Typer

Cross-platform typing trainer CLI written in Go.

## Features

- Practice in the terminal: **passages**, **words**, or **quotes**
- **Words mode** builds each prompt from the bundled word list with a balanced mix of short, medium, and long words (frequency-ranked source; scales with `--words`)
- Typing games: **hangman** and **defense** (`typer play`) — defense uses the same word list with progressive length as your score rises
- Speed and accuracy stats; sessions saved on your machine
- History, stats, and replay when you want to compare runs
- Offline-friendly with bundled text

## Requirements

[Go](https://go.dev/dl/) — only needed to build from this repo; match the `go` line in [`go.mod`](go.mod) or newer.

## Install

**Homebrew (macOS / Linux)**

```bash
brew tap isaryx/collection
brew install typer
```

**Or** use a release from GitHub Releases: put `typer` (macOS/Linux) or `typer.exe` (Windows) on your `PATH`.

## Usage

```bash
typer start
typer play          
typer history
```

Run `typer --help` or `typer -h` for commands and global flags; use `typer start --help` or `typer start -h` for session flags (same `--help`/`-h` pattern on other commands). From a clone, use `go run ./cmd/typer` instead of `typer`.

## Build

```bash
go build -o typer ./cmd/typer
```

Embed a version when building:

```bash
go build -ldflags "-X typer/internal/version.Version=0.01" -o typer ./cmd/typer
```

## Project docs

- [Contributing](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Security](SECURITY.md)

## Privacy & local data

History and settings stay on your device (typical paths include `~/.config/typer` on Linux). Saved sessions include what you typed so you can review them later. By default, practice can show a light ghost from a previous run on the same text; pass `--no-ghost` on `typer start` if you prefer not to.

## Credits

**UI:** [Bubble Tea v2](https://pkg.go.dev/charm.land/bubbletea/v2), [Lip Gloss v2](https://pkg.go.dev/charm.land/lipgloss/v2) (MIT); [golang.org/x/term](https://pkg.go.dev/golang.org/x/term) (BSD-3-Clause). Source repos: [bubbletea](https://github.com/charmbracelet/bubbletea), [lipgloss](https://github.com/charmbracelet/lipgloss).

**Session IDs:** [oklog/ulid](https://github.com/oklog/ulid) ([spec](https://github.com/ulid/spec)), Apache-2.0.

**Bundled content:** Words from [first20hours/google-10000-english](https://github.com/first20hours/google-10000-english) ([details](assets/WORDS_CREDITS.md)) — shared by words mode and defense; override with `typer set --words-file PATH`. Passages from public-domain proverbs ([details](assets/PASSAGES_CREDITS.md)). For live quotes, ZenQuotes batch (`/api/quotes`) fills a local pool (one quote per session until empty, then refetch); optional `zenquotes-random` uses `/api/random`; [type.fit](https://type.fit/api/quotes) is the fallback. Toggle remotes with `typer set --quote-source` (see `typer credits` for attribution). Bundled `assets/quotes.json` is used when remotes are off or unavailable.
