// Package cli implements the CLI functionality, main defers
// execution to the exported methods in this package
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/FollowTheProcess/py/pkg/py"
)

var (
	version = "dev" // py version, set at compile time by ldflags
	commit  = ""    // py version's commit hash, set at compile time by ldflags
)

const (
	vitualEnvKey = "VIRTUAL_ENV" // The key for the python activated venv environment variable
	helpText     = `
Python launcher for Unix

Launch your python interpreter the lazy/smart way 🚀

py is meant to become your go-to command for launching a python interpreter
while writing code.

It does this by trying to find the python interpreter that you most likely
want to use by looking in a few different places:

1) Passed version as an argument
2) An activated virtual environment
3) A virtual environment in the current or parent directories
4) The shebang of the target file (if relevant)
5) The latest version of python on $PATH

The full control flow can be found in the documentation.

If py reaches the end of the list without finding a valid interpreter,
it will exit with an error message.

Usage:

  py [args] [flags]

Examples:

# Launch the latest version of python on $PATH (or a virtual environment)
$ py

# Launch the latest python3 on $PATH
$ py -3

# Launch a specific version on $PATH
$ py -3.10

# Can use normal python flags
$ py -m venv .venv

# List all found interpreters
$ py --list

Flags:
  --help      Help for py
  --list      List all found python interpreters on $PATH
  --version   Show py's version info

Note: This is not *the* python launcher as in brettcannon/python-launcher,
this is FollowTheProcess/python-launcher - an (approximate) port of the original
to Go.
`
)

// App represents the py program
type App struct {
	Out io.Writer
}

// New creates a new default App configured to talk to os.Stdout
func New() *App {
	return &App{Out: os.Stdout}
}

// Version shows py's version information
func (a *App) Version() {
	fmt.Fprintf(a.Out, "py version: %s\n", version)
	fmt.Fprintf(a.Out, "commit: %s\n", commit)
}

// Help shows py's help text and usage info
func (a *App) Help() {
	fmt.Fprintln(a.Out, helpText)
}

// List is the handler for the list command
func (a *App) List() error {
	paths, err := py.GetPath()
	if err != nil {
		return fmt.Errorf("could not get $PATH: %w", err)
	}

	found, err := py.GetAllPythonInterpreters(paths)
	if err != nil {
		return fmt.Errorf("error getting python interpreters: %w", err)
	}

	// Ensure interpreters are sorted latest to oldest regardless of
	// any filepath based sorting from ReadDir
	sort.Sort(found)

	for _, interpreter := range found {
		fmt.Fprintln(a.Out, interpreter)
	}

	return nil
}

// LaunchREPL will follow py's control flow and launch whatever is the most appropriate python REPL
// Control flow is:
// 	1) Activated virtual environment
// 	2) .venv directory
// 	3) PY_PYTHON env variable
// 	4) Latest version on $PATH
func (a *App) LaunchREPL() error {
	// Here we follow the control flow specified, returning to the caller
	// on the first matched condition, thus preventing later conditions
	// from evaluating. This ensures our order of priority is followed

	// Activated virtual environment
	if path := os.Getenv(vitualEnvKey); path != "" {
		exe := filepath.Join(path, "bin", "python")
		if err := launch(exe, []string{}); err != nil {
			return err
		}
		return nil
	}

	// Directory called .venv in cwd
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting cwd: %w", err)
	}
	exe := py.GetVenvDir(cwd)
	if exe != "" {
		// Means we found a python interpreter inside .venv, so launch it
		if err := launch(exe, []string{}); err != nil {
			return err
		}
		return nil
	}

	// PY_PYTHON env variable specifying a X.Y version identifier
	// e.g. 3.10
	// TODO: This

	return nil
}

// LaunchLatest will search through $PATH, find the latest python interpreter
// and launch it
func (a *App) LaunchLatest() error {
	path, err := py.GetPath()
	if err != nil {
		return fmt.Errorf("could not get $PATH: %w", err)
	}
	interpreters, err := py.GetAllPythonInterpreters(path)
	if err != nil {
		return fmt.Errorf("error fetching python interpreters: %w", err)
	}

	sort.Sort(interpreters)

	latest := interpreters[0]

	if err := launch(latest.Path, []string{}); err != nil {
		return err
	}

	return nil
}

// launch will launch a python interpreter at a specific (absolute) path
// and forward any args to the called interpreter. If no args required
// just pass an empty slice
func launch(path string, args []string) error {
	// We must use syscall.Exec here as we must "swap" the process to python
	// simply running a subprocess e.g. (os/exec), even without waiting
	// for the subprocess to complete, will not work
	if err := syscall.Exec(path, args, []string{}); err != nil {
		return fmt.Errorf("error launching %s: %w", path, err)
	}
	return nil
}
