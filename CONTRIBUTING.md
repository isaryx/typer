# Contributing

Thanks for helping improve `typer`.

## Development Setup

Requirements:

- [Go](https://go.dev/dl/) — match the `go` version in [`go.mod`](go.mod) (or newer).

Clone and run locally:

```bash
go run ./cmd/typer              # interactive command menu
go run ./cmd/typer start --mode passages
```

## Project layout

Rough map of where things live:

- `cmd/typer` — `main`, wires `internal/cli`.
- `internal/cli` — Subcommands and flags (`root.go` dispatches; other files per command).
- `internal/session` — Bubble Tea v2 session UI; `typing_state.go` holds core word-lane typing logic (no TUI imports).
- `internal/model` — Shared types (sessions, prompts, replay trace).
- `internal/text` — Prompt providers (words, passages, quotes). Word lists load via `LoadWordCorpus` (`words_corpus.go`); words mode stratifies prompts by length bucket (`provider_words.go`).
- `internal/train` — Structured training: curriculum (`assets/lessons.json`), placement, lesson pass/fail, adaptive drills, and `profile.json` persistence.
- `internal/storage` — Local JSON history, settings, quote cache.
  - `history.json` — session metrics and metadata (typing traces stored separately).
  - `traces.json` — sidecar map of session ID → replay trace events (ghost/replay only).
- `internal/scoring` — WPM and related metrics.
- `internal/game` — Game modes (`hangman`, `defense`). Defense loads the shared word corpus via `LoadWordPool` (`defense/words_loader.go`).

## Performance

Runtime typing performance is handled in `internal/session` (layout cache, keystroke path). Storage uses an in-memory history cache, compact JSON writes, and trace sidecar I/O so list/stats commands avoid parsing large trace payloads. Terminal rendering relies on Bubble Tea cell diffing rather than app-level View caching.

Benchmarks:

```bash
go test -bench=. -benchmem ./internal/storage/... ./internal/analytics/...
```

## Test and Quality Checks

Before opening a PR, run:

```bash
go test ./...
go vet ./...
```

## Pull Requests

- Keep changes focused and small where possible.
- Include tests for behavior changes.
- Update docs when CLI behavior, flags, or defaults change.
- Use clear commit messages that explain why the change is needed.

## Release Notes and Tags

- Releases are cut from tags in the form `v*` (for example `v0.0.4`).
- If your change affects packaging, release behavior, or user-facing output, include a short note in the PR description.
