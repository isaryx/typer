# Contributing

Thanks for helping improve `typer`.

## Development Setup

Requirements:

- Go 1.24+

Clone and run locally:

```bash
go run ./cmd/typer start --mode passages
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
