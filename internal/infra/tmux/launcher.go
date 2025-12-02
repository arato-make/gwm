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
		printTmuxFailure("tmux が見つかりません")
		return launchShell(path)
	}

	sessionName := sanitizeSessionName(wt.Branch)
	if sessionName == "" {
		sessionName = sanitizeSessionName(filepath.Base(path))
	}
	if sessionName == "" {
		sessionName = "gwm-session"
	}

	has, err := l.server.HasSession(sessionName)
	if err != nil {
		// tmux 実行失敗時もシェルにフォールバック
		printTmuxFailure(fmt.Sprintf("セッション確認に失敗しました: %v", err))
		return launchShell(path)
	}

	var session gotmux.Session
	if has {
		session = gotmux.Session{Name: sessionName}
	} else {
		session, err = l.server.NewSession(sessionName, "-c", path)
		if err != nil {
			printTmuxFailure(fmt.Sprintf("セッション作成に失敗しました: %v", err))
			return launchShell(path)
		}
	}

	if err := session.AttachSession(); err != nil {
		printTmuxFailure(fmt.Sprintf("セッションへの接続に失敗しました: %v", err))
		return launchShell(path)
	}
	return nil
}

func isTmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// launchShell starts the user's shell from the given directory.
func launchShell(path string) error {
	if err := os.Chdir(path); err != nil {
		return err
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := exec.Command(shell)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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

func printTmuxFailure(message string) {
	fmt.Fprintf(os.Stderr, "tmux の起動に失敗しました: %s\n", message)
}
