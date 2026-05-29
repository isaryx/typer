package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

func appConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	return filepath.Join(base, "typer"), nil
}

func appCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("user cache dir: %w", err)
	}
	return filepath.Join(base, "typer"), nil
}
