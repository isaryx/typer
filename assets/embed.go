package assets

import _ "embed"

var (
	//go:embed passages.txt
	Passages string

	//go:embed words.txt
	Words string

	//go:embed quotes.json
	QuotesSeed []byte
)
