package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/FollowTheProcess/py/cli"
)

// TODO: Add a logger and support the PY_DEBUG environment variable

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

	// No arguments, means the user wants to launch a REPL
	// follow control flow to find which version to launch
	if n == 0 {
		fmt.Println("0 arguments, following control flow to launch REPL")
		if err := app.Launch([]string{}); err != nil {
			return fmt.Errorf("%w", err)
		}
		return nil
	}

	// We have a single command line argument which could mean several things:
	// 1) known flag (e.g. --list)
	// 2) version specifier of the form -X or -X.Y
	// 3) file e.g. py script.py
	if n == 1 {
		arg := args[0]
		if err := handleSingleArg(app, arg); err != nil {
			return err
		}
		return nil
	}

	// If we get here we have more than 1 argument, which could mean a few things
	// depending on what the first argument is:
	// 1) Known flag: error out as they do not support additional arguments
	// 2) Version specifier (-X or -X.Y): Launch matching version and pass all other args through
	// 3) Unknown: Follow control flow to find a python and pass all args through (also need to check for file)
	if err := handleMultipleArgs(app, args); err != nil {
		return err
	}
	return nil
}

// handleSingleArg handles the case where py is passed a single command line argument
// which could mean several things:
// 	1) known flag (e.g. --list)
// 	2) version specifier of the form -X or -X.Y
// 	3) file e.g. py script.py
func handleSingleArg(app *cli.App, arg string) error {
	switch {
	case arg == "--help":
		app.Help()

	case arg == "--list":
		if err := app.List(); err != nil {
			return fmt.Errorf("%w", err)
		}

	case arg == "--version":
		app.Version()

	case isMajorSpecifier(arg):
		// User has passed something like -3
		fmt.Printf("Detected single arg: major version specifier: %s\n", arg)
		major := parseMajorSpecifier(arg)
		fmt.Printf("Launching REPL with major version: %d\n", major)
		if err := app.LaunchMajor(major, []string{}); err != nil {
			return fmt.Errorf("%w", err)
		}

	case isExactSpecifier(arg):
		// User has passed something like -3.10
		fmt.Printf("Detected single arg: exact version specifier: %s\n", arg)
		major, minor := parseExactSpecifier(arg)
		fmt.Printf("Launching REPL with exact version: %d.%d\n", major, minor)
		if err := app.LaunchExact(major, minor, []string{}); err != nil {
			return fmt.Errorf("%w", err)
		}

	default:
		// If we got here, the argument must be a file (e.g. py script.py)
		// in which call python with the file as the argument
		// TODO: the additional control flow element here is it could be a file, so check and look for a shebang
		// could also be a single python flag e.g. python -V for version info
		fmt.Printf("Detected single unrecognised arg %s, following control flow and passing arg on\n", arg)
		if err := app.Launch([]string{arg}); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

// handleMultipleArgs handles the case in which py was passed > 1 command line argument
// which could mean a few things depending on what the first argument is:
// 	1) Known flag: error out as they do not support arguments
// 	2) Version specifier (-X or -X.Y): Launch matching version and pass all other args through
// 	3) Unknown: Launch latest and pass all args through
func handleMultipleArgs(app *cli.App, args []string) error {
	rest := args[1:]
	switch first := args[0]; {
	case first == "--help":
		return fmt.Errorf("cannot use --help with any other arguments")
	case first == "--list":
		return fmt.Errorf("cannot use --list with any other arguments")
	case first == "--version":
		return fmt.Errorf("cannot use --version with any other arguments")

	case isMajorSpecifier(first):
		// User has passed something like "py -3 first ..."
		fmt.Printf("Detected multiple args: first is major specifier: %s, rest are: %v\n", first, rest)
		major := parseMajorSpecifier(first)
		// Strip off the major version specifier and pass remaining args through
		fmt.Printf("Launching major version %d with args: %v\n", major, rest)
		if err := app.LaunchMajor(major, args[1:]); err != nil {
			return fmt.Errorf("%w", err)
		}

	case isExactSpecifier(first):
		// User has passed something like "py -3.10 first ..."
		fmt.Printf("Detected multiple args: first is exact specifier: %s, rest are %v\n", first, rest)
		major, minor := parseExactSpecifier(first)
		// Strip off the exact version specifier and pass remaining args through
		fmt.Printf("Launching exact version %d.%d with args: %v\n", major, minor, rest)
		if err := app.LaunchExact(major, minor, rest); err != nil {
			return fmt.Errorf("%w", err)
		}

	default:
		// If we get here it's unknown args
		// in which case follow the control flow, launch the resulting python
		// and pass all the arguments through
		fmt.Printf("Multiple arguments, first is unrecognised, following control flow passing args on: %v\n", args)
		if err := app.Launch(args); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

// isMajorSpecifier determines if the argument passed to it
// is a valid major version specifier (e.g. "-3")
func isMajorSpecifier(arg string) bool {
	// If we don't start with a "-" it's not a major specifier
	if !strings.HasPrefix(arg, "-") {
		return false
	}

	// If, after removing the "-", it's not just 1 character, it's not a major specifier
	// this will of course catch 2 digit integers, but I don't see python10
	// coming any time soon
	arg = arg[1:]
	if len(arg) != 1 {
		return false
	}

	// If we can't convert whats left to an integer it's not a major specifier
	if _, err := strconv.Atoi(arg); err != nil {
		return false
	}

	return true
}

// parseMajorSpecifier takes in an argument we already know to be a major specifier
// and returns the integer version.
//
// In the interest of performance, thisfunction assumes that 'arg' is already a valid
// major version specifier in string form
func parseMajorSpecifier(arg string) int {
	// Remove the "-"
	arg = arg[1:]

	// We ignore the error here because this will only get called
	// in the case that isMajorSpecifier has evaluated to true
	major, _ := strconv.Atoi(arg)

	return major
}

// isExactSpecifier determines if the argument passed to it
// is a valid exact version specifier (e.g. "-3.9")
func isExactSpecifier(arg string) bool {
	// If we don't start with a "-" it's not a major specifier
	if !strings.HasPrefix(arg, "-") {
		return false
	}

	// Remove the "-"
	arg = arg[1:]

	// Whats remaining needs to be "X.Y"
	parts := strings.Split(arg, ".")
	if len(parts) != 2 {
		return false
	}

	major, minor := parts[0], parts[1]

	// Both parts need to be an integer
	if _, err := strconv.Atoi(major); err != nil {
		return false
	}

	if _, err := strconv.Atoi(minor); err != nil {
		return false
	}

	return true
}

// parseExactSpecifier takes in an argument we already know to be an exact version specifier
// and returns the integer representations.
//
// In the interest of performance, this function assumes that 'arg' is already a valid
// minor version specifier in string form
func parseExactSpecifier(arg string) (int, int) {
	// Remove the "-"
	arg = arg[1:]

	parts := strings.Split(arg, ".")
	major, minor := parts[0], parts[1]

	// We ignore the error here because this will only get called
	// in the case that isExactSpecifier has evaluated to true
	majorInt, _ := strconv.Atoi(major)
	minorInt, _ := strconv.Atoi(minor)

	return majorInt, minorInt
}
