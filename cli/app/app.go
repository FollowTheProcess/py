// Package app implements the CLI functionality, the CLI defers
// execution to the exported methods in this package
package app

import (
	"fmt"
	"io"
	"os"
	"sort"
	"syscall"

	"github.com/FollowTheProcess/py/pkg/py"
)

// App represents the py program
type App struct {
	Out io.Writer
}

// New creates a new default App configured to talk to os.Stdout
func New() *App {
	return &App{Out: os.Stdout}
}

// Launch is the handler for the main program entry point
func (a *App) Launch(args []string) error {
	// If no args given, launch a REPL with the latest interpreter
	if len(args) == 0 {
		if err := launchLatest(); err != nil {
			return err
		}
	}

	return nil
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

// launchLatest is a convenience function to launch the latest python
// interpreter found on $PATH
func launchLatest() error {
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

	// We must use syscall.Exec here as we must "swap" to python
	// simply running a subprocess will not work how the user expects
	if err := syscall.Exec(latest.Path, []string{}, []string{}); err != nil {
		return fmt.Errorf("error launching %s: %w", latest.Path, err)
	}

	return nil
}
