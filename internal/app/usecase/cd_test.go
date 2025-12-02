package usecase

import (
	"errors"
	"testing"

	"github.com/example/gwm/internal/domain"
)

type mockLauncher struct {
	called bool
	err    error
}

func (m *mockLauncher) Launch(domain.WorktreeInfo) error {
	m.called = true
	return m.err
}

func (m *mockLauncher) Kill(domain.WorktreeInfo) error { return nil }

func TestCdInteractorLaunch(t *testing.T) {
	wt := domain.WorktreeInfo{Path: "/tmp", Branch: "feature/foo"}

	t.Run("no launcher", func(t *testing.T) {
		u := &CdInteractor{}
		if err := u.Launch(wt); err == nil {
			t.Fatalf("expected error when launcher is nil")
		}
	})

	t.Run("launcher called", func(t *testing.T) {
		ml := &mockLauncher{}
		u := &CdInteractor{Launcher: ml}
		if err := u.Launch(wt); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ml.called {
			t.Fatalf("launcher not called")
		}
	})

	t.Run("launcher error bubbles up", func(t *testing.T) {
		want := errors.New("launch failed")
		ml := &mockLauncher{err: want}
		u := &CdInteractor{Launcher: ml}
		if err := u.Launch(wt); !errors.Is(err, want) {
			t.Fatalf("expected %v, got %v", want, err)
		}
	})
}
