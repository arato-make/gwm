package tui

import (
	"fmt"
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

	styles := list.NewDefaultDelegate()
	styles.ShowDescription = false
	styles.Styles.SelectedTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	styles.Styles.NormalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	l := list.New(items, styles, 0, 0)
	l.Title = "Select worktree (Enter to attach, q/Esc to cancel, digits to jump)"
	l.SetShowHelp(false)
	l.SetShowFilter(false)
	l.SetShowStatusBar(false)
	l.DisableQuitKeybindings() // handle quit keys ourselves

	m := model{list: l}
	p := tea.NewProgram(m)
	res, err := p.Run()
	if err != nil {
		return domain.WorktreeInfo{}, err
	}
	final := res.(model)
	if final.cancelled || final.selected == nil {
		return domain.WorktreeInfo{}, fmt.Errorf("selection cancelled")
	}
	return *final.selected, nil
}

type model struct {
	list      list.Model
	selected  *domain.WorktreeInfo
	cancelled bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// Bubble Teaが端末サイズを送ってきたときにリストの表示幅・高さを更新する。
		// 幅が0のままだとデリゲートが何も描画しないため、文字が見えなくなる。
		m.list.SetSize(msg.Width, msg.Height)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.list.View()
}

type worktreeItem struct {
	info domain.WorktreeInfo
}

func (i worktreeItem) Title() string {
	current := ""
	if i.info.IsCurrent {
		current = " *"
	}
	return fmt.Sprintf("%s (%s)%s", i.info.Path, i.info.Branch, current)
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
