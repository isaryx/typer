package storage

import (
	"fmt"
	"math/rand/v2"
	"path/filepath"
	"time"
)

const zenQuotesBatchPoolFilename = "zenquotes_batch_pool.json"

// ZenQuotesBatchPoolFile persists a dequeuable batch of ZenQuotes (/api/quotes) entries.
type ZenQuotesBatchPoolFile struct {
	FetchedAt time.Time     `json:"fetched_at,omitempty"`
	Quotes    []CachedQuote `json:"quotes"`
}

// ZenQuotesBatchPool stores quotes in the app cache dir; each PopRandom removes one entry.
type ZenQuotesBatchPool struct {
	path string
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

func (p *ZenQuotesBatchPool) Load() (ZenQuotesBatchPoolFile, error) {
	out := ZenQuotesBatchPoolFile{Quotes: []CachedQuote{}}
	if _, err := readJSONFile(p.path, &out); err != nil {
		return ZenQuotesBatchPoolFile{}, err
	}
	if out.Quotes == nil {
		out.Quotes = []CachedQuote{}
	}
	return out, nil
}

func (p *ZenQuotesBatchPool) save(file ZenQuotesBatchPoolFile) error {
	return writeJSONFileAtomic(p.path, file)
}

// Refill replaces the pool with quotes and updates FetchedAt.
func (p *ZenQuotesBatchPool) Refill(quotes []CachedQuote) error {
	if quotes == nil {
		quotes = []CachedQuote{}
	}
	return p.save(ZenQuotesBatchPoolFile{
		FetchedAt: time.Now().UTC(),
		Quotes:    quotes,
	})
}

// Len returns the number of quotes in the pool.
func (p *ZenQuotesBatchPool) Len() (int, error) {
	file, err := p.Load()
	if err != nil {
		return 0, err
	}
	return len(file.Quotes), nil
}

// PopRandom removes and returns a random quote. ok is false when the pool is empty.
func (p *ZenQuotesBatchPool) PopRandom() (CachedQuote, bool, error) {
	file, err := p.Load()
	if err != nil {
		return CachedQuote{}, false, err
	}
	n := len(file.Quotes)
	if n == 0 {
		return CachedQuote{}, false, nil
	}
	idx := rand.IntN(n)
	q := file.Quotes[idx]
	file.Quotes = append(file.Quotes[:idx], file.Quotes[idx+1:]...)
	if err := p.save(file); err != nil {
		return CachedQuote{}, false, err
	}
	return q, true, nil
}
