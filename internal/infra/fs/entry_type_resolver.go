package fs

import (
	"os"
	"path/filepath"

	"github.com/example/gwm/internal/domain"
)

// EntryTypeResolver resolves a config entry's type (file/dir) from repo-relative path.
type EntryTypeResolver struct {
	repoDir string
}

func NewEntryTypeResolver(repoDir string) *EntryTypeResolver {
	return &EntryTypeResolver{repoDir: repoDir}
}

func (r *EntryTypeResolver) ResolveEntryType(relPath string) (domain.EntryType, error) {
	info, err := os.Stat(filepath.Join(r.repoDir, relPath))
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return domain.EntryTypeDir, nil
	}
	return domain.EntryTypeFile, nil
}
