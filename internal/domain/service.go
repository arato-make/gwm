package domain

import "errors"

// ConfigRepository persists config entries.
type ConfigRepository interface {
	Load() ([]ConfigEntry, error)
	Save([]ConfigEntry) error
}

// WorktreeService abstracts git worktree operations.
type WorktreeService interface {
	BranchExists(branch string) (bool, error)
	CreateBranch(branch string) error
	AddWorktree(branch string) (string, error)
	ListWorktrees() ([]WorktreeInfo, error)
}

// FileOperator deploys files into a worktree.
type FileOperator interface {
	Deploy(entries []ConfigEntry, worktreePath string) error
}

// ConfigService offers add/list/remove operations on config entries.
type ConfigService struct {
	repo ConfigRepository
}

func NewConfigService(repo ConfigRepository) *ConfigService {
	return &ConfigService{repo: repo}
}

func (s *ConfigService) List() ([]ConfigEntry, error) {
	return s.repo.Load()
}

func (s *ConfigService) Add(entry ConfigEntry) error {
	if err := entry.Validate(); err != nil {
		return err
	}
	entries, err := s.repo.Load()
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Path == entry.Path {
			return errors.New("entry already exists")
		}
	}
	entries = append(entries, entry)
	return s.repo.Save(entries)
}

func (s *ConfigService) Remove(path string) error {
	entries, err := s.repo.Load()
	if err != nil {
		return err
	}
	kept := entries[:0]
	found := false
	for _, e := range entries {
		if e.Path == path {
			found = true
			continue
		}
		kept = append(kept, e)
	}
	if !found {
		return errors.New("entry not found")
	}
	return s.repo.Save(kept)
}
