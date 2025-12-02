package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/example/gwm/internal/app/usecase"
	"github.com/example/gwm/internal/domain"
)

type memoryConfigRepo struct {
	entries []domain.ConfigEntry
}

type stubWorktrees struct {
	branch string
	force  bool
}

func (s *stubWorktrees) BranchExists(string) (bool, error)  { return false, nil }
func (s *stubWorktrees) CreateBranch(string) error          { return nil }
func (s *stubWorktrees) AddWorktree(string) (string, error) { return "", nil }
func (s *stubWorktrees) ListWorktrees() ([]domain.WorktreeInfo, error) {
	if s.branch == "" {
		return []domain.WorktreeInfo{}, nil
	}
	return []domain.WorktreeInfo{{Branch: s.branch, Path: "/tmp/worktrees/" + s.branch}}, nil
}
func (s *stubWorktrees) RemoveWorktree(branch string, force bool) (string, error) {
	s.branch = branch
	s.force = force
	return "/tmp/worktrees/" + branch, nil
}

type stubLauncher struct{}

func (stubLauncher) Launch(domain.WorktreeInfo) error { return nil }
func (stubLauncher) Kill(domain.WorktreeInfo) error   { return nil }

func (m *memoryConfigRepo) Load() ([]domain.ConfigEntry, error) {
	return append([]domain.ConfigEntry{}, m.entries...), nil
}

func (m *memoryConfigRepo) Save(entries []domain.ConfigEntry) error {
	m.entries = append([]domain.ConfigEntry{}, entries...)
	return nil
}

func TestRunConfigAddAllowsModeAfterPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(""), 0o644); err != nil {
		t.Fatalf("prepare file: %v", err)
	}
	repo := &memoryConfigRepo{}
	svc := domain.NewConfigService(repo, dir)
	app := &App{Config: &usecase.ConfigInteractor{Service: svc}}

	exit := app.runConfigAdd([]string{"AGENTS.md", "--mode", "symlink"})
	if exit != 0 {
		t.Fatalf("runConfigAdd returned %d", exit)
	}

	got, _ := repo.Load()
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].Mode != domain.ModeSymlink {
		t.Fatalf("mode = %s, want %s", got[0].Mode, domain.ModeSymlink)
	}
}

func TestRunRemoveAcceptsBranchBeforeFlag(t *testing.T) {
	wt := &stubWorktrees{}
	app := &App{Remove: &usecase.RemoveInteractor{Worktrees: wt, Launcher: stubLauncher{}}}

	if exit := app.runRemove([]string{"feature/foo", "--force"}); exit != 0 {
		t.Fatalf("runRemove returned %d", exit)
	}
	if wt.branch != "feature/foo" || !wt.force {
		t.Fatalf("unexpected input: branch=%s force=%v", wt.branch, wt.force)
	}
}

func TestRunRemoveWithoutBranchUsesSelector(t *testing.T) {
	wt := &stubWorktrees{branch: "feature/foo"}
	selector := func(list []domain.WorktreeInfo) (domain.WorktreeInfo, error) {
		if len(list) == 0 {
			return domain.WorktreeInfo{}, errors.New("empty")
		}
		return list[0], nil
	}
	app := &App{
		Remove: &usecase.RemoveInteractor{Worktrees: wt, Launcher: stubLauncher{}},
		Select: selector,
	}

	if exit := app.runRemove(nil); exit != 0 {
		t.Fatalf("runRemove returned %d", exit)
	}
}
