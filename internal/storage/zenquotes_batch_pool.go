package storage

import (
	"fmt"
	"math/rand/v2"
	"path/filepath"
	"sync"
	"time"
)

const zenQuotesBatchPoolFilename = "zenquotes_batch_pool.json"

// ZenQuotesBatchPoolFile persists a dequeuable batch of ZenQuotes (/api/quotes) entries.
type ZenQuotesBatchPoolFile struct {
	FetchedAt time.Time     `json:"fetched_at,omitempty"`
	Quotes    []CachedQuote `json:"quotes"`
}

// ZenQuotesBatchPool stores quotes in the app cache dir; each PopRandom removes one entry.
// Quotes are held in memory after the first load; disk is written on Refill and when the pool is emptied.
type ZenQuotesBatchPool struct {
	path string

	mu     sync.Mutex
	loaded bool
	file   ZenQuotesBatchPoolFile
}

func NewZenQuotesBatchPool() (*ZenQuotesBatchPool, error) {
	cacheDir, err := appCacheDir()
	if err != nil {
		return nil, fmt.Errorf("zenquotes batch pool: %w", err)
	}
	return NewZenQuotesBatchPoolAt(filepath.Join(cacheDir, zenQuotesBatchPoolFilename)), nil
}

func NewZenQuotesBatchPoolAt(path string) *ZenQuotesBatchPool {
	return &ZenQuotesBatchPool{path: path}
}

func (p *ZenQuotesBatchPool) loadFromDisk() (ZenQuotesBatchPoolFile, error) {
	out := ZenQuotesBatchPoolFile{Quotes: []CachedQuote{}}
	if _, err := readJSONFile(p.path, &out); err != nil {
		return ZenQuotesBatchPoolFile{}, err
	}
	if out.Quotes == nil {
		out.Quotes = []CachedQuote{}
	}
	return out, nil
}

// Load reads the pool file from disk without updating the in-memory cache.
func (p *ZenQuotesBatchPool) Load() (ZenQuotesBatchPoolFile, error) {
	return p.loadFromDisk()
}

func (p *ZenQuotesBatchPool) save(file ZenQuotesBatchPoolFile) error {
	return writeJSONFileAtomic(p.path, file)
}

func (p *ZenQuotesBatchPool) ensureLoadedLocked() error {
	if p.loaded {
		return nil
	}
	file, err := p.loadFromDisk()
	if err != nil {
		return err
	}
	p.file = file
	p.loaded = true
	return nil
}

// Refill replaces the pool with quotes and updates FetchedAt.
func (p *ZenQuotesBatchPool) Refill(quotes []CachedQuote) error {
	if quotes == nil {
		quotes = []CachedQuote{}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.file = ZenQuotesBatchPoolFile{
		FetchedAt: time.Now().UTC(),
		Quotes:    quotes,
	}
	p.loaded = true
	return p.save(p.file)
}

// Len returns the number of quotes in the pool.
func (p *ZenQuotesBatchPool) Len() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.ensureLoadedLocked(); err != nil {
		return 0, err
	}
	return len(p.file.Quotes), nil
}

// PopRandom removes and returns a random quote. ok is false when the pool is empty.
func (p *ZenQuotesBatchPool) PopRandom() (CachedQuote, bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.ensureLoadedLocked(); err != nil {
		return CachedQuote{}, false, err
	}
	n := len(p.file.Quotes)
	if n == 0 {
		return CachedQuote{}, false, nil
	}
	idx := rand.IntN(n)
	q := p.file.Quotes[idx]
	p.file.Quotes = append(p.file.Quotes[:idx], p.file.Quotes[idx+1:]...)
	if len(p.file.Quotes) == 0 {
		if err := p.save(p.file); err != nil {
			return CachedQuote{}, false, err
		}
	}
	return q, true, nil
}
