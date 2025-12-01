package tui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/example/gwm/internal/domain"
)

// SelectWorktree shows a simple numeric menu and returns chosen path.
func SelectWorktree(list []domain.WorktreeInfo) (string, error) {
	if len(list) == 0 {
		return "", fmt.Errorf("no worktrees found")
	}
	fmt.Println("Select worktree:")
	for i, wt := range list {
		current := ""
		if wt.IsCurrent {
			current = " *"
		}
		fmt.Printf(" [%d] %s (%s)%s\n", i, wt.Path, wt.Branch, current)
	}
	fmt.Print("Enter number: ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return "", fmt.Errorf("selection cancelled")
	}
	idx, err := strconv.Atoi(line)
	if err != nil || idx < 0 || idx >= len(list) {
		return "", fmt.Errorf("invalid selection")
	}
	return list[idx].Path, nil
}
