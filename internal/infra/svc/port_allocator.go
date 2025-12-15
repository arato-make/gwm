package svc

import (
	"fmt"
	"net"

	"github.com/example/gwm/internal/domain"
)

const (
	PortRangeStart = 3000
	PortRangeEnd   = 3999
)

// PortAllocator implements domain.PortAllocator.
type PortAllocator struct {
	runner domain.ServiceRunner
}

// NewPortAllocator creates a PortAllocator.
func NewPortAllocator(runner domain.ServiceRunner) *PortAllocator {
	return &PortAllocator{runner: runner}
}

// Allocate returns an available port.
func (a *PortAllocator) Allocate(mode domain.PortMode, fixedPort int) (int, error) {
	if mode == domain.PortModeNone || mode == "" {
		return 0, nil
	}

	usedPorts := a.getUsedPorts()

	if mode == domain.PortModeFixed {
		if fixedPort <= 0 {
			return 0, fmt.Errorf("fixed port must be positive")
		}
		return fixedPort, nil
	}

	// Auto mode
	for port := PortRangeStart; port <= PortRangeEnd; port++ {
		if usedPorts[port] {
			continue
		}
		if isPortAvailable(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports in range %d-%d", PortRangeStart, PortRangeEnd)
}

// GetServiceUsingPort returns the service using a specific port, or nil.
func (a *PortAllocator) GetServiceUsingPort(port int) *domain.RunningService {
	services, err := a.runner.List()
	if err != nil {
		return nil
	}
	for _, s := range services {
		if s.Port == port {
			return &s
		}
	}
	return nil
}

func (a *PortAllocator) getUsedPorts() map[int]bool {
	used := make(map[int]bool)
	services, err := a.runner.List()
	if err != nil {
		return used
	}
	for _, s := range services {
		if s.Port > 0 {
			used[s.Port] = true
		}
	}
	return used
}

func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}
