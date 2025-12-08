package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCreateBranchUsesCurrentBranch(t *testing.T) {
	repoDir := t.TempDir()

	runGit(t, repoDir, "-c", "init.defaultBranch=main", "init")
	writeFile(t, filepath.Join(repoDir, "README.md"), "root\n")
	runGit(t, repoDir, "add", "README.md")
	runGit(t, repoDir, "commit", "-m", "init")

	runGit(t, repoDir, "checkout", "-b", "develop")
	writeFile(t, filepath.Join(repoDir, "README.md"), "develop\n")
	runGit(t, repoDir, "add", "README.md")
	runGit(t, repoDir, "commit", "-m", "develop")

	client := NewWorktreeClient(repoDir)
	if err := client.CreateBranch("feature/foo"); err != nil {
		t.Fatalf("create branch: %v", err)
	}

	feature := revParse(t, repoDir, "feature/foo")
	develop := revParse(t, repoDir, "develop")
	if feature != develop {
		t.Fatalf("feature branch is not based on current branch: feature=%s develop=%s", feature, develop)
	}
}

func runGit(t *testing.T, repoDir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repoDir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v (%s)", args, err, out)
	}
}

func revParse(t *testing.T, repoDir, ref string) string {
	t.Helper()
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", ref)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("rev-parse %s failed: %v", ref, err)
	}
	return string(out)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
