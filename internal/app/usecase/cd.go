package usecase

import (
	"fmt"

	"github.com/example/gwm/internal/domain"
)

type CdInteractor struct {
	Worktrees domain.WorktreeService
	Launcher  domain.SessionLauncher
}

func (u *CdInteractor) List() ([]domain.WorktreeInfo, error) {
	return u.Worktrees.ListWorktrees()
}

// Launch opens the selected worktree via configured launcher (tmux or fallback).
func (u *CdInteractor) Launch(wt domain.WorktreeInfo) error {
	if u.Launcher == nil {
		return fmt.Errorf("no session launcher configured")
	}
	return u.Launcher.Launch(wt)
}
