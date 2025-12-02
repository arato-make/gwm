package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/example/gwm/internal/domain"
)

// RemoveInput represents the parameters for deleting a worktree.
type RemoveInput struct {
	Branch string
	Force  bool
}

// RemoveOutput describes the user-facing messages for the removal command.
type RemoveOutput struct {
	Messages []string
}

// RemoveInteractor deletes a git worktree and its related session (tmux など)。
type RemoveInteractor struct {
	Worktrees domain.WorktreeService
	Launcher  domain.SessionLauncher
}

func (u *RemoveInteractor) Execute(in RemoveInput) (RemoveOutput, error) {
	var out RemoveOutput

	if strings.TrimSpace(in.Branch) == "" {
		return out, errors.New("branch is required")
	}

	path, err := u.Worktrees.RemoveWorktree(in.Branch, in.Force)
	if err != nil {
		return out, err
	}
	out.Messages = append(out.Messages, fmt.Sprintf("worktree removed: %s", path))

	if u.Launcher != nil {
		wt := domain.WorktreeInfo{Branch: in.Branch, Path: path}
		if err := u.Launcher.Kill(wt); err != nil {
			return out, err
		}
		out.Messages = append(out.Messages, "session removed (if existed)")
	}

	return out, nil
}
