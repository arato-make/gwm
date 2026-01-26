package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/gwm/internal/domain"
)

func TestOperatorDeploySubstituteFile(t *testing.T) {
	repoDir := t.TempDir()
	worktreeDir := t.TempDir()

	content := `{
  "server": "` + repoDir + `/server",
  "other": "value"
}`
	if err := os.WriteFile(filepath.Join(repoDir, "config.json"), []byte(content), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	op := NewOperator(repoDir)
	entries := []domain.ConfigEntry{
		{Path: "config.json", Mode: domain.ModeSubstitute, Type: domain.EntryTypeFile},
	}
	if err := op.Deploy(entries, worktreeDir); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(worktreeDir, "config.json"))
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}

	want := `{
  "server": "` + worktreeDir + `/server",
  "other": "value"
}`
	if string(got) != want {
		t.Errorf("content mismatch:\ngot:  %s\nwant: %s", string(got), want)
	}
}

func TestOperatorDeploySubstituteDir(t *testing.T) {
	repoDir := t.TempDir()
	worktreeDir := t.TempDir()

	subDir := filepath.Join(repoDir, ".mcp")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	content1 := `path: ` + repoDir + `/data`
	if err := os.WriteFile(filepath.Join(subDir, "a.yml"), []byte(content1), 0o644); err != nil {
		t.Fatalf("write a.yml: %v", err)
	}

	content2 := `root: ` + repoDir
	if err := os.WriteFile(filepath.Join(subDir, "b.txt"), []byte(content2), 0o644); err != nil {
		t.Fatalf("write b.txt: %v", err)
	}

	op := NewOperator(repoDir)
	entries := []domain.ConfigEntry{
		{Path: ".mcp", Mode: domain.ModeSubstitute, Type: domain.EntryTypeDir},
	}
	if err := op.Deploy(entries, worktreeDir); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	got1, err := os.ReadFile(filepath.Join(worktreeDir, ".mcp", "a.yml"))
	if err != nil {
		t.Fatalf("read a.yml: %v", err)
	}
	want1 := `path: ` + worktreeDir + `/data`
	if string(got1) != want1 {
		t.Errorf("a.yml mismatch:\ngot:  %s\nwant: %s", string(got1), want1)
	}

	got2, err := os.ReadFile(filepath.Join(worktreeDir, ".mcp", "b.txt"))
	if err != nil {
		t.Fatalf("read b.txt: %v", err)
	}
	want2 := `root: ` + worktreeDir
	if string(got2) != want2 {
		t.Errorf("b.txt mismatch:\ngot:  %s\nwant: %s", string(got2), want2)
	}
}

func TestOperatorDeployCopy(t *testing.T) {
	repoDir := t.TempDir()
	worktreeDir := t.TempDir()

	content := "hello world"
	if err := os.WriteFile(filepath.Join(repoDir, "test.txt"), []byte(content), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	op := NewOperator(repoDir)
	entries := []domain.ConfigEntry{
		{Path: "test.txt", Mode: domain.ModeCopy, Type: domain.EntryTypeFile},
	}
	if err := op.Deploy(entries, worktreeDir); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(worktreeDir, "test.txt"))
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != content {
		t.Errorf("content mismatch: got %s, want %s", string(got), content)
	}
}

func TestOperatorDeploySymlink(t *testing.T) {
	repoDir := t.TempDir()
	worktreeDir := t.TempDir()

	srcPath := filepath.Join(repoDir, "link.txt")
	if err := os.WriteFile(srcPath, []byte("symlink content"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	op := NewOperator(repoDir)
	entries := []domain.ConfigEntry{
		{Path: "link.txt", Mode: domain.ModeSymlink, Type: domain.EntryTypeFile},
	}
	if err := op.Deploy(entries, worktreeDir); err != nil {
		t.Fatalf("deploy: %v", err)
	}

	dstPath := filepath.Join(worktreeDir, "link.txt")
	info, err := os.Lstat(dstPath)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected symlink, got regular file")
	}

	target, err := os.Readlink(dstPath)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if target != srcPath {
		t.Errorf("symlink target: got %s, want %s", target, srcPath)
	}
}
