package py

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getPath looks up the $PATH environment variable and will return
// each unique path in a string slice
func getPath() ([]string, error) {
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

	return paths, nil
}

// getPythonInterpreters accepts an absolute path to a directory under which
// it will search for python interpreters, returning any it finds
func getPythonInterpreters(dir string) ([]string, error) {
	contents, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read contents of %s: %w", dir, err)
	}

	interpreterPaths := []string{}

	for _, item := range contents {
		itemPath := filepath.Join(dir, item.Name())
		if isPythonInterpreter(itemPath) {
			interpreterPaths = append(interpreterPaths, itemPath)
		}
	}

	return interpreterPaths, nil
}

// isPythonInterpreter takes a path and returns whether or not the path
// refers to a python interpreter
func isPythonInterpreter(path string) bool {
	return strings.HasPrefix(filepath.Base(path), pythonExePrefix)
}
