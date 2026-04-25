package storage

import (
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
	h.Version = historyVersion
	return writeJSONFileAtomic(s.path, h)
}

func (s *HistoryStore) List(last int) ([]model.SessionResult, error) {
	h, err := s.read()
	if err != nil {
		return nil, err
	}
	start := 0
	if last > 0 && last < len(h.Sessions) {
		start = len(h.Sessions) - last
	}
	out := slices.Clone(h.Sessions[start:])
	slices.Reverse(out)
	return out, nil
}

func (s *HistoryStore) read() (model.HistoryFile, error) {
	h := model.HistoryFile{Version: historyVersion, Sessions: []model.SessionResult{}}
	if _, err := readJSONFile(s.path, &h); err != nil {
		return model.HistoryFile{}, err
	}
	if h.Version == 0 {
		h.Version = historyVersion
	}
	return h, nil
}

func appConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "typer"), nil
}
