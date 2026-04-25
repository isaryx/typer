package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// readJSONFile reads the JSON file at path into v.
// Returns (exists=false, nil) when the file is missing or empty, so callers
// can initialize default state without treating it as an error.
func readJSONFile(path string, v any) (exists bool, err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if len(data) == 0 {
		return false, nil
	}
	if err := json.Unmarshal(data, v); err != nil {
		return false, err
	}
	return true, nil
}

// writeJSONFileAtomic marshals v with 2-space indent and writes to path via
// a tmp-file+rename so readers never see a half-written document.
func writeJSONFileAtomic(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
