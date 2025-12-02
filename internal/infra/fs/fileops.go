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
		if e.Type == "" {
			typ, err := detectEntryType(o.repoDir, e.Path)
			if err != nil {
				return err
			}
			e.Type = typ
		}
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
			if e.Type == domain.EntryTypeDir {
				if err := copyDir(src, dst); err != nil {
					return err
				}
			} else {
				if err := copyFile(src, dst); err != nil {
					return err
				}
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

	inStat, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if err := out.Chmod(inStat.Mode()); err != nil {
		return err
	}
	return out.Close()
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		s := filepath.Join(src, entry.Name())
		d := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(s, d); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(s, d); err != nil {
			return err
		}
	}
	return nil
}

func detectEntryType(repoDir, relPath string) (domain.EntryType, error) {
	info, err := os.Stat(filepath.Join(repoDir, relPath))
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return domain.EntryTypeDir, nil
	}
	return domain.EntryTypeFile, nil
}
