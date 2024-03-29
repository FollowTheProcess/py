// Package cli implements the CLI functionality, main defers
// execution to the exported methods in this package
package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/FollowTheProcess/py/interpreter"
	"github.com/sirupsen/logrus"
)

var (
	version    = "dev"                            // py version, set at compile time by ldflags
	commit     = ""                               // py version's commit hash, set at compile time by ldflags
	majorRegex = regexp.MustCompile(`^\d+$`)      // The regex for a single major version specifier in string form e.g "3"
	exactRegex = regexp.MustCompile(`^\d+\.\d+$`) // The regex for an exact version specifier in string form e.g "3.9"
	helpText   = fmt.Sprintf(`
Python launcher for Unix (The experimental Go port!)

Version: %s
Commit: %s

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

# Follow the control flow and launch the python it finds
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

Environment Variables:
	PY_PYTHON        The version of python you wish to be the default (e.g. "3.10")
	PYLAUNCH_DEBUG   If set to anything will print debug information to stderr
	`, version, commit)
)

const (
	vitualEnvKey   = "VIRTUAL_ENV"    // The key for the python activated venv environment variable
	debugEnvKey    = "PYLAUNCH_DEBUG" // The key for the env variable to trigger verbose logging
	pyPythonEnvKey = "PY_PYTHON"      // The key for py's default python environment variable

	xYParts = 2 // Number of parts in an X.Y version specifier
	xParts  = 1 // Number of parts in an X version specifier

)

// App represents the py program.
type App struct {
	Stdout io.Writer      // Normal CLI output
	Stderr io.Writer      // Where the logger and errors will write to
	Logger *logrus.Logger // The debug logger
	Path   string         // The path to search through i.e. $PATH, passable field to facilitate testing
}

// New creates a new default App configured to write to 'stdout' and DEBUG log to 'stderr'.
func New(stdout, stderr io.Writer) *App {
	log := logrus.New()

	// Get the value of $PATH
	path := os.Getenv("PATH")

	// If the PYLAUNCH_DEBUG environment variable is set to anything
	// set logging level accordingly, otherwise leave at default (InfoLevel)
	if debug := os.Getenv(debugEnvKey); debug != "" {
		log.Level = logrus.DebugLevel
	}
	log.Formatter = &logrus.TextFormatter{DisableLevelTruncation: true, DisableTimestamp: true}
	log.Out = stderr

	return &App{Stdout: stdout, Stderr: stderr, Logger: log, Path: path}
}

// Help shows py's help text and usage info.
func (a *App) Help() {
	fmt.Fprintln(a.Stdout, helpText)
}

// List shows a list of all python interpreters on $PATH, sorted latest to oldest.
func (a *App) List() error {
	interpreters, err := a.getAllPythonInterpreters()
	if err != nil {
		return err
	}

	// Handle the case where the user does not have any pythons
	if len(interpreters) == 0 {
		return fmt.Errorf("no python interpreters found on $PATH")
	}
	// Ensure interpreters are sorted latest to oldest regardless of
	// any filepath based sorting from ReadDir
	interpreter.Sort(interpreters)

	a.Logger.WithField("interpreters", interpreters).Debugln("Found python3 interpreters")

	for _, interpreter := range interpreters {
		fmt.Fprintln(a.Stdout, interpreter.ToString())
	}

	return nil
}

// Launch will follow py's control flow and launch whatever is the most appropriate python
// any arguments specified in 'args' will be passed through to the found python
// Control flow for no args is:
//  1. Activated virtual environment
//  2. .venv directory
//  3. venv directory
//  4. Look for a python shebang line in the file (if we have a file)
//  5. PY_PYTHON env variable
//  6. Latest version on $PATH
func (a *App) Launch(args []string) error {
	// Here we follow the control flow specified, returning to the caller
	// on the first matched condition, thus preventing later conditions
	// from evaluating. This ensures our order of priority is followed

	// 1) Activated virtual environment, as marked by the presence of
	// an environment variable $VIRTUAL_ENV pointing to the directory
	// e.g. /Users/you/Projects/thisproject/.venv
	a.Logger.Debugln("Looking for $VIRTUAL_ENV environment variable")
	if path := os.Getenv(vitualEnvKey); path != "" {
		a.Logger.WithField("$VIRTUAL_ENV", path).Debugln("Found environment variable")
		exe := filepath.Join(path, "bin", "python")
		a.Logger.WithFields(logrus.Fields{"interpreter": exe, "arguments": args}).Debugln("Launching python interpreter with arguments")
		return launch(exe, args)
	}

	// 2) & 3) Directory called .venv or venv in cwd
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting cwd: %w", err)
	}

	a.Logger.WithField("cwd", cwd).Debugln("Looking for virtual environment in cwd")

	exe := a.getVenvPython(cwd)
	if exe != "" {
		// Means we found a python interpreter inside .venv, so launch it and pass on any args
		a.Logger.WithFields(logrus.Fields{"interpreter": exe, "arguments": args}).Debugln("Launching python interpreter with arguments")
		return launch(exe, args)
	}

	// 4) If first arg is a file, look for a python shebang line
	if len(args) == 1 {
		if exists(args[0]) {
			// We have a file as the argument
			if err := a.handlePotentialShebang(args); err != nil {
				return err
			}
			// Note: we don't return nil here as we want to carry on the control flow
		}
	}

	// 5) PY_PYTHON env variable specifying a X.Y version identifier e.g. 3.10
	a.Logger.Debugln("Looking for $PY_PYTHON environment variable")
	if version := os.Getenv(pyPythonEnvKey); version != "" {
		a.Logger.WithField("$PY_PYTHON", version).Debugln("Found environment variable")
		major, minor, err := a.parsePyPython(version)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		// We're good to go
		return a.LaunchExact(major, minor, args)
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
// and launch it, passing through any arguments passed to it.
func (a *App) LaunchLatest(args []string) error {
	interpreters, err := a.getAllPythonInterpreters()
	if err != nil {
		return err
	}

	// Handle the case where none are found
	if len(interpreters) == 0 {
		return fmt.Errorf("no python interpreters found on $PATH")
	}

	interpreter.Sort(interpreters)

	a.Logger.WithField("interpreters", interpreters).Debugln("Found python3 interpreters")

	latest := interpreters[0]

	a.Logger.WithFields(logrus.Fields{"latest": latest, "arguments": args}).Debugln("Launching latest python with arguments")

	return launch(latest.Path, args)
}

// LaunchMajor will search through $PATH, find the latest python interpreter
// satisfying the constraint imposed by 'major' version passed
// launch it, and pass through any arguments passed to it.
func (a *App) LaunchMajor(major int, args []string) error {
	a.Logger.WithField("major", major).Debugln("Searching for latest python with major version")
	interpreters, err := a.getAllPythonInterpreters()
	if err != nil {
		return err
	}

	// Create and populate a list of all the python interpreters that
	// satisfy the specified major version
	var supportingInterpreters []interpreter.Interpreter
	for _, python := range interpreters {
		if python.SatisfiesMajor(major) {
			supportingInterpreters = append(supportingInterpreters, python)
		}
	}

	// Handle the case where none are found
	if len(supportingInterpreters) == 0 {
		return fmt.Errorf("no python%d interpreters found on $PATH", major)
	}

	// Sort so the latest supporting interpreter is first
	interpreter.Sort(supportingInterpreters)

	a.Logger.WithField("matching interpreters", supportingInterpreters).Debugln("Found matching interpreters")

	latest := supportingInterpreters[0]

	a.Logger.WithField("interpreter", latest.Path).Debugln("Launching python")
	return launch(latest.Path, args)
}

// LaunchExact will search through $PATH, find the latest python interpreter
// satisfying the constraint imposed by both 'major' and 'minor' version passed
// launch it, and pass through any args passed to it.
func (a *App) LaunchExact(major, minor int, args []string) error {
	a.Logger.WithField("version", fmt.Sprintf("%d.%d", major, minor)).Debugln("Searching for exact python version")
	interpreters, err := a.getAllPythonInterpreters()
	if err != nil {
		return err
	}

	// Create and populate a list of all the python interpreters that
	// satisfy the specify major version
	var supportingInterpreters []interpreter.Interpreter
	for _, python := range interpreters {
		if python.SatisfiesExact(major, minor) {
			supportingInterpreters = append(supportingInterpreters, python)
		}
	}

	// Handle the case where none are found
	if len(supportingInterpreters) == 0 {
		return fmt.Errorf("no python%d.%d interpreter found on $PATH", major, minor)
	}

	// Sort so the latest supporting interpreter is first
	interpreter.Sort(supportingInterpreters)

	a.Logger.WithField("matching interpreters", supportingInterpreters).Debugln("Found matching interpreters")

	latest := supportingInterpreters[0]

	a.Logger.WithField("python", latest.Path).Debugln("Launching exact python")
	return launch(latest.Path, args)
}

// getPath goes through a.Path (which it expects to be $PATH or similar)
// i.e. separated list of directories, and returns a string slice of the
// entries in that path.
//
// Entries will be de-duplicated prior to returning.
func (a *App) getPathEntries() []string {
	paths := []string{}

	for _, dir := range filepath.SplitList(a.Path) {
		if dir == "" {
			// Unix shell semantics: path element "" means "."
			dir = "."
		}
		// Some new thing on macOS that breaks us, there's a new directory
		// called /var/run/com.apple.security.cryptexd/codex.system/bootstrap/usr/local/bin
		// which cannot be read by the user, so we'll just skip it.
		if strings.HasPrefix(dir, "/var/run") {
			continue
		}
		paths = append(paths, dir)
	}

	// Dedupe
	paths = deDupe(paths)

	return paths
}

// getVenvPython will look for a ".venv/bin/python" or a "venv/bin/python"
// under the cwd, ensure that it exists and then return it's absolute path
// .venv will be preferred over venv, venv will only be used if .venv
// does not exist.
//
// If neither is found, an empty string will be returned.
func (a *App) getVenvPython(cwd string) string {
	dotVenv := filepath.Join(cwd, ".venv", "bin", "python")
	venv := filepath.Join(cwd, "venv", "bin", "python")

	switch {
	case exists(dotVenv):
		a.Logger.WithField("venv dir", dotVenv).Debugln("Found a virtual environment")
		return dotVenv
	case exists(venv):
		a.Logger.WithField("venv dir", venv).Debugln("Found a virtual environment")
		return venv
	default:
		return ""
	}
}

// parsePyPython is a helper that, when given the value of a valid PY_PYTHON env variable
// will return the integer major and minor version parts so we can launch it
//
// A valid value for PY_PYTHON is X.Y, the same as the exact version specifier
// e.g. "3.10"
//
// If 'version' is not a valid format, an error will be returned.
func (a *App) parsePyPython(version string) (int, int, error) {
	parts := strings.Split(version, ".")

	if len(parts) != xYParts {
		return 0, 0, fmt.Errorf("malformed PY_PYTHON: not X.Y format")
	}

	major, minor := parts[0], parts[1]

	majorInt, err := strconv.Atoi(major)
	if err != nil {
		return 0, 0, fmt.Errorf("malformed PY_PYTHON: major component not an integer")
	}

	minorInt, err := strconv.Atoi(minor)
	if err != nil {
		return 0, 0, fmt.Errorf("malformed PY_PYTHON: minor component not an integer")
	}

	// Now we're safe
	return majorInt, minorInt, nil
}

// parseShebang takes a line of text (as read from a file) and returns
// the string version of a python version it may represent
//
// If 'shebang' is not a valid shebang line, or if no python version is specified
// an empty string will be returned. This is the signal to use the remaining control flow to
// determine the appropriate python version to launch
//
// Example
//
//	sh := ParseShebang("#!/usr/local/bin/python3.9")
//	fmt.Println(sh)
//
// Output: "3.9".
func (a *App) parseShebang(shebang string) string {
	a.Logger.Debugln("Checking for a python shebang line")
	if !strings.HasPrefix(shebang, "#!") {
		return ""
	}

	// Trim off the #!
	shebang = strings.Replace(shebang, "#!", "", 1)

	// Whitespace is allowed between #! and the path e.g. #! /usr/bin/python
	shebang = strings.TrimSpace(shebang)

	acceptedPaths := [4]string{
		"python",
		"/usr/bin/python",
		"/usr/local/bin/python",
		"/usr/bin/env python",
	}

	for _, path := range acceptedPaths {
		if strings.HasPrefix(shebang, path) {
			// Valid shebang, let's see if we can get a version
			// from the end of 'path' e.g. /usr/bin/python3 -> 3
			// /usr/bin/python3.9 -> 3.9
			a.Logger.WithField("shebang", shebang).Debugln("Found python shebang line")
			version := shebang[len(path):]
			a.Logger.WithField("version", version).Debugln("Found potential python version in shebang line")
			return version
		}
	}

	return ""
}

// getAllPythonInterpreters does exactly what it says on the tin
// it searches through $PATH and returns a list of all python interpreters.
func (a *App) getAllPythonInterpreters() ([]interpreter.Interpreter, error) {
	a.Logger.Debugln("Checking $PATH environment variable")
	paths := a.getPathEntries()

	a.Logger.Debugf("$PATH: %v\n", paths)

	a.Logger.Debugln("Looking through $PATH for python interpreters")
	interpreters, err := interpreter.GetAll(paths)
	if err != nil {
		return nil, fmt.Errorf("error fetching python interpreters: %w", err)
	}

	return interpreters, nil
}

// handlePotentialShebang is called once we know the first argument is a file
// it attempts to open the file, look for a shebang line, parse it
// and launch the appropriate python interpreter
// if it does not find a valid shebang line or there is no version found in it
// it will return to signal the continuation of the control flow.
func (a *App) handlePotentialShebang(args []string) error {
	a.Logger.WithField("argument", args[0]).Debugln("argument is a file")
	file, err := os.Open(args[0])
	if err != nil {
		return fmt.Errorf("could not open %s: %w", args[0], err)
	}
	defer file.Close()

	// Read the first line from the file
	scanner := bufio.NewScanner(file)
	scanner.Scan()

	version := a.parseShebang(scanner.Text())

	switch {
	case majorRegex.MatchString(version):
		// Shebang is a major version e.g. /usr/bin/python3
		a.Logger.WithField("major version", version).Debugln("Shebang line refers to major version")
		major, err := strconv.Atoi(version)
		if err != nil {
			return fmt.Errorf("shebang major version %v could not be parsed an integer", version)
		}
		if err := a.LaunchMajor(major, args); err != nil {
			return err
		}

	case exactRegex.MatchString(version):
		// Shebang is an exact version e.g. /usr/bin/python3.9
		// we can reuse the parsePyPython functionality here
		a.Logger.WithField("exact version", version).Debugln("Shebang line refers to exact version")
		major, minor, err := a.parsePyPython(version)
		if err != nil {
			return err
		}
		if err := a.LaunchExact(major, minor, args); err != nil {
			return err
		}

	default:
		// The shebang either wasn't valid or had no version identifier e.g. /usr/bin/python
		// in which case, continue the control flow
		a.Logger.WithField("version", version).Debugln("Unrecognised or missing version in shebang line, continuing control flow")
	}

	return nil
}

// launch will launch a python interpreter at a specific (absolute) path
// and forward any args to the called interpreter. If no args required
// just pass an empty slice.
func launch(path string, args []string) error {
	// We must use syscall.Exec here as we must "swap" the process to python
	// simply running a subprocess e.g. (os/exec), even without waiting
	// for the subprocess to complete, will not work as expected

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

// exists returns true if 'path' exists, else false.
func exists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return true
}

// deDupe takes in a list of paths (e.g. those returned from GetPath)
// and returns a de-duplicated list
// it is not that common to have a duplicated $PATH entry but it could happen
// so let's handle it here.
func deDupe(paths []string) []string {
	keys := make(map[string]bool)
	deDuped := []string{}
	for _, item := range paths {
		if _, ok := keys[item]; !ok {
			keys[item] = true
			deDuped = append(deDuped, item)
		}
	}

	return deDuped
}
