package usecase

import (
	"errors"
	"reflect"
	"testing"

	"github.com/example/gwm/internal/domain"
)

type fakeWorktreeService struct {
	removedBranch string
	force         bool
	path          string
	err           error
}

func (f *fakeWorktreeService) BranchExists(string) (bool, error)  { return false, nil }
func (f *fakeWorktreeService) CreateBranch(string) error          { return nil }
func (f *fakeWorktreeService) AddWorktree(string) (string, error) { return "", nil }
func (f *fakeWorktreeService) ListWorktrees() ([]domain.WorktreeInfo, error) {
	return nil, nil
}
func (f *fakeWorktreeService) RemoveWorktree(branch string, force bool) (string, error) {
	f.removedBranch = branch
	f.force = force
	return f.path, f.err
}

type fakeLauncher struct {
	killed []domain.WorktreeInfo
	err    error
}

func (l *fakeLauncher) Launch(domain.WorktreeInfo) error { return nil }
func (l *fakeLauncher) Kill(wt domain.WorktreeInfo) error {
	l.killed = append(l.killed, wt)
	return l.err
}

func TestRemoveInteractorSuccess(t *testing.T) {
	wt := &fakeWorktreeService{path: "/tmp/worktrees/feature"}
	launcher := &fakeLauncher{}
	u := &RemoveInteractor{Worktrees: wt, Launcher: launcher}

	out, err := u.Execute(RemoveInput{Branch: "feature/foo", Force: true})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if wt.removedBranch != "feature/foo" || !wt.force {
		t.Fatalf("unexpected remove call: branch=%s force=%v", wt.removedBranch, wt.force)
	}
	if !reflect.DeepEqual(launcher.killed, []domain.WorktreeInfo{{Branch: "feature/foo", Path: "/tmp/worktrees/feature"}}) {
		t.Fatalf("Kill called with %+v", launcher.killed)
	}
	if len(out.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(out.Messages))
	}
}

func TestRemoveInteractorRequiresBranch(t *testing.T) {
	u := &RemoveInteractor{Worktrees: &fakeWorktreeService{}}

	if _, err := u.Execute(RemoveInput{}); err == nil {
		t.Fatalf("expected error for empty branch")
	}
}

func TestRemoveInteractorKillError(t *testing.T) {
	wt := &fakeWorktreeService{path: "/tmp/worktrees/feature"}
	launcher := &fakeLauncher{err: errors.New("kill failed")}
	u := &RemoveInteractor{Worktrees: wt, Launcher: launcher}

	if _, err := u.Execute(RemoveInput{Branch: "feature", Force: false}); err == nil {
		t.Fatalf("expected error when Kill fails")
	}
}
