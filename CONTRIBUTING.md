# Contributing

Thanks for helping improve `typer`.

## Development Setup

Requirements:

- [Go](https://go.dev/dl/) — match the `go` version in [`go.mod`](go.mod) (or newer).

Clone and run locally:

```bash
go run ./cmd/typer start --mode passages
```

## Project layout

Rough map of where things live:

- `cmd/typer` — `main`, wires `internal/cli`.
- `internal/cli` — Subcommands and flags (`root.go` dispatches; other files per command).
- `internal/session` — Bubble Tea session UI; `typing_state.go` holds core word-lane typing logic (no TUI imports).
- `internal/model` — Shared types (sessions, prompts, replay trace).
- `internal/text` — Prompt providers (words, passages, quotes).
- `internal/storage` — Local JSON history, settings, quote cache.
- `internal/scoring` — WPM and related metrics.

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
