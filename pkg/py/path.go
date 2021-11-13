package py

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetAllPythonInterpreters looks under each path in `paths` for valid python
// interpreters and returns the ones it finds
//
// A valid python interpreter in this context is any filepath with a base name
// that starts with `python`
// This is allowed in this context because in usage in this program, `paths` will
// be populated by searching through $PATH, meaning we don't have to bother checking
// if files are executable etc and $PATH is unlikely to be cluttered with random
// files called `python` unless they are the interpreter executables
func GetAllPythonInterpreters(paths []string) (InterpreterList, error) {
	var interpreters InterpreterList

	for _, path := range paths {
		found, err := getPythonInterpreters(path)
		if err != nil {
			return nil, fmt.Errorf("could not fetch interpreters under %s: %w", path, err)
		}
		interpreters = append(interpreters, found...)
	}

	return interpreters, nil
}

// GetPath looks up the $PATH environment variable and will return
// each unique path in a string slice
func GetPath() ([]string, error) {
	path, ok := os.LookupEnv("PATH")
	if !ok {
		// This should literally never happen on any Unix system
		return nil, fmt.Errorf("could not get $PATH")
	}

	paths := []string{}

	for _, dir := range filepath.SplitList(path) {
		if dir == "" {
			// Unix shell semantics: path element "" means "."
			dir = "."
		}
		paths = append(paths, dir)
	}

	// Dedupe
	paths = deDupe(paths)

	return paths, nil
}

// getPythonInterpreters accepts an absolute path to a directory under which
// it will search for python interpreters, returning any it finds
func getPythonInterpreters(dir string) (InterpreterList, error) {
	contents, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read contents of %s: %w", dir, err)
	}

	var interpreters []Interpreter

	for _, item := range contents {
		var interpreter Interpreter
		itemPath := filepath.Join(dir, item.Name())
		if err := interpreter.FromFilePath(itemPath); err == nil {
			// Only add if the interpreter is valid, the others we don't care about
			interpreters = append(interpreters, interpreter)
		}
	}

	return interpreters, nil
}

// deDupe takes in a list of paths (e.g. those returned from GetPath)
// and returns a de-duplicated list
// it is not that common to have a duplicated $PATH entry but it could happen
// so let's handle it here
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
