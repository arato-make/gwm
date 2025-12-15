package domain

import (
	"errors"
	"fmt"
	"strings"
)

// PortMode indicates how a port should be assigned.
type PortMode string

const (
	PortModeAuto  PortMode = "auto"
	PortModeFixed PortMode = "fixed"
	PortModeNone  PortMode = "none"
)

// ServiceDefinition represents a pre-registered service.
type ServiceDefinition struct {
	Name      string   `json:"name"`
	Command   string   `json:"command"`
	Port      PortMode `json:"port,omitempty"`
	FixedPort int      `json:"fixedPort,omitempty"`
	Unique    bool     `json:"unique,omitempty"`
}

// Validate checks the integrity of ServiceDefinition.
func (s ServiceDefinition) Validate() error {
	if strings.TrimSpace(s.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(s.Command) == "" {
		return errors.New("command is required")
	}
	switch s.Port {
	case PortModeAuto, PortModeFixed, PortModeNone, "":
	default:
		return fmt.Errorf("unsupported port mode: %s", s.Port)
	}
	if s.Port == PortModeFixed && s.FixedPort <= 0 {
		return errors.New("fixed port must be positive")
	}
	return nil
}

// RunningService represents a currently running service.
type RunningService struct {
	Name         string `json:"name"`
	WorktreePath string `json:"worktreePath"`
	SessionName  string `json:"sessionName"`
	Port         int    `json:"port,omitempty"`
}
