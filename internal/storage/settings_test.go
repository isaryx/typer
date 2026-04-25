package storage

import (
	"path/filepath"
	"testing"
)

func TestSettingsStore_SaveLoadWordsFile(t *testing.T) {
	store := NewSettingsStoreAt(filepath.Join(t.TempDir(), "settings.json"))

	if err := store.Save(AppSettings{WordsFile: "/tmp/custom-words.txt"}); err != nil {
		t.Fatalf("save settings: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if got.WordsFile != "/tmp/custom-words.txt" {
		t.Fatalf("unexpected words file: %q", got.WordsFile)
	}
}

func TestSettingsStore_SaveLoadPassagesFile(t *testing.T) {
	store := NewSettingsStoreAt(filepath.Join(t.TempDir(), "settings.json"))

	if err := store.Save(AppSettings{PassagesFile: "/tmp/custom-passages.txt"}); err != nil {
		t.Fatalf("save settings: %v", err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if got.PassagesFile != "/tmp/custom-passages.txt" {
		t.Fatalf("unexpected passages file: %q", got.PassagesFile)
	}
}
