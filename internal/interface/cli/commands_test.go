package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	svc := domain.NewConfigService(repo, osEntryTypeResolver{repoDir: dir})
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

type osEntryTypeResolver struct {
	repoDir string
}

func (r osEntryTypeResolver) ResolveEntryType(relPath string) (domain.EntryType, error) {
	info, err := os.Stat(filepath.Join(r.repoDir, relPath))
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return domain.EntryTypeDir, nil
	}
	return domain.EntryTypeFile, nil
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

type stubWorktreesList struct {
	list   []domain.WorktreeInfo
	branch string
	force  bool
}

func (s *stubWorktreesList) BranchExists(string) (bool, error)  { return false, nil }
func (s *stubWorktreesList) CreateBranch(string) error          { return nil }
func (s *stubWorktreesList) AddWorktree(string) (string, error) { return "", nil }
func (s *stubWorktreesList) ListWorktrees() ([]domain.WorktreeInfo, error) {
	return append([]domain.WorktreeInfo{}, s.list...), nil
}
func (s *stubWorktreesList) RemoveWorktree(branch string, force bool) (string, error) {
	s.branch = branch
	s.force = force
	return "/tmp/worktrees/" + branch, nil
}

func TestRunRemoveWithoutBranchFiltersMainWorktree(t *testing.T) {
	mainDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(mainDir, ".git"), 0o755); err != nil {
		t.Fatalf("prepare main .git dir: %v", err)
	}
	featureDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(featureDir, ".git"), []byte("gitdir: /tmp/irrelevant\n"), 0o644); err != nil {
		t.Fatalf("prepare worktree .git file: %v", err)
	}

	wt := &stubWorktreesList{
		list: []domain.WorktreeInfo{
			{Branch: "refs/heads/main", Path: mainDir},
			{Branch: "refs/heads/feature/foo", Path: featureDir},
		},
	}
	selector := func(list []domain.WorktreeInfo) (domain.WorktreeInfo, error) {
		if len(list) != 1 {
			t.Fatalf("list length = %d, want 1", len(list))
		}
		if list[0].Path == mainDir {
			t.Fatalf("main worktree must be filtered out")
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
	if wt.branch != "refs/heads/feature/foo" {
		t.Fatalf("removed branch = %q, want %q", wt.branch, "refs/heads/feature/foo")
	}
}

func TestRunHelpWithDashH(t *testing.T) {
	app := &App{}

	out := captureStdout(t, func() {
		if code := app.Run([]string{"-h"}); code != 0 {
			t.Fatalf("Run returned %d", code)
		}
	})

	if !strings.Contains(out, "usage: gwm <command>") {
		t.Fatalf("help missing usage, got %q", out)
	}
	if !strings.Contains(out, "create <branch>") {
		t.Fatalf("help missing create command, got %q", out)
	}
}

func TestRunCreateHelpReturnsZero(t *testing.T) {
	app := &App{}
	if code := app.runCreate([]string{"-h"}); code != 0 {
		t.Fatalf("runCreate returned %d on -h", code)
	}
}

func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	f()
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("copy: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}
	return buf.String()
}
