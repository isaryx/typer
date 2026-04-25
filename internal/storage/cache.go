package storage

import (
	"encoding/json"
	"errors"
	"os"
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

func NewQuoteCacheStore() (*QuoteCacheStore, error) {
	cfg, err := appConfigDir()
	if err != nil {
		return nil, err
	}
	return NewQuoteCacheStoreAt(filepath.Join(cfg, "quotes_cache.json")), nil
}

func NewQuoteCacheStoreAt(path string) *QuoteCacheStore {
	return &QuoteCacheStore{
		path: path,
	}
}

func (s *QuoteCacheStore) Load() (QuoteCacheFile, error) {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return QuoteCacheFile{}, err
	}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return QuoteCacheFile{Quotes: []CachedQuote{}}, nil
	}
	if err != nil {
		return QuoteCacheFile{}, err
	}
	if len(data) == 0 {
		return QuoteCacheFile{Quotes: []CachedQuote{}}, nil
	}
	var out QuoteCacheFile
	if err := json.Unmarshal(data, &out); err != nil {
		return QuoteCacheFile{}, err
	}
	return out, nil
}

func (s *QuoteCacheStore) Save(quotes []CachedQuote) error {
	payload := QuoteCacheFile{
		FetchedAt: time.Now().UTC(),
		Quotes:    quotes,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}
