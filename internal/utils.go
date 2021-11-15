package internal

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GetVenvPython will look for a ".venv/bin/python" or a "venv/bin/python"
// under the cwd, ensure that it exists and then return it's absolute path
// .venv will be preferred over venv, venv will only be used if .venv
// does not exist.
//
// If neither is found, an empty string will be returned
func GetVenvPython(cwd string) string {
	dotVenv := filepath.Join(cwd, ".venv", "bin", "python")
	venv := filepath.Join(cwd, "venv", "bin", "python")

	switch {
	case exists(dotVenv):
		return dotVenv
	case exists(venv):
		return venv
	default:
		return ""
	}
}

// ParsePyPython is a helper that, when given the value of a valid PY_PYTHON env variable
// will return the integer major and minor version parts so we can launch it
//
// A valid value for PY_PYTHON is X.Y, the same as the exact version specifier
// e.g. "3.10"
//
// If 'version' is not a valid format, an error will be returned
func ParsePyPython(version string) (int, int, error) {
	parts := strings.Split(version, ".")

	if len(parts) != 2 {
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

func exists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return true
}
