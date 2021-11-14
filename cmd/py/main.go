package main

import (
	"fmt"
	"os"

	"github.com/FollowTheProcess/py/cli"
)

func main() {
	// Note: because we require passing a version specifier (e.g. -X or -X.Y)
	// we can't use the stdlib flag or spf13 pflag packages as these will get
	// confused when passed a -3.9 for example
	// So we handle everything as an argument and deal with the logic manually

	// Run the program, passing all args (other than the binary name) to run
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	// Initialise the cli
	app := cli.New()

	n := len(args)

	// No args, follow control flow to find version of the REPL to launch
	if n == 0 {
		if err := app.LaunchREPL(); err != nil {
			return fmt.Errorf("%w", err)
		}
		return nil
	}

	// 1 arg, most of the work
	// check if matches -X/-X.Y or a supported flag and handle it if it does
	// if not, it could be a file (e.g. python script.py) or a python arg (e.g. python -m venv .venv)
	// in which case, if we have a -X/-X.Y pass all other args to this python, else latest python
	if n == 1 {
		switch arg := args[0]; {
		case arg == "--help":
			app.Help()
		case arg == "--list":
			if err := app.List(); err != nil {
				return fmt.Errorf("%w", err)
			}
		case arg == "--version":
			app.Version()
		default:
			fmt.Printf("default case hit. Arg: %s\n", arg)
		}

		return nil
	}

	// > 1 arg
	// If first arg is -X/-X.Y start this python and pass all args through
	// if not start latest $PATH python and pass all args through
	fmt.Println("> 1 arg, if first matches -X/-X.Y, use this python and pass args through")
	fmt.Println("if not, just pass all args through to latest $PATH python")

	return nil
}
