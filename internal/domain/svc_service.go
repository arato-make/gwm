package domain

import "errors"

// ServiceDefinitionRepository persists service definitions.
type ServiceDefinitionRepository interface {
	Load() ([]ServiceDefinition, error)
	Save([]ServiceDefinition) error
}

// ServiceRunner manages service execution in tmux sessions.
type ServiceRunner interface {
	Start(worktreePath string, def ServiceDefinition, port int) (sessionName string, err error)
	Stop(sessionName string) error
	IsRunning(sessionName string) (bool, error)
	Attach(sessionName string) error
	List() ([]RunningService, error)
}

// PortAllocator manages port assignment.
type PortAllocator interface {
	Allocate(mode PortMode, fixedPort int) (int, error)
	GetServiceUsingPort(port int) *RunningService
}

// ServiceManager provides add/list/remove operations on service definitions.
type ServiceManager struct {
	repo ServiceDefinitionRepository
}

func NewServiceManager(repo ServiceDefinitionRepository) *ServiceManager {
	return &ServiceManager{repo: repo}
}

func (m *ServiceManager) List() ([]ServiceDefinition, error) {
	return m.repo.Load()
}

func (m *ServiceManager) Add(def ServiceDefinition) error {
	if err := def.Validate(); err != nil {
		return err
	}
	defs, err := m.repo.Load()
	if err != nil {
		return err
	}
	for _, d := range defs {
		if d.Name == def.Name {
			return errors.New("service already exists")
		}
	}
	defs = append(defs, def)
	return m.repo.Save(defs)
}

func (m *ServiceManager) Remove(name string) error {
	defs, err := m.repo.Load()
	if err != nil {
		return err
	}
	kept := defs[:0]
	found := false
	for _, d := range defs {
		if d.Name == name {
			found = true
			continue
		}
		kept = append(kept, d)
	}
	if !found {
		return errors.New("service not found")
	}
	return m.repo.Save(kept)
}

func (m *ServiceManager) Get(name string) (*ServiceDefinition, error) {
	defs, err := m.repo.Load()
	if err != nil {
		return nil, err
	}
	for _, d := range defs {
		if d.Name == name {
			return &d, nil
		}
	}
	return nil, nil
}
