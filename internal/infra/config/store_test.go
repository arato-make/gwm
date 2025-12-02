package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/gwm/internal/domain"
)

func TestStoreLoadSave(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	// initial load is empty
	entries, err := s.Load()
	if err != nil {
		t.Fatalf("load err: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty, got %d", len(entries))
	}

	want := []domain.ConfigEntry{
		{Path: "a.txt", Mode: domain.ModeCopy, Type: domain.EntryTypeFile},
		{Path: "b.txt", Mode: domain.ModeSymlink, Type: domain.EntryTypeFile},
	}
	if err := s.Save(want); err != nil {
		t.Fatalf("save err: %v", err)
	}

	got, err := s.Load()
	if err != nil {
		t.Fatalf("load err: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("len mismatch")
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("entry mismatch: %v vs %v", got[i], want[i])
		}
	}

	// corrupted json should error
	path := filepath.Join(dir, ".gwm", "config.json")
	if err := os.WriteFile(path, []byte("{bad json"), 0o644); err != nil {
		t.Fatalf("write err: %v", err)
	}
	if _, err := s.Load(); err == nil {
		t.Fatalf("expected error for bad json")
	}

	// type should be auto-detected when missing
	if err := os.WriteFile(filepath.Join(dir, "c.txt"), []byte(""), 0o644); err != nil {
		t.Fatalf("prepare file: %v", err)
	}
	omitType := []domain.ConfigEntry{{Path: "c.txt", Mode: domain.ModeCopy}}
	if err := s.Save(omitType); err != nil {
		t.Fatalf("save err: %v", err)
	}
	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("load err: %v", err)
	}
	if len(loaded) != 1 || loaded[0].Type != domain.EntryTypeFile {
		t.Fatalf("type not inferred: %+v", loaded)
	}
}
