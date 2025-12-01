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
		{Path: "a.txt", Mode: domain.ModeCopy},
		{Path: "b.txt", Mode: domain.ModeSymlink},
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
}
