package domain

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// Mode indicates how a file should be deployed into a worktree.
type Mode string

const (
	ModeCopy    Mode = "copy"
	ModeSymlink Mode = "symlink"
)

// EntryType describes whether a config target is a file or directory.
type EntryType string

const (
	EntryTypeFile EntryType = "file"
	EntryTypeDir  EntryType = "dir"
)

// ConfigEntry represents a file managed by gwm.
type ConfigEntry struct {
	Path string    `json:"path"`
	Mode Mode      `json:"mode"`
	Type EntryType `json:"type,omitempty"`
}

// Validate checks the integrity of ConfigEntry.
func (c ConfigEntry) Validate() error {
	if strings.TrimSpace(c.Path) == "" {
		return errors.New("path is required")
	}
	if filepath.IsAbs(c.Path) {
		return errors.New("path must be relative")
	}
	switch c.Mode {
	case ModeCopy, ModeSymlink:
	default:
		return fmt.Errorf("unsupported mode: %s", c.Mode)
	}
	switch c.Type {
	case EntryTypeFile, EntryTypeDir:
	default:
		return fmt.Errorf("unsupported type: %s", c.Type)
	}
	return nil
}

// WorktreeInfo describes a git worktree.
type WorktreeInfo struct {
	Branch    string `json:"branch"`
	Path      string `json:"path"`
	IsCurrent bool   `json:"isCurrent"`
}

// CommandResult holds user-facing messages and errors.
type CommandResult struct {
	Messages []string
	Err      error
}
