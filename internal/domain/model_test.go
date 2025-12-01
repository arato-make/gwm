package domain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigEntryValidate(t *testing.T) {
	tests := []struct {
		name string
		e    ConfigEntry
		wantErr bool
	}{
		{"ok copy", ConfigEntry{Path: "file.txt", Mode: ModeCopy}, false},
		{"ok symlink", ConfigEntry{Path: "dir/file", Mode: ModeSymlink}, false},
		{"empty path", ConfigEntry{Path: "", Mode: ModeCopy}, true},
		{"abs path", ConfigEntry{Path: "/abs", Mode: ModeCopy}, true},
		{"bad mode", ConfigEntry{Path: "x", Mode: Mode("bad")}, true},
	}
	for _, tt := range tests {
		err := tt.e.Validate()
		if tt.wantErr && err == nil {
			t.Fatalf("%s: expected error", tt.name)
		}
		if !tt.wantErr && err != nil {
			t.Fatalf("%s: unexpected error %v", tt.name, err)
		}
	}
}

func TestConfigServiceAddAndRemove(t *testing.T) {
	tmp := t.TempDir()
	repo := &inMemoryRepo{}
	svc := NewConfigService(repo)

	if err := svc.Add(ConfigEntry{Path: "a.txt", Mode: ModeCopy}); err != nil {
		t.Fatalf("add err: %v", err)
	}
	if err := svc.Add(ConfigEntry{Path: "a.txt", Mode: ModeCopy}); err == nil {
		t.Fatalf("expected duplicate error")
	}
	if err := svc.Remove("missing"); err == nil {
		t.Fatalf("expected missing error")
	}
	if err := svc.Remove("a.txt"); err != nil {
		t.Fatalf("remove err: %v", err)
	}
	if len(repo.data) != 0 {
		t.Fatalf("expected empty repo after remove")
	}

	_ = os.WriteFile(filepath.Join(tmp, "dummy"), []byte(""), 0o644)
}

type inMemoryRepo struct {
	data []ConfigEntry
}

func (r *inMemoryRepo) Load() ([]ConfigEntry, error) {
	return append([]ConfigEntry{}, r.data...), nil
}

func (r *inMemoryRepo) Save(entries []ConfigEntry) error {
	r.data = append([]ConfigEntry{}, entries...)
	return nil
}
