package usecase

import (
	"fmt"
	"path/filepath"

	"github.com/example/gwm/internal/domain"
)

// ServiceAddInput is input for ServiceAddInteractor.
type ServiceAddInput struct {
	Name      string
	Command   string
	PortMode  domain.PortMode
	FixedPort int
}

// ServiceAddInteractor adds a service definition.
type ServiceAddInteractor struct {
	Manager *domain.ServiceManager
}

func (u *ServiceAddInteractor) Execute(in ServiceAddInput) error {
	def := domain.ServiceDefinition{
		Name:      in.Name,
		Command:   in.Command,
		Port:      in.PortMode,
		FixedPort: in.FixedPort,
	}
	return u.Manager.Add(def)
}

// ServiceStartInput is input for ServiceStartInteractor.
type ServiceStartInput struct {
	Name         string
	WorktreePath string
}

// ServiceStartOutput is output for ServiceStartInteractor.
type ServiceStartOutput struct {
	SessionName string
	Port        int
	Messages    []string
}

// ServiceStartInteractor starts a service.
type ServiceStartInteractor struct {
	Manager   *domain.ServiceManager
	Runner    domain.ServiceRunner
	Allocator domain.PortAllocator
}

func (u *ServiceStartInteractor) Execute(in ServiceStartInput) (ServiceStartOutput, error) {
	var out ServiceStartOutput

	def, err := u.Manager.Get(in.Name)
	if err != nil {
		return out, err
	}
	if def == nil {
		return out, fmt.Errorf("service %s not found", in.Name)
	}

	port, err := u.Allocator.Allocate(def.Port, def.FixedPort)
	if err != nil {
		return out, err
	}

	// Handle port conflict for fixed port
	if def.Port == domain.PortModeFixed && port > 0 {
		existing := u.Allocator.GetServiceUsingPort(port)
		if existing != nil {
			if err := u.Runner.Stop(existing.SessionName); err != nil {
				return out, fmt.Errorf("failed to stop conflicting service: %w", err)
			}
			out.Messages = append(out.Messages,
				fmt.Sprintf("stopped conflicting service %s on port %d", existing.Name, port))
		}
	}

	sessionName, err := u.Runner.Start(in.WorktreePath, *def, port)
	if err != nil {
		return out, err
	}

	out.SessionName = sessionName
	out.Port = port
	out.Messages = append(out.Messages, fmt.Sprintf("started service %s", in.Name))
	if port > 0 {
		out.Messages = append(out.Messages, fmt.Sprintf("assigned port %d", port))
	}

	return out, nil
}

// ServiceStopInput is input for ServiceStopInteractor.
type ServiceStopInput struct {
	Name         string
	WorktreePath string
}

// ServiceStopOutput is output for ServiceStopInteractor.
type ServiceStopOutput struct {
	Messages []string
}

// ServiceStopInteractor stops a service.
type ServiceStopInteractor struct {
	Runner domain.ServiceRunner
}

func (u *ServiceStopInteractor) Execute(in ServiceStopInput) (ServiceStopOutput, error) {
	var out ServiceStopOutput

	services, err := u.Runner.List()
	if err != nil {
		return out, err
	}

	// Compare worktree by basename since session name only contains basename
	inputWorktreeBase := filepath.Base(in.WorktreePath)

	var toStop []domain.RunningService
	for _, s := range services {
		if in.Name != "" && s.Name != in.Name {
			continue
		}
		if in.WorktreePath != "" && s.WorktreePath != inputWorktreeBase {
			continue
		}
		toStop = append(toStop, s)
	}

	if len(toStop) == 0 {
		return out, fmt.Errorf("no matching services found")
	}

	for _, s := range toStop {
		if err := u.Runner.Stop(s.SessionName); err != nil {
			return out, fmt.Errorf("failed to stop %s: %w", s.Name, err)
		}
		out.Messages = append(out.Messages, fmt.Sprintf("stopped %s", s.Name))
	}

	return out, nil
}

// ServiceListOutput is output for ServiceListInteractor.
type ServiceListOutput struct {
	Services []domain.RunningService
}

// ServiceListInteractor lists running services.
type ServiceListInteractor struct {
	Runner domain.ServiceRunner
}

func (u *ServiceListInteractor) Execute() (ServiceListOutput, error) {
	services, err := u.Runner.List()
	if err != nil {
		return ServiceListOutput{}, err
	}
	return ServiceListOutput{Services: services}, nil
}

// ServiceAttachInput is input for ServiceAttachInteractor.
type ServiceAttachInput struct {
	Name         string
	WorktreePath string
}

// ServiceAttachInteractor attaches to a service session.
type ServiceAttachInteractor struct {
	Runner domain.ServiceRunner
}

func (u *ServiceAttachInteractor) Execute(in ServiceAttachInput) error {
	services, err := u.Runner.List()
	if err != nil {
		return err
	}

	inputWorktreeBase := filepath.Base(in.WorktreePath)

	for _, s := range services {
		if s.Name == in.Name {
			if in.WorktreePath == "" || s.WorktreePath == inputWorktreeBase {
				return u.Runner.Attach(s.SessionName)
			}
		}
	}

	return fmt.Errorf("service %s not found", in.Name)
}

// ServiceRemoveInput is input for ServiceRemoveInteractor.
type ServiceRemoveInput struct {
	Name string
}

// ServiceRemoveInteractor removes a service definition.
type ServiceRemoveInteractor struct {
	Manager *domain.ServiceManager
}

func (u *ServiceRemoveInteractor) Execute(in ServiceRemoveInput) error {
	return u.Manager.Remove(in.Name)
}

// ServiceDefinitionListOutput is output for ServiceDefinitionListInteractor.
type ServiceDefinitionListOutput struct {
	Definitions []domain.ServiceDefinition
}

// ServiceDefinitionListInteractor lists service definitions.
type ServiceDefinitionListInteractor struct {
	Manager *domain.ServiceManager
}

func (u *ServiceDefinitionListInteractor) Execute() (ServiceDefinitionListOutput, error) {
	defs, err := u.Manager.List()
	if err != nil {
		return ServiceDefinitionListOutput{}, err
	}
	return ServiceDefinitionListOutput{Definitions: defs}, nil
}

// StopServicesForWorktree stops all services for a worktree.
func StopServicesForWorktree(runner domain.ServiceRunner, worktreePath string) error {
	services, err := runner.List()
	if err != nil {
		return err
	}

	worktreeBase := filepath.Base(worktreePath)

	for _, s := range services {
		if s.WorktreePath == worktreeBase {
			runner.Stop(s.SessionName)
		}
	}
	return nil
}
