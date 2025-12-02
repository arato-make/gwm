package usecase

import (
	"fmt"

	"github.com/example/gwm/internal/domain"
)

type CreateInput struct {
	Branch string
}

type CreateOutput struct {
	Messages []string
	Worktree string
}

type CreateInteractor struct {
	Worktrees domain.WorktreeService
	Config    domain.ConfigRepository
	FileOps   domain.FileOperator
	Launcher  domain.SessionLauncher
}

func (u *CreateInteractor) Execute(in CreateInput) (CreateOutput, error) {
	var out CreateOutput

	exists, err := u.Worktrees.BranchExists(in.Branch)
	if err != nil {
		return out, err
	}
	if !exists {
		if err := u.Worktrees.CreateBranch(in.Branch); err != nil {
			return out, fmt.Errorf("branch create failed: %w", err)
		}
		out.Messages = append(out.Messages, "branch created")
	}

	path, err := u.Worktrees.AddWorktree(in.Branch)
	if err != nil {
		return out, err
	}
	out.Worktree = path
	out.Messages = append(out.Messages, "worktree added at "+path)

	entries, err := u.Config.Load()
	if err != nil {
		return out, err
	}
	if err := u.FileOps.Deploy(entries, path); err != nil {
		return out, err
	}
	out.Messages = append(out.Messages, fmt.Sprintf("%d file(s) deployed", len(entries)))

	if u.Launcher != nil {
		wt := domain.WorktreeInfo{Branch: in.Branch, Path: path}
		if err := u.Launcher.Launch(wt); err != nil {
			return out, err
		}
		out.Messages = append(out.Messages, "tmux session launched")
	}
	return out, nil
}
