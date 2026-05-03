package storage

import (
	"fmt"
	"path/filepath"
	"strings"

	"typer/internal/model"
)

type AppSettings struct {
	WordsFile    string `json:"words_file,omitempty"`
	PassagesFile string `json:"passages_file,omitempty"`
	// ShowHint controls the typing hint line. When nil (missing from JSON), the hint is shown.
	ShowHint *bool `json:"show_hint,omitempty"`
	// InputPosition is e.g. "bottom-left", "on-top", "on-bottom-dynamic" / "obd" (omitted or empty → on-top-dynamic / otd default).
	InputPosition string `json:"input_position,omitempty"`
	// QuoteRemoteEnabled toggles remote quote APIs by registry ID (e.g. zenquotes, typefit).
	// Nil or missing key means enabled for that source.
	QuoteRemoteEnabled map[string]bool `json:"quote_remote_enabled,omitempty"`
}

// HintVisible reports whether the typing hint line should be shown. Missing config defaults to true.
func (a AppSettings) HintVisible() bool {
	if a.ShowHint == nil {
		return true
	}
	return *a.ShowHint
}

// QuoteRemoteIsEnabled reports whether a remote quote registry ID is enabled.
// Missing key or nil map defaults to true.
func (a AppSettings) QuoteRemoteIsEnabled(id string) bool {
	if a.QuoteRemoteEnabled == nil {
		return true
	}
	if v, ok := a.QuoteRemoteEnabled[id]; ok {
		return v
	}
	return true
}

// InputPlacement returns the parsed input line placement, or DefaultInputPlacement if unset or invalid.
func (a AppSettings) InputPlacement() model.InputPlacement {
	s := strings.TrimSpace(a.InputPosition)
	if s == "" {
		return model.DefaultInputPlacement()
	}
	p, err := model.ParseInputPosition(s)
	if err != nil {
		return model.DefaultInputPlacement()
	}
	return p
}

type SettingsStore struct {
	path string
}

func NewSettingsStore() (*SettingsStore, error) {
	cfg, err := appConfigDir()
	if err != nil {
		return nil, fmt.Errorf("settings store: %w", err)
	}
	return NewSettingsStoreAt(filepath.Join(cfg, "settings.json")), nil
}

func NewSettingsStoreAt(path string) *SettingsStore {
	return &SettingsStore{path: path}
}

func (s *SettingsStore) Load() (AppSettings, error) {
	var out AppSettings
	if _, err := readJSONFile(s.path, &out); err != nil {
		return AppSettings{}, err
	}
	return out, nil
}

func (s *SettingsStore) Save(settings AppSettings) error {
	return writeJSONFileAtomic(s.path, settings)
}
