package model

import "time"

// Canonical mode identifiers persisted in history and used by providers.
const (
	ModePassage = "passage"
	ModeWords   = "words"
	ModeQuote   = "quote"
)

type Prompt struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Author  string `json:"author,omitempty"`
	Source  string `json:"source"`
	Mode    string `json:"mode"`
}

type SessionOptions struct {
	Mode       string
	Words      int
	Source     string
	Strict     bool
	Indefinite bool
}

// SessionResultSchema is bumped when SessionResult JSON fields change incompatibly.
const SessionResultSchema = 1

// SessionOptionsSnapshot is the persisted form of SessionOptions (e.g. in history.json).
type SessionOptionsSnapshot struct {
	Mode       string `json:"mode"`
	Words      int    `json:"words"`
	Source     string `json:"source"`
	Strict     bool   `json:"strict"`
	Indefinite bool   `json:"indefinite"`
}

// OptionsSnapshot returns o as a value suitable for JSON in SessionResult.
func OptionsSnapshot(o SessionOptions) SessionOptionsSnapshot {
	return SessionOptionsSnapshot{
		Mode:       o.Mode,
		Words:      o.Words,
		Source:     o.Source,
		Strict:     o.Strict,
		Indefinite: o.Indefinite,
	}
}

type SessionMetrics struct {
	GrossWPM    float64 `json:"gross_wpm"`
	NetWPM      float64 `json:"net_wpm"`
	AdjustedWPM float64 `json:"adjusted_wpm"`
	Accuracy    float64 `json:"accuracy"`
	Consistency float64 `json:"consistency"`
	Errors      int     `json:"errors"`
}

type SessionResult struct {
	ID        string         `json:"id"`
	StartedAt time.Time      `json:"started_at"`
	EndedAt   time.Time      `json:"ended_at"`
	ElapsedMS int64          `json:"elapsed_ms"`
	Prompt    Prompt         `json:"prompt"`
	TypedText string         `json:"typed_text"`
	Metrics   SessionMetrics `json:"metrics"`
	Aborted   bool           `json:"aborted,omitempty"`

	ResultSchema      int                    `json:"result_schema,omitempty"`
	TyperVersion      string                 `json:"typer_version,omitempty"`
	Options           SessionOptionsSnapshot `json:"options,omitempty"`
	TotalKeystrokes   int                    `json:"total_keystrokes,omitempty"`
	CorrectKeystrokes int                    `json:"correct_keystrokes,omitempty"`
	WPMSamples        []float64              `json:"wpm_samples,omitempty"`
}

type HistoryFile struct {
	Version  int             `json:"version"`
	Sessions []SessionResult `json:"sessions"`
}
