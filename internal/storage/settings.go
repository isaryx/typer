package storage

import (
	"fmt"
	"path/filepath"
)

type AppSettings struct {
	WordsFile    string `json:"words_file,omitempty"`
	PassagesFile string `json:"passages_file,omitempty"`
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
