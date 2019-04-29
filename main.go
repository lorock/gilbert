package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/go-gilbert/gilbert/cli/scaffold"
	"github.com/go-gilbert/gilbert/cli/tasks"
	"github.com/go-gilbert/gilbert/log"
	"github.com/go-gilbert/gilbert/scope"
	"github.com/urfave/cli"
)

var (
	// These values will override by linker
	version = "dev"
	commit  = "local build"
)

// unfortunately, urface/cli ignores '--verbose' global flag :(
// so it should be defined implicitly in each task
var verboseFlag = cli.BoolFlag{
	Name:        "verbose",
	Usage:       "shows debug information, useful for troubleshooting",
	Destination: &scope.Debug,
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	app := cli.NewApp()
	app.Name = "gilbert"
	app.Usage = "Build automation tool for Go"
	app.Version = version
	app.HideVersion = true
	app.Commands = []cli.Command{
		{
			Name:        "version",
			Description: "shows application version",
			Usage:       "Shows application version",
			Action:      printVersion,
		},
		{
			Name:        "run",
			Description: "Runs a task declared in manifest file",
			Usage:       "Runs a task declared in manifest file",
			Action:      tasks.RunTask,
			Before:      bootstrap,
			Flags: []cli.Flag{
				verboseFlag,
			},
		},
		{
			Name:        "ls",
			Description: "Lists all tasks defiled in gilbert.yaml",
			Usage:       "Lists all tasks defiled in gilbert.yaml",
			Action:      tasks.ListTasksAction,
			Before:      bootstrap,
			Flags: []cli.Flag{
				verboseFlag,
			},
		},
		{
			Name:        "init",
			Description: "Scaffolds a new gilbert.yaml file",
			Usage:       "Scaffolds a new gilbert.yaml file",
			Action:      scaffold.RunScaffoldManifest,
			Before:      bootstrap,
			Flags: []cli.Flag{
				verboseFlag,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		color.Red("ERROR: %v", err)
	}
}

func bootstrap(_ *cli.Context) error {
	level := log.LevelInfo
	if scope.Debug {
		level = log.LevelDebug
	}

	log.UseConsoleLogger(level)
	return nil
}

func printVersion(_ *cli.Context) error {
	fmt.Printf("Gilbert version %s (%s)\n", version, commit)
	return nil
}
