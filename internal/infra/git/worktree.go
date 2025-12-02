package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/example/gwm/internal/domain"
)

// WorktreeClient implements domain.WorktreeService using git CLI.
type WorktreeClient struct {
	repoDir string
}

func NewWorktreeClient(repoDir string) *WorktreeClient {
	return &WorktreeClient{repoDir: repoDir}
}

func (c *WorktreeClient) BranchExists(branch string) (bool, error) {
	cmd := exec.Command("git", "-C", c.repoDir, "rev-parse", "--verify", "--quiet", branch)
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
		return false, nil
	}
	return false, err
}

func (c *WorktreeClient) defaultBranch() string {
	cmd := exec.Command("git", "-C", c.repoDir, "symbolic-ref", "refs/remotes/origin/HEAD")
	out, err := cmd.Output()
	if err != nil {
		// fallback
		return "main"
	}
	parts := strings.Split(strings.TrimSpace(string(out)), "/")
	return parts[len(parts)-1]
}

func (c *WorktreeClient) CreateBranch(branch string) error {
	base := c.defaultBranch()
	cmd := exec.Command("git", "-C", c.repoDir, "branch", branch, base)
	return cmd.Run()
}

func (c *WorktreeClient) AddWorktree(branch string) (string, error) {
	path := filepath.Join(c.repoDir, "worktrees", branch)
	if err := exec.Command("mkdir", "-p", filepath.Dir(path)).Run(); err != nil {
		return "", err
	}
	cmd := exec.Command("git", "-C", c.repoDir, "worktree", "add", path, branch)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git worktree add failed: %w (%s)", err, string(out))
	}
	return path, nil
}

func (c *WorktreeClient) ListWorktrees() ([]domain.WorktreeInfo, error) {
	cmd := exec.Command("git", "-C", c.repoDir, "worktree", "list", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	sc := bufio.NewScanner(bytes.NewReader(out))
	var list []domain.WorktreeInfo
	var current domain.WorktreeInfo
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "worktree ") {
			if current.Path != "" {
				list = append(list, current)
			}
			current = domain.WorktreeInfo{Path: strings.TrimPrefix(line, "worktree ")}
		} else if strings.HasPrefix(line, "branch ") {
			current.Branch = strings.TrimPrefix(line, "branch ")
		} else if strings.HasPrefix(line, "bare") {
			// ignore
		} else if strings.HasPrefix(line, "detached") {
			current.Branch = "(detached)"
		} else if strings.HasPrefix(line, "HEAD ") {
			current.IsCurrent = true
		}
	}
	if current.Path != "" {
		list = append(list, current)
	}
	return list, sc.Err()
}

func (c *WorktreeClient) RemoveWorktree(branch string, force bool) (string, error) {
	list, err := c.ListWorktrees()
	if err != nil {
		return "", err
	}

	normalized := branch
	if !strings.HasPrefix(branch, "refs/heads/") {
		normalized = "refs/heads/" + branch
	}

	var target *domain.WorktreeInfo
	for i := range list {
		if list[i].Branch == branch || list[i].Branch == normalized || strings.TrimPrefix(list[i].Branch, "refs/heads/") == branch {
			copy := list[i]
			target = &copy
			break
		}
	}
	if target == nil {
		return "", fmt.Errorf("worktree not found for branch %s", branch)
	}

	args := []string{"-C", c.repoDir, "worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, target.Path)

	cmd := exec.Command("git", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git worktree remove failed: %w (%s)", err, string(out))
	}

	return target.Path, nil
}
