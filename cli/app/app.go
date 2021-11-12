// Package app implements the CLI functionality, the CLI defers
// execution to the exported methods in this package
package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	fmt.Fprintf(a.Out, "Launching: %v\n", args)
	return nil
}

// List is the handler for the list command
func (a *App) List() error {
	fmt.Fprintln(a.Out, "Here is your path:")
	path := os.Getenv("PATH")
	for i, dir := range filepath.SplitList(path) {
		if dir == "" {
			dir = "."
		}
		fmt.Fprintf(a.Out, "%d: %s\n", i, dir)
	}
	return nil
}
