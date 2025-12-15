package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/example/gwm/internal/app/usecase"
	"github.com/example/gwm/internal/domain"
)

type App struct {
	Create *usecase.CreateInteractor
	Config *usecase.ConfigInteractor
	Cd     *usecase.CdInteractor
	Remove *usecase.RemoveInteractor
	Select func([]domain.WorktreeInfo) (domain.WorktreeInfo, error)

	// Service management
	ServiceAdd            *usecase.ServiceAddInteractor
	ServiceStart          *usecase.ServiceStartInteractor
	ServiceStop           *usecase.ServiceStopInteractor
	ServiceList           *usecase.ServiceListInteractor
	ServiceAttach         *usecase.ServiceAttachInteractor
	ServiceRemove         *usecase.ServiceRemoveInteractor
	ServiceDefinitionList *usecase.ServiceDefinitionListInteractor
}

func (a *App) Run(args []string) int {
	if len(args) < 1 {
		printRootUsage()
		return 1
	}
	if isHelp(args[0]) {
		printRootUsage()
		return 0
	}
	switch args[0] {
	case "create":
		return a.runCreate(args[1:])
	case "config":
		return a.runConfig(args[1:])
	case "cd":
		return a.runCd(args[1:])
	case "remove":
		return a.runRemove(args[1:])
	case "service":
		return a.runService(args[1:])
	default:
		fmt.Println("unknown command:", args[0])
		printRootUsage()
		return 1
	}
}

func (a *App) runCreate(args []string) int {
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
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
	fs.SetOutput(os.Stdout)
	mode := fs.String("mode", "copy", "copy|symlink")
	if err := fs.Parse(reorderConfigAddArgs(args)); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
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
	wt, err := a.Select(list)
	if err != nil {
		fmt.Println("error:", err)
		return 1
	}
	if err := a.Cd.Launch(wt); err != nil {
		fmt.Println("error:", err)
		return 1
	}
	return 0
}

func (a *App) runRemove(args []string) int {
	fs := flag.NewFlagSet("remove", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	force := fs.Bool("force", false, "force removal even if dirty")
	if err := fs.Parse(reorderRemoveArgs(args)); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}
	if a.Remove == nil {
		fmt.Println("error: remove usecase not configured")
		return 1
	}

	branch := ""
	if fs.NArg() == 1 {
		branch = fs.Arg(0)
	} else if fs.NArg() == 0 {
		list, err := a.Remove.Worktrees.ListWorktrees()
		if err != nil {
			fmt.Println("error:", err)
			return 1
		}
		list = filterRemovableWorktrees(list)
		if len(list) == 0 {
			fmt.Println("error: no worktrees")
			return 1
		}
		if a.Select == nil {
			return respondForCd(list)
		}
		wt, err := a.Select(list)
		if err != nil {
			fmt.Println("error:", err)
			return 1
		}
		branch = wt.Branch
	} else {
		fmt.Println("usage: gwm remove <branch> [--force]")
		return 1
	}

	in := usecase.RemoveInput{Branch: branch, Force: *force}
	out, err := a.Remove.Execute(in)
	if err != nil {
		fmt.Println("error:", err)
		return 1
	}
	for _, m := range out.Messages {
		fmt.Println(m)
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

// reorderConfigAddArgs allows "gwm config add <path> --mode ..." by moving the
// first positional argument to the end so that flag parsing still works.
func reorderConfigAddArgs(args []string) []string {
	return moveFirstPositionalArgToEnd(args)
}

func reorderRemoveArgs(args []string) []string {
	return moveFirstPositionalArgToEnd(args)
}

func moveFirstPositionalArgToEnd(args []string) []string {
	if len(args) == 0 {
		return args
	}
	if !strings.HasPrefix(args[0], "-") {
		return append(args[1:], args[0])
	}
	return args
}

func filterRemovableWorktrees(list []domain.WorktreeInfo) []domain.WorktreeInfo {
	out := make([]domain.WorktreeInfo, 0, len(list))
	for _, wt := range list {
		if isMainWorktreePath(wt.Path) {
			continue
		}
		out = append(out, wt)
	}
	return out
}

// isMainWorktreePath returns true for the primary worktree (repo root) in typical git worktree setups.
// The main worktree usually has ".git" as a directory, while linked worktrees have ".git" as a file.
func isMainWorktreePath(worktreePath string) bool {
	info, err := os.Stat(filepath.Join(worktreePath, ".git"))
	if err != nil {
		return false
	}
	return info.IsDir()
}

func isHelp(arg string) bool {
	return arg == "-h" || arg == "--help"
}

func printRootUsage() {
	fmt.Println("usage: gwm <command>")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  create <branch>              create a worktree and expand config files")
	fmt.Println("  config add <path> --mode ... manage tracked files (copy|symlink)")
	fmt.Println("  config list                  list tracked files")
	fmt.Println("  config remove <path>         untrack a file")
	fmt.Println("  cd                           select and attach to a worktree")
	fmt.Println("  remove <branch> [--force]    delete a worktree and optionally force")
	fmt.Println("  service <subcommand>         manage per-worktree services")
}

func (a *App) runService(args []string) int {
	if len(args) == 0 {
		printServiceUsage()
		return 1
	}
	switch args[0] {
	case "add":
		return a.runServiceAdd(args[1:])
	case "start":
		return a.runServiceStart(args[1:])
	case "stop":
		return a.runServiceStop(args[1:])
	case "list":
		return a.runServiceList(args[1:])
	case "attach":
		return a.runServiceAttach(args[1:])
	case "remove":
		return a.runServiceRemove(args[1:])
	case "definitions":
		return a.runServiceDefinitions(args[1:])
	default:
		fmt.Println("unknown service command:", args[0])
		printServiceUsage()
		return 1
	}
}

func (a *App) runServiceAdd(args []string) int {
	fs := flag.NewFlagSet("service add", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	command := fs.String("command", "", "command to run")
	port := fs.String("port", "auto", "auto|none|<number>")

	if err := fs.Parse(reorderServiceAddArgs(args)); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}

	if fs.NArg() < 1 || *command == "" {
		fmt.Println("usage: gwm service add <name> --command \"...\" [--port auto|none|<number>]")
		return 1
	}

	name := fs.Arg(0)
	portMode, fixedPort := parsePort(*port)

	if err := a.ServiceAdd.Execute(usecase.ServiceAddInput{
		Name:      name,
		Command:   *command,
		PortMode:  portMode,
		FixedPort: fixedPort,
	}); err != nil {
		fmt.Println("error:", err)
		return 1
	}

	fmt.Println("added service:", name)
	return 0
}

func (a *App) runServiceStart(args []string) int {
	if len(args) < 1 {
		fmt.Println("usage: gwm service start <name>")
		return 1
	}

	name := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("error:", err)
		return 1
	}

	out, err := a.ServiceStart.Execute(usecase.ServiceStartInput{
		Name:         name,
		WorktreePath: cwd,
	})
	if err != nil {
		fmt.Println("error:", err)
		return 1
	}

	for _, m := range out.Messages {
		fmt.Println(m)
	}
	return 0
}

func (a *App) runServiceStop(args []string) int {
	if len(args) < 1 {
		fmt.Println("usage: gwm service stop <name>")
		return 1
	}

	name := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("error:", err)
		return 1
	}

	out, err := a.ServiceStop.Execute(usecase.ServiceStopInput{
		Name:         name,
		WorktreePath: cwd,
	})
	if err != nil {
		fmt.Println("error:", err)
		return 1
	}

	for _, m := range out.Messages {
		fmt.Println(m)
	}
	return 0
}

func (a *App) runServiceList(args []string) int {
	out, err := a.ServiceList.Execute()
	if err != nil {
		fmt.Println("error:", err)
		return 1
	}

	if len(out.Services) == 0 {
		fmt.Println("no running services")
		return 0
	}

	data, _ := json.MarshalIndent(out.Services, "", "  ")
	fmt.Println(string(data))
	return 0
}

func (a *App) runServiceAttach(args []string) int {
	if len(args) < 1 {
		fmt.Println("usage: gwm service attach <name>")
		return 1
	}

	name := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("error:", err)
		return 1
	}

	if err := a.ServiceAttach.Execute(usecase.ServiceAttachInput{
		Name:         name,
		WorktreePath: cwd,
	}); err != nil {
		fmt.Println("error:", err)
		return 1
	}
	return 0
}

func (a *App) runServiceRemove(args []string) int {
	if len(args) < 1 {
		fmt.Println("usage: gwm service remove <name>")
		return 1
	}

	if err := a.ServiceRemove.Execute(usecase.ServiceRemoveInput{
		Name: args[0],
	}); err != nil {
		fmt.Println("error:", err)
		return 1
	}

	fmt.Println("removed service definition:", args[0])
	return 0
}

func (a *App) runServiceDefinitions(args []string) int {
	out, err := a.ServiceDefinitionList.Execute()
	if err != nil {
		fmt.Println("error:", err)
		return 1
	}

	if len(out.Definitions) == 0 {
		fmt.Println("no service definitions")
		return 0
	}

	data, _ := json.MarshalIndent(out.Definitions, "", "  ")
	fmt.Println(string(data))
	return 0
}

func reorderServiceAddArgs(args []string) []string {
	return moveFirstPositionalArgToEnd(args)
}

func parsePort(s string) (domain.PortMode, int) {
	if s == "" || s == "auto" {
		return domain.PortModeAuto, 0
	}
	if s == "none" {
		return domain.PortModeNone, 0
	}
	port, err := strconv.Atoi(s)
	if err != nil {
		return domain.PortModeAuto, 0
	}
	return domain.PortModeFixed, port
}

func printServiceUsage() {
	fmt.Println("usage: gwm service <command>")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  add <name> --command \"...\" [--port auto|none|<n>]  register service")
	fmt.Println("  start <name>                                        start service")
	fmt.Println("  stop <name>                                         stop service")
	fmt.Println("  list                                                 list running services")
	fmt.Println("  attach <name>                                        attach to service session")
	fmt.Println("  remove <name>                                        remove service definition")
	fmt.Println("  definitions                                          list service definitions")
}
