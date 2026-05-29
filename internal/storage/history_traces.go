package storage

import (
	"path/filepath"

	"typer/internal/model"
)

const tracesVersion = 1

type tracesFile struct {
	Version int                                 `json:"version"`
	Traces  map[string][]model.ReplayEvent      `json:"traces"`
}

type tracesCache struct {
	file  tracesFile
	dirty bool
}

func newTracesCache() *tracesCache {
	return &tracesCache{
		file: tracesFile{
			Version: tracesVersion,
			Traces:  map[string][]model.ReplayEvent{},
		},
	}
}

func (s *HistoryStore) tracesPathFor() string {
	if s.tracesPath != "" {
		return s.tracesPath
	}
	return filepath.Join(filepath.Dir(s.path), "traces.json")
}

func (s *HistoryStore) loadTracesFromDisk() (*tracesCache, error) {
	tc := newTracesCache()
	if _, err := readJSONFile(s.tracesPathFor(), &tc.file); err != nil {
		return nil, err
	}
	if tc.file.Version == 0 {
		tc.file.Version = tracesVersion
	}
	if tc.file.Traces == nil {
		tc.file.Traces = map[string][]model.ReplayEvent{}
	}
	return tc, nil
}

func (s *HistoryStore) ensureTracesLoaded() (*tracesCache, error) {
	if s.traces != nil {
		return s.traces, nil
	}
	tc, err := s.loadTracesFromDisk()
	if err != nil {
		return nil, err
	}
	s.traces = tc
	return tc, nil
}

func (s *HistoryStore) persistTracesIfDirty() error {
	if s.traces == nil || !s.traces.dirty {
		return nil
	}
	s.traces.file.Version = tracesVersion
	if err := writeJSONFileAtomicCompact(s.tracesPathFor(), s.traces.file); err != nil {
		return err
	}
	s.traces.dirty = false
	return nil
}

func (tc *tracesCache) setTrace(id string, trace []model.ReplayEvent) {
	if id == "" || len(trace) == 0 {
		return
	}
	tc.file.Traces[id] = append([]model.ReplayEvent(nil), trace...)
	tc.dirty = true
}

func (tc *tracesCache) traceFor(id string) []model.ReplayEvent {
	if id == "" || tc == nil {
		return nil
	}
	return tc.file.Traces[id]
}

func (s *HistoryStore) migrateInlineTraces(c *historyCache, tc *tracesCache) {
	if c == nil || tc == nil {
		return
	}
	migrated := false
	for i := range c.file.Sessions {
		sess := &c.file.Sessions[i]
		if len(sess.TypingTrace) == 0 {
			continue
		}
		tc.setTrace(sess.ID, sess.TypingTrace)
		sess.TypingTrace = nil
		migrated = true
	}
	if migrated {
		c.rebuildIndexes(tc.file.Traces)
		c.dirty = true
		tc.dirty = true
	}
}

func (s *HistoryStore) attachTrace(sess *model.SessionResult) {
	if sess == nil || len(sess.TypingTrace) > 0 {
		return
	}
	tc, err := s.ensureTracesLoaded()
	if err != nil || tc == nil {
		return
	}
	if trace := tc.traceFor(sess.ID); len(trace) > 0 {
		sess.TypingTrace = append([]model.ReplayEvent(nil), trace...)
	}
}

func (s *HistoryStore) splitTraceForStorage(result model.SessionResult) (model.SessionResult, []model.ReplayEvent) {
	stored := result
	trace := append([]model.ReplayEvent(nil), result.TypingTrace...)
	stored.TypingTrace = nil
	return stored, trace
}
