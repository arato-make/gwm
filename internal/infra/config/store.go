package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/example/gwm/internal/domain"
)

// Store persists configuration as JSON on local filesystem.
type Store struct {
	path    string
	repoDir string
}

// NewStore creates a Store rooted at repoDir/.gwm/config.json.
func NewStore(repoDir string) *Store {
	return &Store{path: filepath.Join(repoDir, ".gwm", "config.json"), repoDir: repoDir}
}

func (s *Store) ensureDir() error {
	return os.MkdirAll(filepath.Dir(s.path), 0o755)
}

// Load reads config entries. Empty file or missing file returns empty slice.
func (s *Store) Load() ([]domain.ConfigEntry, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return []domain.ConfigEntry{}, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []domain.ConfigEntry
	if len(data) == 0 {
		return []domain.ConfigEntry{}, nil
	}
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	for i := range entries {
		if entries[i].Type == "" {
			typ, err := detectEntryType(s.repoDir, entries[i].Path)
			if err != nil {
				return nil, err
			}
			entries[i].Type = typ
		}
	}
	return entries, nil
}

// Save writes entries atomically.
func (s *Store) Save(entries []domain.ConfigEntry) error {
	if err := s.ensureDir(); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func detectEntryType(repoDir, relPath string) (domain.EntryType, error) {
	info, err := os.Stat(filepath.Join(repoDir, relPath))
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return domain.EntryTypeDir, nil
	}
	return domain.EntryTypeFile, nil
}
