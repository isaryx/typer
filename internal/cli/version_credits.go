package cli

import (
	"fmt"
	"io"

	"typer/internal/version"
)

const creditsMessage = `Data Credits
------------
- Words mode default list (assets/words.txt): first20hours/google-10000-english
- Original corpus source: Google Web Trillion Word Corpus (via LDC), cleaned by Josh Kaufman; subsets by Peter Norvig
- Passages mode bundled list (assets/passages.txt): Thomas Preston's Dictionary of English Proverbs and Proverbial Phrases (Project Gutenberg #39281)
- Quotes mode bundled list (assets/quotes.json): curated from dwyl/quotes
- Optional remote quote APIs (when --source=remote): https://zenquotes.io/api/ (primary; free tier requires attribution per https://docs.zenquotes.io/zenquotes-documentation/) and https://type.fit/api/quotes (fallback). Toggle sources with: typer set --quote-source ID=on|off

Libraries
---------
- Session identifiers (ULID): github.com/oklog/ulid/v2 — https://github.com/oklog/ulid (Apache-2.0)

See README "Credits" for bundled content notes and library links.
`

func runVersion(out io.Writer) error {
	_, err := fmt.Fprintf(out, "typer %s\n", version.Version)
	return err
}

func runCredits(out io.Writer) error {
	_, err := fmt.Fprint(out, creditsMessage)
	return err
}
