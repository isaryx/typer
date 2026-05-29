package storage

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"

	"typer/internal/model"
)

// ErrNoGhostCandidate is returned when no completed session with a typing trace exists for the hash.
var ErrNoGhostCandidate = errors.New("no session with typing trace for this prompt content")

// ErrSessionNotFound is returned when GetByID finds no matching session.
var ErrSessionNotFound = errors.New("session not found")

// ErrInsufficientHistory is returned when NthNewest cannot satisfy n (not enough sessions).
var ErrInsufficientHistory = errors.New("insufficient history")

// ErrNthOutOfRange is returned when NthNewest is called with n < 1.
var ErrNthOutOfRange = errors.New("nth out of range")

const historyVersion = 1

// HistoryStore persists session results in history.json with an in-memory cache
// and optional traces sidecar (traces.json).
type HistoryStore struct {
	path       string
	tracesPath string
	cache      *historyCache
	traces     *tracesCache
}

func NewHistoryStore() (*HistoryStore, error) {
	cfg, err := appConfigDir()
	if err != nil {
		return nil, fmt.Errorf("history store: %w", err)
	}
	return NewHistoryStoreAt(filepath.Join(cfg, "history.json")), nil
}

func NewHistoryStoreAt(path string) *HistoryStore {
	dir := filepath.Dir(path)
	return &HistoryStore{
		path:       path,
		tracesPath: filepath.Join(dir, "traces.json"),
	}
}

func (s *HistoryStore) ensureLoaded() error {
	if s.cache != nil {
		return nil
	}
	file := model.HistoryFile{Version: historyVersion, Sessions: []model.SessionResult{}}
	if _, err := readJSONFile(s.path, &file); err != nil {
		return err
	}
	if file.Version == 0 {
		file.Version = historyVersion
	}
	if file.Sessions == nil {
		file.Sessions = []model.SessionResult{}
	}
	tc, err := s.loadTracesFromDisk()
	if err != nil {
		return err
	}
	s.traces = tc
	s.cache = newHistoryCache(file)
	s.migrateInlineTraces(s.cache, tc)
	return nil
}

func (s *HistoryStore) persistIfDirty() error {
	if s.cache == nil || !s.cache.dirty {
		return nil
	}
	s.cache.file.Version = historyVersion
	if err := writeJSONFileAtomicCompact(s.path, s.cache.file); err != nil {
		return err
	}
	s.cache.dirty = false
	return nil
}

func (s *HistoryStore) Append(result model.SessionResult) error {
	if err := s.ensureLoaded(); err != nil {
		return err
	}
	stored, trace := s.splitTraceForStorage(result)
	if len(trace) > 0 {
		s.traces.setTrace(stored.ID, trace)
	}
	s.cache.file.Sessions = append(s.cache.file.Sessions, stored)
	idx := len(s.cache.file.Sessions) - 1
	s.cache.indexSessionAt(idx, s.traces.file.Traces)
	s.cache.dirty = true
	s.cache.trimToMax(model.MaxRetainedHistorySessions, s.traces)
	if err := s.persistTracesIfDirty(); err != nil {
		return err
	}
	return s.persistIfDirty()
}

func (s *HistoryStore) GetByID(id string) (model.SessionResult, error) {
	if err := s.ensureLoaded(); err != nil {
		return model.SessionResult{}, err
	}
	idx, ok := s.cache.byID[id]
	if !ok {
		return model.SessionResult{}, fmt.Errorf("no session with id %q: %w", id, ErrSessionNotFound)
	}
	sess := s.cache.file.Sessions[idx]
	s.attachTrace(&sess)
	return sess, nil
}

func (s *HistoryStore) NthNewest(n int) (model.SessionResult, error) {
	if n < 1 {
		return model.SessionResult{}, fmt.Errorf("nth must be at least 1: %w", ErrNthOutOfRange)
	}
	list, err := s.List(n)
	if err != nil {
		return model.SessionResult{}, fmt.Errorf("nth newest session: %w", err)
	}
	if len(list) < n {
		return model.SessionResult{}, fmt.Errorf("only %d session(s) in history: %w", len(list), ErrInsufficientHistory)
	}
	sess := list[n-1]
	s.attachTrace(&sess)
	return sess, nil
}

func (s *HistoryStore) SessionsWithContentHash(hash string) ([]model.SessionResult, error) {
	if err := s.ensureLoaded(); err != nil {
		return nil, err
	}
	if hash == "" {
		return nil, nil
	}
	var out []model.SessionResult
	for _, idx := range s.cache.byHash[hash] {
		out = append(out, s.cache.file.Sessions[idx])
	}
	return out, nil
}

func (s *HistoryStore) BestSessionForGhost(hash string) (model.SessionResult, error) {
	if err := s.ensureLoaded(); err != nil {
		return model.SessionResult{}, err
	}
	if hash == "" {
		return model.SessionResult{}, ErrNoGhostCandidate
	}
	bestID, ok := s.cache.ghostBest[hash]
	if !ok {
		return model.SessionResult{}, ErrNoGhostCandidate
	}
	sess, err := s.GetByID(bestID)
	if err != nil {
		return model.SessionResult{}, err
	}
	return sess, nil
}

func (s *HistoryStore) List(last int) ([]model.SessionResult, error) {
	if err := s.ensureLoaded(); err != nil {
		return nil, err
	}
	start := 0
	sessions := s.cache.file.Sessions
	if last > 0 && last < len(sessions) {
		start = len(sessions) - last
	}
	out := slices.Clone(sessions[start:])
	slices.Reverse(out)
	return out, nil
}

func (s *HistoryStore) Reset() error {
	s.cache = newHistoryCache(model.HistoryFile{Version: historyVersion, Sessions: []model.SessionResult{}})
	s.traces = newTracesCache()
	s.cache.dirty = true
	s.traces.dirty = true
	if err := s.persistTracesIfDirty(); err != nil {
		return err
	}
	return s.persistIfDirty()
}

func (s *HistoryStore) read() (model.HistoryFile, error) {
	if err := s.ensureLoaded(); err != nil {
		return model.HistoryFile{}, err
	}
	return s.cache.file, nil
}
