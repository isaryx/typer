# Passage attribution

`assets/passages.txt` is generated (200 passages) from English proverb sentences in Thomas Preston,
*Dictionary of English Proverbs and Proverbial Phrases* (Project Gutenberg eBook #39281).

The generator (`scripts/build_passages.py`) downloads the plain text file, strips everything outside
Gutenberg’s standard `*** START ...` / `*** END ...` markers (so the PG license blurb is not included),
then extracts the “proverb text” portion of each numbered entry (for example: `401. ...`).

Because many individual dictionary lines are short, the generator **bundles consecutive proverbs** into a
single passage until it lands in a **15–50 word** window (good for typing sessions without turning each
passage into a full page).

- Plain text (UTF-8) source: `https://www.gutenberg.org/files/39281/39281-0.txt`
- eBook page: <https://www.gutenberg.org/ebooks/39281>

## Regenerate

- `python3 scripts/build_passages.py`

## Reuse / licensing note

- Public-domain / reuse guidance (U.S.): <https://www.gutenberg.org/policy/permission>
- The “Project Gutenberg” name is a registered trademark; if you use their branding, follow Gutenberg’s
  trademark guidance.
