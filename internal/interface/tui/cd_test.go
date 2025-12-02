package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/example/gwm/internal/domain"
)

func TestModelHandlesWindowSize(t *testing.T) {
	items := []list.Item{
		worktreeItem{info: domain.WorktreeInfo{Path: "worktrees/feature", Branch: "feature/foo"}},
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	delegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	l := list.New(items, delegate, 0, 0)
	l.SetShowHelp(false)
	l.SetShowFilter(false)
	l.SetShowStatusBar(false)
	l.DisableQuitKeybindings()

	m := model{list: l}
	mAny, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	m = mAny.(model)

	if m.list.Width() != 40 || m.list.Height() != 10 {
		t.Fatalf("list size not updated, got (%d, %d)", m.list.Width(), m.list.Height())
	}

	if view := m.View(); view == "" {
		t.Fatal("view should not be empty after setting size")
	}
}

func TestParseDigit(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"0", 0, false},
		{"5", 5, false},
		{"a", 0, true},
		{"12", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		got, err := parseDigit(tt.in)
		if tt.wantErr && err == nil {
			t.Fatalf("parseDigit(%q) expected error", tt.in)
		}
		if !tt.wantErr && (err != nil || got != tt.want) {
			t.Fatalf("parseDigit(%q) = (%d, %v), want (%d, nil)", tt.in, got, err, tt.want)
		}
	}
}
