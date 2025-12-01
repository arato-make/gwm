package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/example/gwm/internal/app/usecase"
	"github.com/example/gwm/internal/domain"
)

type App struct {
	Create *usecase.CreateInteractor
	Config *usecase.ConfigInteractor
	Cd     *usecase.CdInteractor
	Select func([]domain.WorktreeInfo) (string, error)
}

func (a *App) Run(args []string) int {
	if len(args) < 1 {
		fmt.Println("usage: gwm <command>")
		return 1
	}
	switch args[0] {
	case "create":
		return a.runCreate(args[1:])
	case "config":
		return a.runConfig(args[1:])
	case "cd":
		return a.runCd(args[1:])
	default:
		fmt.Println("unknown command:", args[0])
		return 1
	}
}

func (a *App) runCreate(args []string) int {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() < 1 {
		fmt.Println("usage: gwm create <branch>")
		return 1
	}
	branch := fs.Arg(0)
	out, err := a.Create.Execute(usecase.CreateInput{Branch: branch})
	if err != nil {
		fmt.Println("error:", err)
		return 1
	}
	for _, m := range out.Messages {
		fmt.Println(m)
	}
	fmt.Println("worktree:", out.Worktree)
	return 0
}

func (a *App) runConfig(args []string) int {
	if len(args) == 0 {
		fmt.Println("usage: gwm config <add|list|remove> ...")
		return 1
	}
	switch args[0] {
	case "add":
		return a.runConfigAdd(args[1:])
	case "list":
		return a.runConfigList(args[1:])
	case "remove":
		return a.runConfigRemove(args[1:])
	default:
		fmt.Println("unknown config command:", args[0])
		return 1
	}
}

func (a *App) runConfigAdd(args []string) int {
	fs := flag.NewFlagSet("config add", flag.ContinueOnError)
	mode := fs.String("mode", "copy", "copy|symlink")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() < 1 {
		fmt.Println("usage: gwm config add <path> --mode copy|symlink")
		return 1
	}
	entry := domain.ConfigEntry{Path: fs.Arg(0), Mode: domain.Mode(*mode)}
	if err := a.Config.Add(entry); err != nil {
		fmt.Println("error:", err)
		return 1
	}
	fmt.Println("added:", entry.Path, "(", entry.Mode, ")")
	return 0
}

func (a *App) runConfigList(args []string) int {
	if len(args) > 0 {
		fmt.Println("usage: gwm config list")
		return 1
	}
	entries, err := a.Config.List()
	if err != nil {
		fmt.Println("error:", err)
		return 1
	}
	if len(entries) == 0 {
		fmt.Println("no entries")
		return 0
	}
	data, _ := json.MarshalIndent(entries, "", "  ")
	fmt.Println(string(data))
	return 0
}

func (a *App) runConfigRemove(args []string) int {
	if len(args) != 1 {
		fmt.Println("usage: gwm config remove <path>")
		return 1
	}
	if err := a.Config.Remove(args[0]); err != nil {
		fmt.Println("error:", err)
		return 1
	}
	fmt.Println("removed:", args[0])
	return 0
}

func (a *App) runCd(args []string) int {
	if len(args) != 0 {
		fmt.Println("usage: gwm cd")
		return 1
	}
	list, err := a.Cd.List()
	if err != nil {
		fmt.Println("error:", err)
		return 1
	}
	if a.Select == nil {
		return respondForCd(list)
	}
	path, err := a.Select(list)
	if err != nil {
		fmt.Println("error:", err)
		return 1
	}
	fmt.Println("cd", path)
	if err := os.Chdir(path); err != nil {
		fmt.Println("error:", err)
		return 1
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := exec.Command(shell)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("error:", err)
		return 1
	}
	return 0
}

// respondForCd prints JSON to stdout so wrapper can use it; if empty, error.
func respondForCd(list []domain.WorktreeInfo) int {
	if len(list) == 0 {
		fmt.Println("error: no worktrees")
		return 1
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
		return 1
	}
	fmt.Println(string(data))
	return 0
}

var ErrCancel = errors.New("cancelled")
