package train

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const profileFileName = "profile.json"

const (
	configDirPerm  os.FileMode = 0o700
	configFilePerm os.FileMode = 0o600
)

// ErrNoProfile is returned when profile.json does not exist.
var ErrNoProfile = errors.New("no training profile")

type ProfileStore struct {
	path string
}

func NewProfileStore() (*ProfileStore, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("profile store: %w", err)
	}
	return NewProfileStoreAt(filepath.Join(base, "typer", profileFileName)), nil
}

func NewProfileStoreAt(path string) *ProfileStore {
	return &ProfileStore{path: path}
}

func (s *ProfileStore) Path() string {
	return s.path
}

func (s *ProfileStore) Load() (Profile, error) {
	var p Profile
	exists, err := readJSONFile(s.path, &p)
	if err != nil {
		return Profile{}, err
	}
	if !exists {
		return Profile{}, ErrNoProfile
	}
	if p.Keys == nil {
		p.Keys = make(map[string]KeyStat)
	}
	return p, nil
}

func (s *ProfileStore) Save(p Profile) error {
	if p.Keys == nil {
		p.Keys = make(map[string]KeyStat)
	}
	p.Version = ProfileVersion
	return writeJSONFileAtomic(s.path, p)
}

func (s *ProfileStore) Reset() error {
	if err := os.Remove(s.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove profile: %w", err)
	}
	return nil
}

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
	_ = os.Chmod(path, configFilePerm)
	return nil
}
