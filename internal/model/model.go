package model

import "time"

type Prompt struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Author  string `json:"author,omitempty"`
	Source  string `json:"source"`
	Mode    string `json:"mode"`
}

type SessionOptions struct {
	Mode            string
	Words           int
	DurationSeconds int
	Source          string
	Strict          bool
	Indefinite      bool
}

type SessionMetrics struct {
	GrossWPM    float64 `json:"gross_wpm"`
	NetWPM      float64 `json:"net_wpm"`
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
}

type HistoryFile struct {
	Version  int             `json:"version"`
	Sessions []SessionResult `json:"sessions"`
}
