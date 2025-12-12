package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// DetectMainWorktreeDir returns the path to the main worktree directory.
// If it cannot be determined, it returns repoDir as-is.
func DetectMainWorktreeDir(repoDir string) (string, error) {
	repoAbs, err := filepath.Abs(repoDir)
	if err != nil {
		return repoDir, err
	}
	repoAbs = canonicalizeExisting(repoAbs)

	commonDir, err := revParsePath(repoAbs, "--git-common-dir")
	if err != nil {
		return repoAbs, err
	}
	commonDir = canonicalizeExisting(commonDir)

	entries, err := listWorktreesPorcelain(repoAbs)
	if err != nil {
		return repoAbs, err
	}
	for _, e := range entries {
		gitDirAbs, err := absFromBase(repoAbs, e.gitDir)
		if err != nil {
			return repoAbs, err
		}
		gitDirAbs = canonicalizeExisting(gitDirAbs)
		if gitDirAbs == commonDir {
			mainDir, err := absFromBase(repoAbs, e.worktreeDir)
			if err != nil {
				return repoAbs, err
			}
			return canonicalizeExisting(mainDir), nil
		}
	}

	if len(entries) > 0 {
		mainDir, err := absFromBase(repoAbs, entries[0].worktreeDir)
		if err != nil {
			return repoAbs, err
		}
		return canonicalizeExisting(mainDir), nil
	}

	return repoAbs, nil
}

type worktreeEntry struct {
	worktreeDir string
	gitDir      string
}

func listWorktreesPorcelain(repoDir string) ([]worktreeEntry, error) {
	cmd := exec.Command("git", "-C", repoDir, "worktree", "list", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	sc := bufio.NewScanner(bytes.NewReader(out))
	var entries []worktreeEntry
	var cur worktreeEntry
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "worktree ") {
			if cur.worktreeDir != "" {
				entries = append(entries, cur)
			}
			cur = worktreeEntry{worktreeDir: strings.TrimPrefix(line, "worktree ")}
			continue
		}
		if strings.HasPrefix(line, "gitdir ") {
			cur.gitDir = strings.TrimPrefix(line, "gitdir ")
			continue
		}
	}
	if cur.worktreeDir != "" {
		entries = append(entries, cur)
	}
	return entries, sc.Err()
}

func revParsePath(repoDir string, arg string) (string, error) {
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", arg)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return absFromBase(repoDir, strings.TrimSpace(string(out)))
}

func absFromBase(baseDir, p string) (string, error) {
	if filepath.IsAbs(p) {
		return filepath.Clean(p), nil
	}
	abs, err := filepath.Abs(filepath.Join(baseDir, p))
	if err != nil {
		return "", fmt.Errorf("resolve path %s: %w", p, err)
	}
	return abs, nil
}

func canonicalizeExisting(p string) string {
	cp, err := filepath.EvalSymlinks(p)
	if err != nil {
		return filepath.Clean(p)
	}
	return filepath.Clean(cp)
}
