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

	sessionName := sanitizeSessionName(wt.Branch)
	if sessionName == "" {
		sessionName = sanitizeSessionName(filepath.Base(path))
	}
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

func printTmuxFailure(message string) {
	fmt.Fprintf(os.Stderr, "tmux の起動に失敗しました: %s\n", message)
}
