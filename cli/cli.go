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

	"github.com/FollowTheProcess/py/internal"
	"github.com/FollowTheProcess/py/pkg/interpreter"
	"github.com/sirupsen/logrus"
)

var (
	version = "dev" // py version, set at compile time by ldflags
	commit  = ""    // py version's commit hash, set at compile time by ldflags
)

const (
	vitualEnvKey   = "VIRTUAL_ENV"    // The key for the python activated venv environment variable
	pathEnvKey     = "PATH"           // The key for the os $PATH environment variable
	debugEnvKey    = "PYLAUNCH_DEBUG" // The key for the env variable to trigger verbose logging
	pyPythonEnvKey = "PY_PYTHON"      // The key for py's default python environment variable
	helpText       = `
Python launcher for Unix (The experimental Go port!)

Launch your python interpreter the lazy/smart way 🚀

py is meant to become your go-to command for launching a python interpreter
while writing code.

It does this by trying to find the python interpreter that you most likely
want to use by looking in a few different places:

1) Passed version as an argument
2) An activated virtual environment
3) A virtual environment in the current directory
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
`
)

// App represents the py program
type App struct {
	Out    io.Writer
	Logger *logrus.Logger
}

// New creates a new default App configured to talk to os.Stdout
func New() *App {
	l := logrus.New()

	// If the PYLAUNCH_DEBUG environment variable is set to anything
	// set logging level accordingly, otherwise leave at default (InfoLevel)
	if debug := os.Getenv(debugEnvKey); debug != "" {
		l.Level = logrus.DebugLevel
	}
	l.Formatter = &logrus.TextFormatter{DisableLevelTruncation: true, DisableTimestamp: true}

	return &App{Out: os.Stdout, Logger: l}
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
	a.Logger.Debugln("Checking PATH environment variable")
	paths, err := interpreter.GetPath(pathEnvKey)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	a.Logger.Debugf("$PATH: %v\n", paths)

	a.Logger.Debugln("Looking through PATH for python3 interpreters")
	found, err := interpreter.GetAll(paths)
	if err != nil {
		return fmt.Errorf("error getting python interpreters: %w", err)
	}

	// Handle the case where the user does not have any pythons
	if len(found) == 0 {
		return fmt.Errorf("no python interpreters found on $PATH")
	}

	// Ensure interpreters are sorted latest to oldest regardless of
	// any filepath based sorting from ReadDir
	sort.Sort(found)

	for _, interpreter := range found {
		fmt.Fprintln(a.Out, interpreter)
	}

	return nil
}

// Launch will follow py's control flow and launch whatever is the most appropriate python
// any arguments specified in 'args' will be passed through to the found python
// Control flow for no args is:
// 	1) Activated virtual environment
// 	2) .venv directory
// 	3) venv directory
// 	4) Look for a python shebang line in the file (if we have a file
// 	5) PY_PYTHON env variable
// 	6) Latest version on $PATH
func (a *App) Launch(args []string) error {
	// Here we follow the control flow specified, returning to the caller
	// on the first matched condition, thus preventing later conditions
	// from evaluating. This ensures our order of priority is followed

	// 1) Activated virtual environment, as marked by the presence of
	// an environment variable $VIRTUAL_ENV pointing to the directory
	// e.g. /Users/you/Projects/thisproject/.venv
	a.Logger.Debugln("Looking for $VIRTUAL_ENV")
	if path := os.Getenv(vitualEnvKey); path != "" {
		a.Logger.WithField("$VIRTUAL_ENV", path).Debugln("Found environment variable")
		exe := filepath.Join(path, "bin", "python")
		a.Logger.WithField("interpreter", exe).Debugln("Launching python interpreter")
		if err := launch(exe, args); err != nil {
			return err
		}
		return nil
	}

	// 2) & 3) Directory called .venv or venv in cwd
	a.Logger.Debugln("Looking for a .venv or venv in cwd")
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting cwd: %w", err)
	}
	exe := internal.GetVenvPython(cwd)
	if exe != "" {
		// Means we found a python interpreter inside .venv, so launch it and pass on any args
		a.Logger.WithField("interpreter", exe).Debugln("Launching python interpreter")
		if err := launch(exe, args); err != nil {
			return err
		}
		return nil
	}

	// 4) If first arg is a file, look for a python shebang line
	// TODO: Figure out how to implement a quick shebang check and get a version back out

	// 5) PY_PYTHON env variable specifying a X.Y version identifier e.g. 3.10
	a.Logger.Debugln("Looking for $PY_PYTHON environment variable")
	if version := os.Getenv(pyPythonEnvKey); version != "" {
		a.Logger.WithField("$PY_PYTHON", version).Debugln("Found environment variable")
		major, minor, err := internal.ParsePyPython(version)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		// We're good to go
		if err := a.LaunchExact(major, minor, args); err != nil {
			return err
		}
		return nil
	}

	// 6) Launch latest on $PATH and pass the args through
	a.Logger.Debugln("Falling back to latest python on $PATH")
	if err := a.LaunchLatest(args); err != nil {
		return err
	}

	// If we get here, user has no python so return an error
	return fmt.Errorf("no python interpreters found after executing control flow")
}

// LaunchLatest will search through $PATH, find the latest python interpreter
// and launch it, with optional arguments provided
func (a *App) LaunchLatest(args []string) error {
	a.Logger.Debugln("Searching for latest python on $PATH")
	path, err := interpreter.GetPath(pathEnvKey)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	interpreters, err := interpreter.GetAll(path)
	if err != nil {
		return fmt.Errorf("error fetching python interpreters: %w", err)
	}

	// Handle the case where none are found
	if len(interpreters) == 0 {
		return fmt.Errorf("no python interpreters found on $PATH")
	}

	sort.Sort(interpreters)

	latest := interpreters[0]

	a.Logger.WithField("latest", latest).Debugln("Launching latest python")

	if err := launch(latest.Path, args); err != nil {
		return err
	}

	return nil
}

// LaunchMajor will search through $PATH, find the latest python interpreter
// satisfying the constraint imposed by 'major' version passed
func (a *App) LaunchMajor(major int, args []string) error {
	a.Logger.WithField("major", major).Debugln("Searching for latest major version")
	path, err := interpreter.GetPath(pathEnvKey)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	interpreters, err := interpreter.GetAll(path)
	if err != nil {
		return fmt.Errorf("error fetching python interpreters: %w", err)
	}

	// Create and populate a list of all the python interpreters that
	// satisfy the specify major version
	var supportingInterpreters interpreter.List
	for _, python := range interpreters {
		if python.SatisfiesMajor(major) {
			supportingInterpreters = append(supportingInterpreters, python)
		}
	}

	// Handle the case where none are found
	if len(supportingInterpreters) == 0 {
		return fmt.Errorf("no interpreters found supporting major version %d", major)
	}

	// Sort so the latest supporting interpreter is first
	sort.Sort(supportingInterpreters)

	latest := supportingInterpreters[0]

	a.Logger.WithField("interpreter", latest.Path).Debugln("Launching python")
	if err := launch(latest.Path, args); err != nil {
		return err
	}

	return nil
}

// LaunchExact will search through $PATH, find the latest python interpreter
// satisfying the constraint imposed by both 'major' and 'minor' version passed
func (a *App) LaunchExact(major, minor int, args []string) error {
	a.Logger.Debugf("Searching for python %d.%d", major, minor)
	path, err := interpreter.GetPath(pathEnvKey)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	interpreters, err := interpreter.GetAll(path)
	if err != nil {
		return fmt.Errorf("error fetching python interpreters: %w", err)
	}

	// Create and populate a list of all the python interpreters that
	// satisfy the specify major version
	var supportingInterpreters interpreter.List
	for _, python := range interpreters {
		if python.SatisfiesExact(major, minor) {
			supportingInterpreters = append(supportingInterpreters, python)
		}
	}

	// Handle the case where none are found
	if len(supportingInterpreters) == 0 {
		return fmt.Errorf("no interpreters found supporting exact version %d.%d", major, minor)
	}

	// Sort so the latest supporting interpreter is first
	sort.Sort(supportingInterpreters)

	latest := supportingInterpreters[0]

	a.Logger.WithField("python", latest.Path).Debugln("Launching exact python")
	if err := launch(latest.Path, args); err != nil {
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

	// Note on syscall.Exec here as this was not obvious to me until I looked up
	// https://pkg.go.dev/golang.org/x/sys@v0.0.0-20211113001501-0c823b97ae02/unix#Exec
	// argv0 is the absolute path to the executable as expected
	// argv is a string slice with the name of argv0 as the first element and the intended args as the rest
	// so correct usage is something like syscall.Exec("/usr/bin/ls", "ls -l")
	argv := []string{filepath.Base(path)}
	argv = append(argv, args...)
	if err := syscall.Exec(path, argv, []string{}); err != nil {
		return fmt.Errorf("error launching %s: %w", path, err)
	}
	return nil
}
