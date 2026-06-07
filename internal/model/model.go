package model

import "time"

// Canonical mode identifiers persisted in history and used by providers.
const (
	ModePassage = "passage"
	ModeWords   = "words"
	ModeQuote   = "quote"
	ModeHangman  = "hangman"
	ModeDefense  = "defense"
	ModeTrain    = "train"
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
	FingerHint bool
	// NoInput hides the typed input line in the TUI and places the hint directly under the title line.
	NoInput bool
	// HideHint hides the typing hint line under the session title (default false = show hint).
	HideHint bool
	// InputPlacement positions and aligns the typed input line. The Go zero value is bottom-left;
	// CLI paths apply settings (DefaultInputPlacement / parsed input_position) before Run.
	InputPlacement InputPlacement
	// NoAudible disables terminal bell on mistakes (default false = bell on).
	NoAudible bool
	// RemoteQuoteFetchSplash, when true, shows a loading line while the first prompt loads
	// (set by typer start for quote mode with --source remote and at least one remote API enabled).
	RemoteQuoteFetchSplash bool
	// DurationMS ends the session after this many milliseconds from first keystroke (0 = until prompt complete).
	DurationMS int
	// LessonID identifies the training lesson for history (train mode).
	LessonID string
}

// SessionResultSchema is bumped when SessionResult JSON fields change incompatibly.
const SessionResultSchema = 3

// ReplayEventKind marks a single timed input in TypingTrace.
type ReplayEventKind string

const (
	ReplayEventKey       ReplayEventKind = "key"
	ReplayEventBackspace ReplayEventKind = "bs"
	ReplayEventCommit    ReplayEventKind = "commit"
)

// ReplayEvent is MS relative to session start (first keystroke). Used to replay a shadow typing animation.
type ReplayEvent struct {
	AtMS int64           `json:"at_ms"`
	Kind ReplayEventKind `json:"kind"`
	Rune string          `json:"rune,omitempty"` // single UTF-8 character when Kind==ReplayEventKey
}

// SessionOptionsSnapshot is the persisted form of SessionOptions (e.g. in history.json).
type SessionOptionsSnapshot struct {
	Mode       string `json:"mode"`
	Words      int    `json:"words"`
	Source     string `json:"source"`
	Strict     bool   `json:"strict"`
	Indefinite bool   `json:"indefinite"`
	FingerHint bool   `json:"finger_hint,omitempty"`
	LessonID   string `json:"lesson_id,omitempty"`
	DurationMS int    `json:"duration_ms,omitempty"`
}

// OptionsSnapshot returns o as a value suitable for JSON in SessionResult.
func OptionsSnapshot(o SessionOptions) SessionOptionsSnapshot {
	return SessionOptionsSnapshot{
		Mode:       o.Mode,
		Words:      o.Words,
		Source:     o.Source,
		Strict:     o.Strict,
		Indefinite: o.Indefinite,
		FingerHint: o.FingerHint,
		LessonID:   o.LessonID,
		DurationMS: o.DurationMS,
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
	ID        string         `json:"id"` // ULID (new sessions); legacy entries may use RFC3339Nano or other strings
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
	TypingTrace       []ReplayEvent          `json:"typing_trace,omitempty"`
	// ContentHash is SHA-256 hex of CanonicalPromptContent(Prompt.Content); same text → same hash (any mode).
	ContentHash string `json:"content_hash,omitempty"`
}

type HistoryFile struct {
	Version  int             `json:"version"`
	Sessions []SessionResult `json:"sessions"`
}

// SessionOptionsForReplay reconstructs typing options from a stored session.
// Indefinite is always false (replay is one passage; indefinite mode would repeat the same prompt).
func SessionOptionsForReplay(r SessionResult) SessionOptions {
	o := SessionOptions{
		Mode:       r.Prompt.Mode,
		Words:      0,
		Source:     r.Prompt.Source,
		Strict:     false,
		Indefinite: false,
	}
	if r.ResultSchema >= 1 {
		if r.Options.Mode != "" {
			o.Mode = r.Options.Mode
		}
		o.Words = r.Options.Words
		if r.Options.Source != "" {
			o.Source = r.Options.Source
		}
		o.Strict = r.Options.Strict
		o.FingerHint = r.Options.FingerHint
	}
	return o
}
