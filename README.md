# Typer

Cross-platform typing trainer CLI written in Go.

## Features

- English-only UI and bundled practice text
- Sessions: **passages**, **words**, or **quotes**
- Gross/net WPM, accuracy, errors, elapsed time; history saved locally as JSON
- Offline-first with bundled text and quote seeds

## Requirements

Go **1.24+** (see `go.mod`).

## Install

**Homebrew (macOS / Linux)**

```bash
brew tap isaryx/collection
brew install typer
```

**Or** grab a release from GitHub Releases: extract the archive and put `typer` (macOS/Linux) or `typer.exe` (Windows) on your `PATH`.

## Usage

After install, run `typer` from your terminal. Examples:

```bash
typer start --mode passages
typer start --mode words --words 50
typer start --mode quotes
typer history --last 20
typer stats --last 20
typer version
typer credits
```

From a clone, use `go run ./cmd/typer …` instead of `typer`. Use `typer --help` and `typer start --help` for flags (strict mode, custom word/passage files, quote source, reset progress, etc.).

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

Session history and settings stay on your machine (for example under `~/.config/typer` / `~/.cache/typer` on Linux). History stores what you typed so you can review sessions later. Each session also stores a **SHA-256 hash** of the canonical prompt text (not reversible to recover the passage from the hash alone) so runs with the same text can be grouped—`typer start` uses a ghost overlay from the **best** saved run of the same text when one exists (with a typing trace). Pass `--no-ghost` to turn that off.

## Credits

**UI:** [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss) (MIT).

**Session IDs (ULID):** [oklog/ulid](https://github.com/oklog/ulid) — Go package [`github.com/oklog/ulid/v2`](https://pkg.go.dev/github.com/oklog/ulid/v2), Apache License 2.0. ([ULID specification](https://github.com/ulid/spec).)

**Bundled content:** Words list from [first20hours/google-10000-english](https://github.com/first20hours/google-10000-english); passages built from public-domain proverbs (details in `assets/PASSAGES_CREDITS.md`). Quotes mode can use [type.fit](https://type.fit/api/quotes) or bundled seed data.

Scoring formulas and edge cases live in `internal/scoring/scoring.go`.
