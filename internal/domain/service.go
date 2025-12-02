package domain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

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
	RemoveWorktree(branch string, force bool) (string, error)
}

// FileOperator deploys files into a worktree.
type FileOperator interface {
	Deploy(entries []ConfigEntry, worktreePath string) error
}

// SessionLauncher launches or attaches to a session (tmuxなど) rooted at the worktree.
type SessionLauncher interface {
	Launch(worktree WorktreeInfo) error
	Kill(worktree WorktreeInfo) error
}

// ConfigService offers add/list/remove operations on config entries.
type ConfigService struct {
	repo    ConfigRepository
	repoDir string
}

func NewConfigService(repo ConfigRepository, repoDir string) *ConfigService {
	return &ConfigService{repo: repo, repoDir: repoDir}
}

func (s *ConfigService) List() ([]ConfigEntry, error) {
	entries, err := s.repo.Load()
	if err != nil {
		return nil, err
	}
	return s.populateMissingTypes(entries)
}

func (s *ConfigService) Add(entry ConfigEntry) error {
	if err := s.assignType(&entry); err != nil {
		return err
	}
	if err := entry.Validate(); err != nil {
		return err
	}
	entries, err := s.repo.Load()
	if err != nil {
		return err
	}
	entries, err = s.populateMissingTypes(entries)
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
	entries, err = s.populateMissingTypes(entries)
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

func (s *ConfigService) assignType(entry *ConfigEntry) error {
	info, err := os.Stat(filepath.Join(s.repoDir, entry.Path))
	if err != nil {
		return err
	}
	actual := EntryTypeFile
	if info.IsDir() {
		actual = EntryTypeDir
	}
	if entry.Type == "" {
		entry.Type = actual
		return nil
	}
	if entry.Type != actual {
		return fmt.Errorf("type mismatch: %s is %s", entry.Path, actual)
	}
	return nil
}

func (s *ConfigService) populateMissingTypes(entries []ConfigEntry) ([]ConfigEntry, error) {
	for i := range entries {
		if entries[i].Type == "" {
			if err := s.assignType(&entries[i]); err != nil {
				return nil, err
			}
		}
		if err := entries[i].Validate(); err != nil {
			return nil, err
		}
	}
	return entries, nil
}
