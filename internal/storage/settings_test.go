package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const settingsJSONFile = "settings.json"

func typerConfigDirForTestHome(home string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "typer")
	case "windows":
		roamingAppData := filepath.Join(home, "AppData", "Roaming")
		return filepath.Join(roamingAppData, "typer")
	default:
		return filepath.Join(home, ".config", "typer")
	}
}

func TestNewSettingsStoreUsesConfigDir(t *testing.T) {
	home := t.TempDir()
	setTestUserDirs(t, home)

	store, err := NewSettingsStore()
	if err != nil {
		t.Fatalf("NewSettingsStore: %v", err)
	}
	want := filepath.Join(typerConfigDirForTestHome(home), settingsJSONFile)
	if store.path != want {
		t.Fatalf("settings path = %q, want %q", store.path, want)
	}
}

func TestSettingsStoreSaveLoadWordsFile(t *testing.T) {
	store := NewSettingsStoreAt(filepath.Join(t.TempDir(), settingsJSONFile))

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

func TestSettingsStoreSaveLoadPassagesFile(t *testing.T) {
	store := NewSettingsStoreAt(filepath.Join(t.TempDir(), settingsJSONFile))

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

func TestAppSettingsHintVisible(t *testing.T) {
	var a AppSettings
	if !a.HintVisible() {
		t.Fatal("missing show_hint should default to visible")
	}
	v := true
	a.ShowHint = &v
	if !a.HintVisible() {
		t.Fatal("show_hint true")
	}
	f := false
	a.ShowHint = &f
	if a.HintVisible() {
		t.Fatal("show_hint false")
	}
}

func TestSettingsStoreLoadMissingFileReturnsDefaults(t *testing.T) {
	store := NewSettingsStoreAt(filepath.Join(t.TempDir(), settingsJSONFile))
	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != (AppSettings{}) {
		t.Fatalf("expected zero-value settings, got %#v", got)
	}
}

func TestSettingsStoreLoadMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), settingsJSONFile)
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatalf("write malformed settings: %v", err)
	}
	store := NewSettingsStoreAt(path)
	if _, err := store.Load(); err == nil {
		t.Fatal("expected load error for malformed settings JSON")
	}
}
