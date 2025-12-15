package main

import (
	"fmt"
	"os"

	"github.com/example/gwm/internal/app/usecase"
	"github.com/example/gwm/internal/domain"
	"github.com/example/gwm/internal/infra/config"
	"github.com/example/gwm/internal/infra/fs"
	"github.com/example/gwm/internal/infra/git"
	"github.com/example/gwm/internal/infra/setting"
	"github.com/example/gwm/internal/infra/svc"
	tmuxinfra "github.com/example/gwm/internal/infra/tmux"
	"github.com/example/gwm/internal/interface/cli"
	"github.com/example/gwm/internal/interface/tui"
)

func main() {
	repoDir, err := os.Getwd()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	mainRepoDir, err := git.DetectMainWorktreeDir(repoDir)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	cfgRepo := config.NewStore(repoDir)
	settings, err := setting.Load(repoDir)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	entryTypes := fs.NewEntryTypeResolver(repoDir)
	configSvc := domain.NewConfigService(cfgRepo, entryTypes)
	wtClient := git.NewWorktreeClient(repoDir, mainRepoDir)
	fileOps := fs.NewOperator(repoDir)
	sessionLauncher := tmuxinfra.NewLauncher(settings)

	// Service management infrastructure
	svcDefStore := svc.NewDefinitionStore(mainRepoDir)
	svcRunner := svc.NewRunner(settings)
	svcAllocator := svc.NewPortAllocator(svcRunner)
	svcManager := domain.NewServiceManager(svcDefStore)

	app := cli.App{
		Create: &usecase.CreateInteractor{
			Worktrees: wtClient,
			Config:    configSvc,
			FileOps:   fileOps,
			Launcher:  sessionLauncher,
		},
		Config: &usecase.ConfigInteractor{Service: configSvc},
		Cd:     &usecase.CdInteractor{Worktrees: wtClient, Launcher: sessionLauncher},
		Remove: &usecase.RemoveInteractor{
			Worktrees:     wtClient,
			Launcher:      sessionLauncher,
			ServiceRunner: svcRunner,
		},
		Select: tui.SelectWorktree,

		// Service management use cases
		ServiceAdd:            &usecase.ServiceAddInteractor{Manager: svcManager},
		ServiceStart:          &usecase.ServiceStartInteractor{Manager: svcManager, Runner: svcRunner, Allocator: svcAllocator},
		ServiceStop:           &usecase.ServiceStopInteractor{Runner: svcRunner},
		ServiceList:           &usecase.ServiceListInteractor{Runner: svcRunner},
		ServiceAttach:         &usecase.ServiceAttachInteractor{Runner: svcRunner},
		ServiceRemove:         &usecase.ServiceRemoveInteractor{Manager: svcManager},
		ServiceDefinitionList: &usecase.ServiceDefinitionListInteractor{Manager: svcManager},
	}

	code := app.Run(os.Args[1:])
	os.Exit(code)
}
