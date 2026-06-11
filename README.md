# Typer

Cross-platform typing trainer CLI written in Go. Practice and track speed in the terminal — no account, no browser.

## Features

- **Practice** (`typer start`) — type passages, words, or quotes; optional strict mode, finger hints, and a ghost of your best run on the same text
- **Train** (`typer train`) — placement test, guided lessons, then adaptive drills on keys you miss most
- **Review** (`typer history`, `stats`, `replay`) — browse past sessions, see averages and common mistakes, race a previous run
- **Play** (`typer play`) — hangman and defense typing games
- **Customize** (`typer set`) — your own word lists and passages; tune hints, input layout, and quote sources
- **Offline and private** — bundled text by default; sessions and progress saved locally on your machine

## Install

**Homebrew (macOS / Linux)**

```bash
brew tap isaryx/collection
brew install typer
```

**Or** download a binary from [GitHub Releases](https://github.com/isaryx/typer/releases) and put `typer` (macOS/Linux) or `typer.exe` (Windows) on your `PATH`. Use a modern terminal for the interactive UI.

Verify:

```bash
typer --version
```

Go is only required when [building from source](#build) — match the `go` line in [`go.mod`](go.mod) or newer ([Go downloads](https://go.dev/dl/)).

## Quick start

| Goal | Command |
|------|---------|
| Learn from scratch | `typer train -e` then `typer train` |
| Daily practice | `typer start` (defaults to quotes; try `--mode words` or `--mode passages`) |
| Check progress | `typer stats` |
| Retry your last run | `typer replay -l` |

Run `typer` with no arguments for an interactive menu (start, train, play, history, stats).

## Usage

```bash
typer                          # interactive menu
typer start                    # free practice (quotes by default)
typer start --mode words -w 25 # words mode, 25 words per prompt
typer train -e                 # placement test (first time)
typer train                    # continue lessons
typer train status             # training profile summary
typer history                  # recent sessions
typer stats                    # averages and common mistakes
typer replay -l                # race your newest session
typer set --show-hint off      # change settings (see typer set --help)
typer play                     # typing games
```

Command help: `typer --help`, `typer start --help`, and so on. From a clone, use `go run ./cmd/typer` instead of `typer`.

## Privacy & local data

History and settings stay on your device under your OS config directory:

- **Linux:** `~/.config/typer`
- **macOS:** `~/Library/Application Support/typer`
- **Windows:** `%AppData%\typer`

Saved sessions include what you typed so you can review them later. Training progress lives in `profile.json` in the same directory.

- Clear **session history:** `typer --reset-progress`
- Clear **training profile** (history kept): `typer train reset`

By default, practice shows a ghost from your best prior run on the same text. Pass `--no-ghost` on `typer start` to turn that off.

## Build

```bash
go build -o typer ./cmd/typer
```

Embed a version when building:

```bash
go build -ldflags "-X typer/internal/version.Version=dev" -o typer ./cmd/typer
```

## Project docs

- [Contributing](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Security](SECURITY.md)
- [Changelog](CHANGELOG.md)
- [Roadmap](ROADMAP.md)

## Credits

Run `typer credits` for full attribution. Libraries: [Bubble Tea v2](https://github.com/charmbracelet/bubbletea), [Lip Gloss v2](https://github.com/charmbracelet/lipgloss), [oklog/ulid](https://github.com/oklog/ulid).

**Bundled content:** [google-10000-english](https://github.com/first20hours/google-10000-english) words ([details](assets/WORDS_CREDITS.md)), public-domain proverbs for passages ([details](assets/PASSAGES_CREDITS.md)), and offline quotes in `assets/quotes.json`. Live quote APIs are optional — toggle with `typer set --quote-source`; bundled quotes are used when remotes are off or unavailable.

## Release checklist

- [x] Core features, tests, CI, and multi-platform release builds
- [x] Security, privacy, and contributor docs
- [x] Homebrew tap public and install verified (`isaryx/collection`)
- [x] Changelog (`CHANGELOG.md`) and agent skill (`.agents/skills/changelog/`)
- [ ] Release notes on GitHub (copy version section from `CHANGELOG.md` on each tag)
- [ ] Optional: Windows/macOS CI, shell completions, binary signing
