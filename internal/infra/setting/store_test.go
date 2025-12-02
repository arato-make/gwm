package setting

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/gwm/internal/domain"
)

func TestLoad_ReturnsDefaultWhenMissing(t *testing.T) {
	dir := t.TempDir()

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if got != domain.DefaultSettings() {
		t.Fatalf("expected default settings, got %+v", got)
	}
}

func TestLoad_ReadsFile(t *testing.T) {
	dir := t.TempDir()
	gwmDir := filepath.Join(dir, ".gwm")
	if err := os.MkdirAll(gwmDir, 0o755); err != nil {
		t.Fatalf("failed to prepare dir: %v", err)
	}

	content := `{"tmuxControlMode": true}`
	if err := os.WriteFile(filepath.Join(gwmDir, "setting.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if !got.TmuxControlMode {
		t.Fatalf("TmuxControlMode should be true, got %+v", got)
	}
}
