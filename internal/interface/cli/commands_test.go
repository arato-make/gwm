package cli

import (
	"testing"

	"github.com/example/gwm/internal/app/usecase"
	"github.com/example/gwm/internal/domain"
)

type memoryConfigRepo struct {
	entries []domain.ConfigEntry
}

func (m *memoryConfigRepo) Load() ([]domain.ConfigEntry, error) {
	return append([]domain.ConfigEntry{}, m.entries...), nil
}

func (m *memoryConfigRepo) Save(entries []domain.ConfigEntry) error {
	m.entries = append([]domain.ConfigEntry{}, entries...)
	return nil
}

func TestRunConfigAddAllowsModeAfterPath(t *testing.T) {
	repo := &memoryConfigRepo{}
	svc := domain.NewConfigService(repo)
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
