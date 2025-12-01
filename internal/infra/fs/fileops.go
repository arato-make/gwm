package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/example/gwm/internal/domain"
)

type Operator struct {
	repoDir string
}

func NewOperator(repoDir string) *Operator {
	return &Operator{repoDir: repoDir}
}

// Deploy copies or symlinks files defined in entries into worktreePath.
func (o *Operator) Deploy(entries []domain.ConfigEntry, worktreePath string) error {
	for _, e := range entries {
		if err := e.Validate(); err != nil {
			return err
		}
		src := filepath.Join(o.repoDir, e.Path)
		dst := filepath.Join(worktreePath, e.Path)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		switch e.Mode {
		case domain.ModeCopy:
			if err := copyFile(src, dst); err != nil {
				return err
			}
		case domain.ModeSymlink:
			if err := os.RemoveAll(dst); err != nil {
				return err
			}
			if err := os.Symlink(src, dst); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown mode: %s", e.Mode)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
