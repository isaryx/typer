package storage

import (
	"typer/internal/model"
)

type historyCache struct {
	file      model.HistoryFile
	dirty     bool
	byID      map[string]int
	byHash    map[string][]int
	ghostBest map[string]string // content hash -> session ID
}

func newHistoryCache(file model.HistoryFile) *historyCache {
	c := &historyCache{
		file:      file,
		byID:      make(map[string]int, len(file.Sessions)),
		byHash:    make(map[string][]int),
		ghostBest: make(map[string]string),
	}
	c.rebuildIndexes(nil)
	return c
}

func (c *historyCache) rebuildIndexes(traces map[string][]model.ReplayEvent) {
	c.byID = make(map[string]int, len(c.file.Sessions))
	c.byHash = make(map[string][]int)
	c.ghostBest = make(map[string]string)
	for i := range c.file.Sessions {
		c.indexSessionAt(i, traces)
	}
}

func (c *historyCache) indexSessionAt(i int, traces map[string][]model.ReplayEvent) {
	sess := c.file.Sessions[i]
	if sess.ID != "" {
		c.byID[sess.ID] = i
	}
	hash := model.SessionContentHashKey(sess)
	if hash != "" {
		c.byHash[hash] = append(c.byHash[hash], i)
	}
	if !ghostEligible(sess, traces) {
		return
	}
	if bestID, ok := c.ghostBest[hash]; !ok {
		c.ghostBest[hash] = sess.ID
		return
	} else if bestIdx, ok := c.byID[bestID]; ok {
		best := c.file.Sessions[bestIdx]
		if model.BetterGhostCandidate(sess, best) {
			c.ghostBest[hash] = sess.ID
		}
	} else {
		c.ghostBest[hash] = sess.ID
	}
}

func ghostEligible(sess model.SessionResult, traces map[string][]model.ReplayEvent) bool {
	if sess.Aborted {
		return false
	}
	if len(sess.TypingTrace) > 0 {
		return true
	}
	if traces == nil {
		return false
	}
	return len(traces[sess.ID]) > 0
}

func (c *historyCache) trimToMax(max int, tc *tracesCache) {
	if max <= 0 || len(c.file.Sessions) <= max {
		return
	}
	evicted := c.file.Sessions[:len(c.file.Sessions)-max]
	if tc != nil {
		for _, sess := range evicted {
			if sess.ID != "" {
				if _, ok := tc.file.Traces[sess.ID]; ok {
					delete(tc.file.Traces, sess.ID)
					tc.dirty = true
				}
			}
		}
	}
	c.file.Sessions = c.file.Sessions[len(c.file.Sessions)-max:]
	var traces map[string][]model.ReplayEvent
	if tc != nil {
		traces = tc.file.Traces
	}
	c.rebuildIndexes(traces)
}
