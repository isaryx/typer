package storage

import (
	"os"
	"path/filepath"
	"slices"

	"typer/internal/model"
)

const historyVersion = 1

// maxHistorySessions caps the retained-session count so history.json cannot
// grow without bound. Append rewrites the full file, so an unbounded log
// would eventually become a local-DoS pattern.
const maxHistorySessions = 5000

type HistoryStore struct {
	path string
}

func NewHistoryStore() (*HistoryStore, error) {
	cfg, err := appConfigDir()
	if err != nil {
		return nil, err
	}
	return NewHistoryStoreAt(filepath.Join(cfg, "history.json")), nil
}

func NewHistoryStoreAt(path string) *HistoryStore {
	return &HistoryStore{path: path}
}

func (s *HistoryStore) Append(result model.SessionResult) error {
	h, err := s.read()
	if err != nil {
		return err
	}
	h.Sessions = append(h.Sessions, result)
	if len(h.Sessions) > maxHistorySessions {
		h.Sessions = slices.Clone(h.Sessions[len(h.Sessions)-maxHistorySessions:])
	}
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

func (s *HistoryStore) Reset() error {
	return writeJSONFileAtomic(s.path, model.HistoryFile{
		Version:  historyVersion,
		Sessions: []model.SessionResult{},
	})
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

func appCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "typer"), nil
}
