package usecase

import "github.com/example/gwm/internal/domain"

type CdInteractor struct {
	Worktrees domain.WorktreeService
}

func (u *CdInteractor) List() ([]domain.WorktreeInfo, error) {
	return u.Worktrees.ListWorktrees()
}
