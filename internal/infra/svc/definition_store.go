package svc

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/example/gwm/internal/domain"
)

// DefinitionStore persists service definitions as JSON.
type DefinitionStore struct {
	path string
}

// NewDefinitionStore creates a DefinitionStore rooted at repoDir/.gwm/services.json.
func NewDefinitionStore(repoDir string) *DefinitionStore {
	return &DefinitionStore{path: filepath.Join(repoDir, ".gwm", "services.json")}
}

func (s *DefinitionStore) ensureDir() error {
	return os.MkdirAll(filepath.Dir(s.path), 0o755)
}

// Load reads service definitions. Empty file or missing file returns empty slice.
func (s *DefinitionStore) Load() ([]domain.ServiceDefinition, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return []domain.ServiceDefinition{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return []domain.ServiceDefinition{}, nil
	}
	var defs []domain.ServiceDefinition
	if err := json.Unmarshal(data, &defs); err != nil {
		return nil, err
	}
	return defs, nil
}

// Save writes service definitions atomically.
func (s *DefinitionStore) Save(defs []domain.ServiceDefinition) error {
	if err := s.ensureDir(); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	data, err := json.MarshalIndent(defs, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
