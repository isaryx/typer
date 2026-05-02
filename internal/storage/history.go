package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"typer/internal/model"
)

// ErrNoGhostCandidate is returned when no completed session with a typing trace exists for the hash.
var ErrNoGhostCandidate = errors.New("no session with typing trace for this prompt content")

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
		return nil, fmt.Errorf("history store: %w", err)
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

// GetByID returns the session with the given id, or an error if not found.
func (s *HistoryStore) GetByID(id string) (model.SessionResult, error) {
	h, err := s.read()
	if err != nil {
		return model.SessionResult{}, err
	}
	for _, sess := range h.Sessions {
		if sess.ID == id {
			return sess, nil
		}
	}
	return model.SessionResult{}, fmt.Errorf("no session with id %q", id)
}

// NthNewest returns the n-th newest session (n=1 is most recent), matching the order of history --last.
func (s *HistoryStore) NthNewest(n int) (model.SessionResult, error) {
	if n < 1 {
		return model.SessionResult{}, fmt.Errorf("nth must be >= 1")
	}
	list, err := s.List(n)
	if err != nil {
		return model.SessionResult{}, fmt.Errorf("nth newest session: %w", err)
	}
	if len(list) < n {
		return model.SessionResult{}, fmt.Errorf("only %d session(s) in history", len(list))
	}
	return list[n-1], nil
}

// SessionsWithContentHash returns sessions whose stored or derived content hash equals hash.
func (s *HistoryStore) SessionsWithContentHash(hash string) ([]model.SessionResult, error) {
	h, err := s.read()
	if err != nil {
		return nil, err
	}
	if hash == "" {
		return nil, nil
	}
	var out []model.SessionResult
	for _, sess := range h.Sessions {
		if model.SessionContentHashKey(sess) == hash {
			out = append(out, sess)
		}
	}
	return out, nil
}

// BestSessionForGhost picks the strongest prior run for shadow replay: non-aborted, has TypingTrace,
// best by model.BetterGhostCandidate among sessions matching the content hash.
func (s *HistoryStore) BestSessionForGhost(hash string) (model.SessionResult, error) {
	sessions, err := s.SessionsWithContentHash(hash)
	if err != nil {
		return model.SessionResult{}, err
	}
	var candidates []model.SessionResult
	for _, sess := range sessions {
		if sess.Aborted || len(sess.TypingTrace) == 0 {
			continue
		}
		candidates = append(candidates, sess)
	}
	if len(candidates) == 0 {
		return model.SessionResult{}, ErrNoGhostCandidate
	}
	best := candidates[0]
	for _, c := range candidates[1:] {
		if model.BetterGhostCandidate(c, best) {
			best = c
		}
	}
	return best, nil
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
		return "", fmt.Errorf("user config dir: %w", err)
	}
	return filepath.Join(base, "typer"), nil
}

func appCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("user cache dir: %w", err)
	}
	return filepath.Join(base, "typer"), nil
}
