package storage

import (
	"path/filepath"
	"time"
)

type CachedQuote struct {
	Content string `json:"content"`
	Author  string `json:"author"`
	Source  string `json:"source"`
}

type QuoteCacheFile struct {
	FetchedAt time.Time     `json:"fetched_at"`
	Quotes    []CachedQuote `json:"quotes"`
}

type QuoteCacheStore struct {
	path string
}

const quotesCacheFilename = "quotes_cache.json"

func NewQuoteCacheStore() (*QuoteCacheStore, error) {
	cacheDir, err := appCacheDir()
	if err != nil {
		return nil, err
	}
	newPath := filepath.Join(cacheDir, quotesCacheFilename)

	return NewQuoteCacheStoreAt(newPath), nil
}

func NewQuoteCacheStoreAt(path string) *QuoteCacheStore {
	return &QuoteCacheStore{path: path}
}

func (s *QuoteCacheStore) Load() (QuoteCacheFile, error) {
	out := QuoteCacheFile{Quotes: []CachedQuote{}}
	if _, err := readJSONFile(s.path, &out); err != nil {
		return QuoteCacheFile{}, err
	}
	return out, nil
}

func (s *QuoteCacheStore) Save(quotes []CachedQuote) error {
	return writeJSONFileAtomic(s.path, QuoteCacheFile{
		FetchedAt: time.Now().UTC(),
		Quotes:    quotes,
	})
}
