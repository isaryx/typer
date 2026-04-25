package storage

import (
	"path/filepath"
	"runtime"
	"testing"
)

func typerCacheDirForTestHome(t *testing.T, home string) string {
	t.Helper()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Caches", "typer")
	case "windows":
		// os.UserCacheDir uses %LOCALAPPDATA% when set.
		localAppData := filepath.Join(home, "AppData", "Local")
		return filepath.Join(localAppData, "typer")
	default:
		// Typical Linux/XDG layout when XDG_* is not explicitly set.
		return filepath.Join(home, ".cache", "typer")
	}
}

func TestNewQuoteCacheStoreUsesCacheDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
		t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
		t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	}

	wantCacheDir := typerCacheDirForTestHome(t, home)

	store, err := NewQuoteCacheStore()
	if err != nil {
		t.Fatalf("NewQuoteCacheStore: %v", err)
	}
	if got := filepath.Dir(store.path); got != wantCacheDir {
		t.Fatalf("cache dir = %q, want %q", got, wantCacheDir)
	}
	if filepath.Base(store.path) != quotesCacheFilename {
		t.Fatalf("unexpected cache filename: %q", store.path)
	}
}
