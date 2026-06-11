---
name: changelog
description: >-
  Maintain CHANGELOG.md for Typer using Keep a Changelog format. Use when updating
  the changelog, cutting a release, writing GitHub release notes, or when the user
  mentions changelog, release notes, unreleased, or versioning.
---

# Typer changelog

## Files

| File | Role |
|------|------|
| `CHANGELOG.md` | User-facing history ([Keep a Changelog](https://keepachangelog.com/)) |
| `ROADMAP.md` | Forward plan — do not duplicate future work in changelog |
| `IMPLEMENTATION.md` | Maintainer phases and per-PR/release checklists |

## Categories

Use only when relevant: **Added**, **Changed**, **Deprecated**, **Removed**, **Fixed**, **Security**.

## Update workflow (`changelog-update`)

Trigger: user shipped a feature, fixed a bug, or asks to update the changelog.

1. Read `CHANGELOG.md` `[Unreleased]` and `git log` / `git diff` since last tag (`git describe --tags --abbrev=0`).
2. Append bullets under `[Unreleased]` only — never edit released sections without explicit request.
3. Write for **CLI users**: command names, flags, behavior changes.
4. **Skip**: tests-only, CI, refactors, internal renames unless contributor-facing.
5. One bullet per user-visible change; merge duplicates.
6. Do not commit unless the user asks.

### Typer-specific wording

- Name commands: `typer start`, `typer train`, not internal package paths.
- Distinguish **session history** (`history.json`, `--reset-progress`) vs **training profile** (`profile.json`, `typer train reset`).
- Breaking flag or default changes → **Changed** with migration note.

## Release cut workflow (`release-cut`)

Trigger: user says cut/prepare/ship release `X.Y.Z`.

1. Confirm semver version with user if unclear.
2. Review `[Unreleased]`; ensure every entry is user-facing; use `git log` since last tag to catch gaps.
3. Rename `## [Unreleased]` → `## [X.Y.Z] - YYYY-MM-DD` (today, UTC or local per user preference).
4. Insert new `## [Unreleased]` with empty category headers (or omit empty categories).
5. Update footer links at bottom of `CHANGELOG.md`:
   - `[Unreleased]: https://github.com/isaryx/typer/compare/vX.Y.Z...HEAD`
   - `[X.Y.Z]: https://github.com/isaryx/typer/releases/tag/vX.Y.Z`
6. Output **GitHub Release notes** markdown (copy of the new version section, no `[Unreleased]`).
7. Remind maintainer checklist from `IMPLEMENTATION.md` (tag, push, release body, Homebrew if needed).
8. Suggest updating `ROADMAP.md` shipped items if a roadmap feature landed.
9. Do not `git tag` or push unless the user explicitly requests.

## Do not use git-cliff

Typer uses manual changelog + this skill. Do not add git-cliff unless the user adopts conventional commits and asks for automation.

## Examples

**After adding `--time 60` to start:**

```markdown
### Added
- Timed practice in `typer start` (`--time` accepts seconds, e.g. `--time 60`)
```

**Release cut 0.7.0:** move Unreleased content to `## [0.7.0] - 2026-06-10`, fresh Unreleased, update links, paste section for GitHub Release.
