// Package ui holds shared terminal styling primitives for CLI and session TUI.
package ui

// Palette: ANSI 256-color indices except BorderHex (true color).
const (
	ColorTitle        = "70"      // muted green
	ColorMeta         = "245"     // dim gray
	ColorBorderHex    = "#ffffff" // passage / metrics frame
	ColorCompletedOK  = "70"
	ColorCompletedBad = "203"
	ColorActiveFg     = "45"
	ColorUpcoming       = "252"
	ColorInputFg      = "229"
	ColorInputBadFg   = "203"
	ColorErrorFg      = "203"
)

// MaxContentWidth caps layout width on wide terminals (session TUI uses the same cap).
const MaxContentWidth = 88
