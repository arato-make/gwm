package svc

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	gotmux "github.com/jubnzv/go-tmux"

	"github.com/example/gwm/internal/domain"
)

const portPlaceholder = "{port}"

const sessionPrefix = "gwm-svc-"

// Runner implements domain.ServiceRunner using tmux.
type Runner struct {
	server *gotmux.Server
}

// NewRunner creates a Runner.
func NewRunner(settings domain.Settings) *Runner {
	return &Runner{
		server: gotmux.NewServer("", "", nil),
	}
}

// GenerateSessionName creates a session name for a service.
// Format: gwm-svc-<worktree>-<name>[-p<port>]
func GenerateSessionName(worktreePath, serviceName string, port int) string {
	base := filepath.Base(worktreePath)
	sanitizedWorktree := sanitizeName(base)
	sanitizedService := sanitizeName(serviceName)
	if port > 0 {
		return fmt.Sprintf("%s%s-%s-p%d", sessionPrefix, sanitizedWorktree, sanitizedService, port)
	}
	return fmt.Sprintf("%s%s-%s", sessionPrefix, sanitizedWorktree, sanitizedService)
}

func sanitizeName(name string) string {
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

// Start creates a new tmux session and runs the command.
func (r *Runner) Start(worktreePath string, def domain.ServiceDefinition, port int) (string, error) {
	if !isTmuxAvailable() {
		return "", fmt.Errorf("tmux is not available")
	}

	sessionName := GenerateSessionName(worktreePath, def.Name, port)

	has, err := r.server.HasSession(sessionName)
	if err == nil && has {
		return "", fmt.Errorf("session %s already exists", sessionName)
	}

	absPath := worktreePath
	if !filepath.IsAbs(worktreePath) {
		absPath, err = filepath.Abs(worktreePath)
		if err != nil {
			return "", err
		}
	}

	_, err = r.server.NewSession(sessionName, "-c", absPath)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	// Replace {port} placeholder with actual port number
	cmd := def.Command
	if port > 0 {
		cmd = strings.ReplaceAll(cmd, portPlaceholder, strconv.Itoa(port))
	}

	// Send command to the session using tmux send-keys
	_, _, err = gotmux.RunCmd([]string{"send-keys", "-t", sessionName, cmd, "Enter"})
	if err != nil {
		r.server.KillSession(sessionName)
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	return sessionName, nil
}

// Stop terminates a service's tmux session.
func (r *Runner) Stop(sessionName string) error {
	if !isTmuxAvailable() {
		return nil
	}

	has, err := r.server.HasSession(sessionName)
	if err != nil {
		return fmt.Errorf("failed to check session: %w", err)
	}
	if !has {
		return nil
	}

	return r.server.KillSession(sessionName)
}

// IsRunning checks if a session exists.
func (r *Runner) IsRunning(sessionName string) (bool, error) {
	if !isTmuxAvailable() {
		return false, nil
	}
	return r.server.HasSession(sessionName)
}

// Attach attaches to a service's tmux session.
func (r *Runner) Attach(sessionName string) error {
	if !isTmuxAvailable() {
		return fmt.Errorf("tmux is not available")
	}

	has, err := r.server.HasSession(sessionName)
	if err != nil {
		return fmt.Errorf("failed to check session: %w", err)
	}
	if !has {
		return fmt.Errorf("session %s does not exist", sessionName)
	}

	session := gotmux.Session{Name: sessionName}
	return session.AttachSession()
}

// List returns all running gwm services by parsing tmux session names.
func (r *Runner) List() ([]domain.RunningService, error) {
	if !isTmuxAvailable() {
		return nil, nil
	}

	sessions, err := r.server.ListSessions()
	if err != nil {
		return nil, nil
	}

	var services []domain.RunningService
	for _, sess := range sessions {
		if !strings.HasPrefix(sess.Name, sessionPrefix) {
			continue
		}
		svc := parseSessionName(sess.Name)
		if svc != nil {
			services = append(services, *svc)
		}
	}
	return services, nil
}

// parseSessionName extracts service info from session name.
// Format: gwm-svc-<worktree>-<name>[-p<port>]
func parseSessionName(name string) *domain.RunningService {
	if !strings.HasPrefix(name, sessionPrefix) {
		return nil
	}

	rest := strings.TrimPrefix(name, sessionPrefix)

	// Try to extract port suffix
	var port int
	portRegex := regexp.MustCompile(`-p(\d+)$`)
	if matches := portRegex.FindStringSubmatch(rest); len(matches) == 2 {
		port, _ = strconv.Atoi(matches[1])
		rest = portRegex.ReplaceAllString(rest, "")
	}

	// Split remaining by '-' to get worktree and service name
	// This is tricky because both can contain '-'
	// We assume the last segment is the service name
	parts := strings.Split(rest, "-")
	if len(parts) < 2 {
		return nil
	}

	serviceName := parts[len(parts)-1]
	worktree := strings.Join(parts[:len(parts)-1], "-")

	return &domain.RunningService{
		Name:         serviceName,
		WorktreePath: worktree,
		SessionName:  name,
		Port:         port,
	}
}

func isTmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}
