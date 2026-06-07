Word list attribution
=====================

The file `words.txt` is sourced from:

- Repository: https://github.com/first20hours/google-10000-english
- Source list: https://raw.githubusercontent.com/first20hours/google-10000-english/master/google-10000-english-no-swears.txt

Upstream license notes (from the source project):

- Data is derived from the Google Web Trillion Word Corpus (LDC2006T13), with subsets distributed by Peter Norvig.
- Cleanup/editing by Josh Kaufman.
- Educational and personal/research use is permitted under the noted licenses and fair use.
- Commercial use may require licensing from the Linguistic Data Consortium (LDC).

Bundled cleanup
---------------

The bundled list keeps upstream frequency order and omits web-corpus outliers (single-letter tokens,
abbreviations, tech and brand tokens, given names and place names).

Usage in typer
--------------

- **Words mode** (`typer start --mode words`): each prompt mixes word lengths (roughly 20% short, 40% core, 27% mid, 13% long) while drawing from this frequency-ranked list.
- **Defense** (`typer play defense`): uses the same list, filtered to 3–12 characters; shorter words appear first and longer words unlock as your score rises.
- **Custom list:** `typer set --words-file PATH` replaces the bundled list for both modes.
