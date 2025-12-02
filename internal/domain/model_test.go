package domain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigEntryValidate(t *testing.T) {
	tests := []struct {
		name    string
		e       ConfigEntry
		wantErr bool
	}{
		{"ok copy", ConfigEntry{Path: "file.txt", Mode: ModeCopy, Type: EntryTypeFile}, false},
		{"ok symlink dir", ConfigEntry{Path: "dir", Mode: ModeSymlink, Type: EntryTypeDir}, false},
		{"empty path", ConfigEntry{Path: "", Mode: ModeCopy, Type: EntryTypeFile}, true},
		{"abs path", ConfigEntry{Path: "/abs", Mode: ModeCopy, Type: EntryTypeFile}, true},
		{"bad mode", ConfigEntry{Path: "x", Mode: Mode("bad"), Type: EntryTypeFile}, true},
		{"bad type", ConfigEntry{Path: "x", Mode: ModeCopy, Type: EntryType("bad")}, true},
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
	if err := os.WriteFile(filepath.Join(tmp, "a.txt"), []byte(""), 0o644); err != nil {
		t.Fatalf("prepare file: %v", err)
	}
	repo := &inMemoryRepo{}
	svc := NewConfigService(repo, tmp)

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
