package tui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/example/gwm/internal/domain"
)

// SelectWorktree shows a Bubble Tea list UI and returns the chosen worktree.
func SelectWorktree(wts []domain.WorktreeInfo) (domain.WorktreeInfo, error) {
	if len(wts) == 0 {
		return domain.WorktreeInfo{}, fmt.Errorf("no worktrees found")
	}
	items := make([]list.Item, len(wts))
	for i, wt := range wts {
		items[i] = worktreeItem{info: wt}
	}

	m := &model{width: 80}
	delegate := worktreeDelegate{
		selectedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true),
		normalStyle:   lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "242"}),
		width:         &m.width,
	}

	l := list.New(items, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowFilter(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(true)
	l.DisableQuitKeybindings() // handle quit keys ourselves

	m.list = l
	p := tea.NewProgram(m)
	res, err := p.Run()
	if err != nil {
		return domain.WorktreeInfo{}, err
	}
	final := res.(*model)
	if final.cancelled || final.selected == nil {
		return domain.WorktreeInfo{}, fmt.Errorf("selection cancelled")
	}
	return *final.selected, nil
}

type model struct {
	list      list.Model
	selected  *domain.WorktreeInfo
	cancelled bool
	width     int
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(worktreeItem); ok {
				m.selected = &item.info
				return m, tea.Quit
			}
		default:
			if idx, err := parseDigit(msg.String()); err == nil {
				if idx >= 0 && idx < len(m.list.Items()) {
					m.list.Select(idx)
				}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		h := len(m.list.Items()) + 2 // items + pagination
		if h < 3 {
			h = 3
		}
		if h > msg.Height {
			h = msg.Height
		}
		m.list.SetSize(msg.Width, h)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	return m.list.View()
}

type worktreeItem struct {
	info domain.WorktreeInfo
}

type worktreeDelegate struct {
	selectedStyle lipgloss.Style
	normalStyle   lipgloss.Style
	width         *int
}

func (d worktreeDelegate) Height() int                             { return 1 }
func (d worktreeDelegate) Spacing() int                            { return 0 }
func (d worktreeDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d worktreeDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	wt, ok := item.(worktreeItem)
	if !ok {
		return
	}

	title := wt.Title()
	maxWidth := *d.width - 4
	if maxWidth > 0 && len(title) > maxWidth {
		title = truncateTitle(maxWidth, wt.info.Branch, wt.info.Path, wt.info.IsCurrent)
	}

	if index == m.Index() {
		title = d.selectedStyle.Render(title)
	} else {
		title = d.normalStyle.Render(title)
	}

	fmt.Fprint(w, title)
}

func truncateTitle(maxWidth int, branch, absPath string, isCurrent bool) string {
	current := ""
	if isCurrent {
		current = " *"
	}
	relPath := relativePath(absPath)

	minRequired := len(branch) + len(" ()") + len(current) + 3
	if maxWidth < minRequired {
		return truncateString(branch, maxWidth-len(current)) + current
	}

	pathWidth := maxWidth - len(branch) - len(" ()") - len(current)
	truncatedPath := truncateString(relPath, pathWidth)
	return fmt.Sprintf("%s (%s)%s", branch, truncatedPath, current)
}

func (i worktreeItem) Title() string {
	branch := i.info.Branch
	relPath := relativePath(i.info.Path)
	current := ""
	if i.info.IsCurrent {
		current = " *"
	}
	return fmt.Sprintf("%s (%s)%s", branch, relPath, current)
}

func (i worktreeItem) Description() string { return "" }
func (i worktreeItem) FilterValue() string { return i.info.Path }

func parseDigit(s string) (int, error) {
	s = strings.TrimSpace(s)
	if len(s) != 1 {
		return 0, fmt.Errorf("not single digit")
	}
	return strconv.Atoi(s)
}

func relativePath(absPath string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return absPath
	}
	rel, err := filepath.Rel(cwd, absPath)
	if err != nil {
		return absPath
	}
	return rel
}

func truncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
