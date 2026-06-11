# Implementation plan

Consolidated decisions and next steps from project review, docs work, and changelog planning.

## Decisions

| Topic | Decision | Rationale |
|-------|----------|-----------|
| Changelog format | [Keep a Changelog](https://keepachangelog.com/) in `CHANGELOG.md` | User-facing, scannable, standard for OSS |
| Changelog tooling | **Manual + agent skill** | Commit style is prose, not conventional; small release cadence |
| git-cliff | **Skip for now** | Noisy without conventional commits; revisit if commit format changes |
| Roadmap | `ROADMAP.md` (forward-looking) | Separate from changelog history |
| Release notes | Copy from `CHANGELOG.md` section → GitHub Release | Matches README release checklist |
| git-cliff later | Only if adopting `feat:` / `fix:` commits + CI generation | Optional phase 3 |

## Done (docs)

- [x] README: practical features, quick start, usage, privacy paths, credits
- [x] `ROADMAP.md`: shipped / add next / add later / simplify
- [x] `CHANGELOG.md`: initial file with `[Unreleased]` and `0.6.0`
- [x] `.agents/skills/changelog/`: agent skill for update + release cut
- [x] `IMPLEMENTATION.md`: this file

## Phase 1 — Changelog workflow (implement now)

Low-risk; unblocks release checklist.

| Step | Action | Owner |
|------|--------|-------|
| 1.1 | After user-facing changes, add bullets under `[Unreleased]` | Dev / agent (`changelog` skill) |
| 1.2 | On release: run release-cut workflow (skill or manual) | Maintainer |
| 1.3 | Tag `vX.Y.Z` → existing CI publishes binaries | CI |
| 1.4 | Paste `## [X.Y.Z]` section into GitHub Release body | Maintainer |
| 1.5 | Tick README release checklist when first release uses changelog | Maintainer |

**Agent triggers**

- “update changelog”, “changelog this”, “add to unreleased” → `changelog-update`
- “cut release 0.7.0”, “prepare release”, “release notes” → `release-cut`

**Optional follow-up (phase 1b)**

- [ ] CI: extract latest changelog section into `softprops/action-gh-release` `body`
- [x] `CONTRIBUTING.md`: changelog + release paragraph
- [ ] Backfill older versions in `CHANGELOG.md` from git tags (optional, low priority)

## Phase 2 — Product features (from ROADMAP)

Prioritized by daily-use impact. Implement one item per PR; update `CHANGELOG.md` + `ROADMAP.md` when shipped.

### 2a — Add next (high impact)

| # | Feature | CLI sketch | Main touchpoints |
|---|---------|------------|------------------|
| 1 | Timed practice in `start` | `typer start --time 60` | `internal/cli/start.go`, `internal/session/`, `model.SessionOptions` |
| 2 | Weak-key drill bridge | `typer train --drill weak` or `typer start --focus weak` | `internal/cli/`, `internal/train/`, `internal/text/provider_adaptive.go` |
| 3 | Show config | `typer config` or `typer set --show` | `internal/cli/set.go`, new `config.go` |
| 4 | One-shot custom text | `typer start --text "..."` or stdin | `internal/text/provider_local.go`, `start.go` |
| 5 | Richer interactive menu | `replay`, `set`, `train status` in `typer` menu | `internal/cli/root_select.go` |

**Suggested order:** 5 (small) → 3 (small) → 1 (medium) → 4 (medium) → 2 (medium, ties train + stats)

### 2b — Add later

- History/stats filters (`--mode`, `--since`, `--lesson`)
- Punctuation / numbers / symbols mode
- Saved defaults in `settings.json`
- Shell completions (bash, zsh, fish)
- `stats --trend` progress view
- Dvorak / Colemak finger hints
- History export (JSON/JSONL)
- Custom lesson overlay docs

### 2c — Simplify / deprioritize

Do as focused cleanup PRs; note **Changed** / **Removed** in changelog.

- Quote-source UX: one default chain; hide `zenquotes-random` from primary docs
- Input placement: document 3 presets (`classic`, `border`, `dynamic`)
- Session end: one primary WPM; details behind `--verbose` or flag
- `typer history reset` replacing `--reset-progress`
- Depromote `key-press`, `--no-input`, hangman in menus/help
- Hangman: align with word corpus or demote in `play` menu

## Phase 3 — Release & distribution

From README release checklist (not product features).

- [ ] Release notes on GitHub (process defined in phase 1)
- [ ] Windows/macOS CI (optional)
- [ ] Shell completions (also in roadmap 2b)
- [ ] Binary signing (optional)

## File map

| File | Purpose |
|------|---------|
| `README.md` | Install, quick start, usage — user entry |
| `ROADMAP.md` | What we plan to build |
| `CHANGELOG.md` | What we shipped |
| `IMPLEMENTATION.md` | This plan (maintainer) |
| `.agents/skills/changelog/SKILL.md` | Agent workflow for changelog |
| `.github/workflows/release.yml` | Build on tag; release notes TBD in 1b |

## Per-PR checklist (features)

When implementing a roadmap item:

1. Code + tests
2. `typer <cmd> --help` and README if user-facing
3. `CHANGELOG.md` → `[Unreleased]`
4. `ROADMAP.md` → move item to **Shipped** (or check box)
5. Do not commit unless asked

## Per-release checklist

1. `[Unreleased]` complete and user-facing only
2. Cut version section + date; fresh `[Unreleased]`
3. Update footer compare links in `CHANGELOG.md`
4. `git tag vX.Y.Z && git push origin vX.Y.Z`
5. GitHub Release body = that version section
6. Homebrew tap bump if applicable
