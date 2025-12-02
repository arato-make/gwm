package tmux

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gotmux "github.com/jubnzv/go-tmux"

	"github.com/example/gwm/internal/domain"
)

// Launcher implements domain.SessionLauncher using tmux with a shell fallback.
type Launcher struct {
	server *gotmux.Server
}

func NewLauncher() *Launcher {
	return &Launcher{server: gotmux.NewServer("", "", nil)}
}

func (l *Launcher) Launch(wt domain.WorktreeInfo) error {
	if strings.TrimSpace(wt.Path) == "" {
		return errors.New("worktree path is empty")
	}
	path := wt.Path
	if !filepath.IsAbs(path) {
		var err error
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
	}

	// tmux が無い場合は従来のシェル起動にフォールバック
	if !isTmuxAvailable() {
		return errors.New("tmux が見つかりません")
	}

	sessionName := firstNonEmpty(sessionNameCandidates(wt))
	if sessionName == "" {
		sessionName = "gwm-session"
	}

	has, err := l.server.HasSession(sessionName)
	if err != nil {
		// サーバーがまだ立ち上がっていない場合でも新規作成を試みる
		printTmuxFailure(fmt.Sprintf("セッション確認に失敗しました (新規作成を試みます): %v", err))
	}

	var session gotmux.Session
	if err == nil && has {
		session = gotmux.Session{Name: sessionName}
	} else {
		session, err = l.server.NewSession(sessionName, "-c", path)
		if err != nil {
			printTmuxFailure(fmt.Sprintf("セッション作成に失敗しました: %v", err))
			return err
		}
	}

	if err := session.AttachSession(); err != nil {
		printTmuxFailure(fmt.Sprintf("セッションへの接続に失敗しました: %v", err))
		return err
	}
	return nil
}

// Kill terminates a tmux session related to the worktree if it exists.
func (l *Launcher) Kill(wt domain.WorktreeInfo) error {
	if !isTmuxAvailable() {
		return nil
	}

	for _, name := range sessionNameCandidates(wt) {
		if name == "" {
			continue
		}
		has, err := l.server.HasSession(name)
		if err != nil {
			printTmuxFailure(fmt.Sprintf("セッション確認に失敗しました: %v", err))
			continue
		}
		if !has {
			continue
		}
		if err := l.server.KillSession(name); err != nil {
			printTmuxFailure(fmt.Sprintf("セッション削除に失敗しました: %v", err))
			return err
		}
	}
	return nil
}

func isTmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func sanitizeSessionName(name string) string {
	s := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') {
			return r
		}
		if r == '-' || r == '_' {
			return r
		}
		if r == '/' || r == '\\' {
			return '-'
		}
		return -1
	}, name)
	return strings.Trim(s, "-_")
}

func sessionNameCandidates(wt domain.WorktreeInfo) []string {
	branch := sanitizeSessionName(wt.Branch)
	trimmedBranch := sanitizeSessionName(strings.TrimPrefix(wt.Branch, "refs/heads/"))
	base := sanitizeSessionName(filepath.Base(wt.Path))

	primary := firstNonEmpty([]string{branch, trimmedBranch, base})
	var candidates []string
	if primary != "" {
		candidates = append(candidates, "gwm-"+primary)
	}

	// Legacy names (without prefix) are kept for cleanup/compatibility.
	for _, n := range []string{branch, trimmedBranch, base} {
		if n == "" {
			continue
		}
		if contains(candidates, n) || contains(candidates, "gwm-"+n) {
			continue
		}
		candidates = append(candidates, n)
	}
	return candidates
}

func contains(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func firstNonEmpty(list []string) string {
	for _, v := range list {
		if v != "" {
			return v
		}
	}
	return ""
}

func printTmuxFailure(message string) {
	fmt.Fprintf(os.Stderr, "tmux の起動に失敗しました: %s\n", message)
}
