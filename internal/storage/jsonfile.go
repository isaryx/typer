package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config files can contain typed text echoed from sessions; keep them private
// to the current user (0700 dirs, 0600 files).
const (
	configDirPerm  os.FileMode = 0o700
	configFilePerm os.FileMode = 0o600
)

// readJSONFile reads the JSON file at path into v.
// Returns (exists=false, nil) when the file is missing or empty, so callers
// can initialize default state without treating it as an error.
func readJSONFile(path string, v any) (exists bool, err error) {
	if err := os.MkdirAll(filepath.Dir(path), configDirPerm); err != nil {
		return false, fmt.Errorf("mkdir for %s: %w", path, err)
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read %s: %w", path, err)
	}
	if len(data) == 0 {
		return false, nil
	}
	if err := json.Unmarshal(data, v); err != nil {
		return false, fmt.Errorf("parse json %s: %w", path, err)
	}
	return true, nil
}

// writeJSONFileAtomic marshals v with 2-space indent and writes to path via
// a tmp-file+rename so readers never see a half-written document.
func writeJSONFileAtomic(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), configDirPerm); err != nil {
		return fmt.Errorf("mkdir for %s: %w", path, err)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json for %s: %w", path, err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, configFilePerm); err != nil {
		return fmt.Errorf("write temp %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename %s to %s: %w", tmp, path, err)
	}
	// Rename preserves the tmp file's mode, but if path pre-existed with wider
	// perms we want to tighten it. Ignore errors on platforms (e.g. Windows)
	// where chmod semantics differ.
	_ = os.Chmod(path, configFilePerm)
	return nil
}
