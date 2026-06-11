# Roadmap

Product direction for Typer — what ships today, what to build next, and what to simplify.

## Shipped

- [x] Free practice: passages, words, and quotes (`typer start`)
- [x] Structured training: placement test, tiered lessons, adaptive weak-key drills (`typer train`)
- [x] Ghost overlay and session replay with comparison
- [x] History, aggregated stats, and per-character mistake breakdown
- [x] Typing games: defense (speed under pressure) and hangman
- [x] Custom word and passage files (`typer set`)
- [x] Local-only history, settings, and training profile

## Add next

- [ ] Timed practice in `typer start` (e.g. `--time 15|30|60|120`; timed drills exist in train lessons only today)
- [ ] Weak-key drill bridge: jump from `stats` mistake analysis into focused practice
- [ ] Show current config (`typer config` or `typer set --show`)
- [ ] One-shot custom text (`--text` or stdin pipe) without editing corpus files
- [ ] Interactive menu: replay, settings, and `train status`

## Add later

- [ ] History and stats filters (`--mode`, `--since`, `--lesson`)
- [ ] Punctuation, numbers, and symbols practice mode
- [ ] Saved defaults for mode, word count, strict, ghost, and related flags
- [ ] Shell completions (bash, zsh, fish)
- [ ] Progress trend view over recent sessions (`stats --trend` or similar)
- [ ] Alternate keyboard layouts for finger hints (Dvorak, Colemak)
- [ ] History export (JSON/JSONL)
- [ ] Documented custom lesson overlays

## Simplify or deprioritize

- [ ] Collapse quote-source UX to a sensible default chain; reduce `zenquotes-random` prominence
- [ ] Input placement presets in docs/help (`classic`, `border`, `dynamic`) instead of the full matrix
- [ ] Single primary WPM on session end; detailed metrics behind a flag
- [ ] `typer history reset` (replace global `--reset-progress` naming)
- [ ] Depromote or hide `key-press`, `--no-input`, and hangman in primary flows
- [ ] Align hangman with word-corpus speed practice, or move it out of the main `play` menu
