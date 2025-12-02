package main

import (
	"fmt"
	"os"

	"github.com/example/gwm/internal/app/usecase"
	"github.com/example/gwm/internal/domain"
	"github.com/example/gwm/internal/infra/config"
	"github.com/example/gwm/internal/infra/fs"
	"github.com/example/gwm/internal/infra/git"
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

	cfgRepo := config.NewStore(repoDir)
	configSvc := domain.NewConfigService(cfgRepo)
	wtClient := git.NewWorktreeClient(repoDir)
	fileOps := fs.NewOperator(repoDir)
	sessionLauncher := tmuxinfra.NewLauncher()

	app := cli.App{
		Create: &usecase.CreateInteractor{
			Worktrees: wtClient,
			Config:    cfgRepo,
			FileOps:   fileOps,
			Launcher:  sessionLauncher,
		},
		Config: &usecase.ConfigInteractor{Service: configSvc},
		Cd:     &usecase.CdInteractor{Worktrees: wtClient, Launcher: sessionLauncher},
		Select: tui.SelectWorktree,
	}

	code := app.Run(os.Args[1:])
	os.Exit(code)
}
