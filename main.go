// Package manager
// Code based on sources of cmd/go and source of godep project.
package main

import (
	"flag"
	"fmt"
	"go/build"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/go-env/pkglib"
)

var (
	goroot    = filepath.Clean(runtime.GOROOT())
	gobin     = os.Getenv("GOBIN")
	gorootBin = filepath.Join(goroot, "bin")
	gorootPkg = filepath.Join(goroot, "pkg")
	gorootSrc = filepath.Join(goroot, "src")
)

// A Command is an implementation of a gopkg command
// type Command struct {
// 	// Run runs the command.
// 	// The args are the arguments after the command name.
// 	Run func(cmd *Command, args []string)

// 	// Usage is the one-line usage message.
// 	// The first word in the line is taken to be the command name.
// 	UsageLine string

// 	// Short is the short description shown in the 'gopkg help' output.
// 	Short string

// 	// Long is the long message shown in the
// 	// 'gopkg help <this-command>' output.
// 	Long string

// 	// Flag is a set of flags specific to this command.
// 	Flag flag.FlagSet

// 	// CustomFlags indicates that the command will do its own
// 	// flag parsing.
// 	CustomFlags bool
// }

// func (c *Command) Name() string {
// 	name := c.UsageLine
// 	i := strings.Index(name, " ")
// 	if i >= 0 {
// 		name = name[:i]
// 	}
// 	return name
// }

// func (c *Command) UsageExit() {
// 	fmt.Fprintf(os.Stderr, "Usage: gopkg %s\n\n", c.UsageLine)
// 	fmt.Fprintf(os.Stderr, "Run 'gopkg help %s' for help.\n", c.Name())
// 	os.Exit(2)
// }

// Commands lists the available commands and help topics.
// The order here is the order in which they are printed
// by 'gopkg help'.
var commands = []*pkglib.Command{
	//	cmdInit,
	cmdGet,
	//	cmdList,
}

func main() {
	flag.Usage = usageExit
	flag.Parse()
	log.SetFlags(0)
	log.SetPrefix("gopkg: ")
	args := flag.Args()
	if len(args) < 1 {
		usageExit()
	}

	if args[0] == "help" {
		help(args[1:])
		return
	}

	// Diagnose common mistake: GOPATH==GOROOT.
	// This setting is equivalent to not setting GOPATH at all,
	// which is not what most people want when they do it.
	if gopath := os.Getenv("GOPATH"); gopath == runtime.GOROOT() {
		fmt.Fprintf(os.Stderr, "warning: GOPATH set to GOROOT (%s) has no effect\n", gopath)
	} else {
		for _, p := range filepath.SplitList(gopath) {
			// Note: using HasPrefix instead of Contains because a ~ can appear
			// in the middle of directory elements, such as /tmp/git-1.8.2~rc3
			// or C:\PROGRA~1. Only ~ as a path prefix has meaning to the shell.
			if strings.HasPrefix(p, "~") {
				fmt.Fprintf(os.Stderr, "gopkg: GOPATH entry cannot start with shell metacharacter '~': %q\n", p)
				os.Exit(2)
			}
			if build.IsLocalImport(p) {
				fmt.Fprintf(os.Stderr, "gopkg: GOPATH entry is relative; must be absolute path: %q.\nRun 'go help gopath' for usage.\n", p)
				os.Exit(2)
			}
		}
	}

	if fi, err := os.Stat(goroot); err != nil || !fi.IsDir() {
		fmt.Fprintf(os.Stderr, "gopkg: cannot find GOROOT directory: %v\n", goroot)
		os.Exit(2)
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() { cmd.Usage() }
			if cmd.CustomFlags {
				args = args[1:]
			} else {
				cmd.Flag.Parse(args[1:])
				args = cmd.Flag.Args()
			}
			cmd.Run(cmd, args)
			exit()
			return
		}
	}

	// for _, cmd := range commands {
	// 	if cmd.Name() == args[0] {
	// 		cmd.Flag.Usage = func() { cmd.Usage() }
	// 		cmd.Flag.Parse(args[1:])
	// 		cmd.Run(cmd, cmd.Flag.Args())
	// 		return
	// 	}
	// }

	fmt.Fprintf(os.Stderr, "gopkg: unknown command %q\n", args[0])
	fmt.Fprintf(os.Stderr, "Run 'gopkg help' for usage.\n")
	os.Exit(2)
}

var usageTemplate = `
Package manager — part of GET (golang environment tools) project.

Usage:

	gopkg command [arguments]

The commands are:
{{range .}}
    {{.Name | printf "%-8s"}} {{.Short}}{{end}}

Use "gopkg help [command]" for more information about a command.
`

var helpTemplate = `
Usage: gpkg {{.UsageLine}}

{{.Long | trim}}
`

func help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		return
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: gopkg help command\n\n")
		fmt.Fprintf(os.Stderr, "Too many arguments given.\n")
		os.Exit(2)
	}
	for _, cmd := range commands {
		if cmd.Name() == args[0] {
			tmpl(os.Stdout, helpTemplate, cmd)
			return
		}
	}
}

func usageExit() {
	printUsage(os.Stderr)
	os.Exit(2)
}

func printUsage(w io.Writer) {
	tmpl(w, usageTemplate, commands)
}

// tmpl executes the given template text on data, writing the result to w.
func tmpl(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{
		"trim": strings.TrimSpace,
	})
	template.Must(t.Parse(strings.TrimSpace(text) + "\n\n"))
	if err := t.Execute(w, data); err != nil {
		panic(err)
	}
}

var (
	exitStatus  = 0
	atexitFuncs []func()
)

func exit() {
	for _, f := range atexitFuncs {
		f()
	}
	os.Exit(exitStatus)
}

func exitIfErrors() {
	if exitStatus != 0 {
		exit()
	}
}
