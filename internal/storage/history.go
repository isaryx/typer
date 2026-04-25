package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"

	"typer/internal/model"
)

const historyVersion = 1

type HistoryStore struct {
	path string
}

func NewHistoryStore() (*HistoryStore, error) {
	cfg, err := appConfigDir()
	if err != nil {
		return nil, err
	}

	return &HistoryStore{
		path: filepath.Join(cfg, "history.json"),
	}, nil
}

func (s *HistoryStore) Append(result model.SessionResult) error {
	h, err := s.read()
	if err != nil {
		return err
	}
	h.Sessions = append(h.Sessions, result)
	return s.write(h)
}

func (s *HistoryStore) List(last int) ([]model.SessionResult, error) {
	h, err := s.read()
	if err != nil {
		return nil, err
	}
	if last <= 0 || last >= len(h.Sessions) {
		out := slices.Clone(h.Sessions)
		slices.Reverse(out)
		return out, nil
	}
	start := len(h.Sessions) - last
	out := slices.Clone(h.Sessions[start:])
	slices.Reverse(out)
	return out, nil
}

func (s *HistoryStore) read() (model.HistoryFile, error) {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return model.HistoryFile{}, err
	}

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return model.HistoryFile{Version: historyVersion, Sessions: []model.SessionResult{}}, nil
	}
	if err != nil {
		return model.HistoryFile{}, err
	}
	if len(data) == 0 {
		return model.HistoryFile{Version: historyVersion, Sessions: []model.SessionResult{}}, nil
	}

	var h model.HistoryFile
	if err := json.Unmarshal(data, &h); err != nil {
		return model.HistoryFile{}, err
	}
	if h.Version == 0 {
		h.Version = historyVersion
	}
	return h, nil
}

func (s *HistoryStore) write(h model.HistoryFile) error {
	h.Version = historyVersion
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func appConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "typer"), nil
}
