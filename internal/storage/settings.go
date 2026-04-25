package storage

import (
	"encoding/json"
	"errors"
	"os"
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
		return nil, err
	}
	return NewSettingsStoreAt(filepath.Join(cfg, "settings.json")), nil
}

func NewSettingsStoreAt(path string) *SettingsStore {
	return &SettingsStore{
		path: path,
	}
}

func (s *SettingsStore) Load() (AppSettings, error) {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return AppSettings{}, err
	}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return AppSettings{}, nil
	}
	if err != nil {
		return AppSettings{}, err
	}
	if len(data) == 0 {
		return AppSettings{}, nil
	}
	var out AppSettings
	if err := json.Unmarshal(data, &out); err != nil {
		return AppSettings{}, err
	}
	return out, nil
}

func (s *SettingsStore) Save(settings AppSettings) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}
