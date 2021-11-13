package py

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	pythonExePrefix = "python"
)

// Interpreter represents a version of a python interpreter
// only major and minor are included because this is how the executables
// are stored on disk (e.g. /usr/local/bin/python3.9)
type Interpreter struct {
	Major int    // The interpreter major version e.g. 3
	Minor int    // The interpreter minor version e.g. 10
	Path  string // The absolute path to the interpreter executable
}

// FromFilePath extracts the version information from a python interpreter's filepath
// and loads the information into the calling `Interpreter`
// If the filename does not start with `python` or does not have a valid
// two digit version after it, an error will be returned
//
// A valid filepath will look like `/usr/local/bin/python3.9`
// things like `/usr/local/bin/python` will be rejected as these
// typically refer to the system version of python which should not be used
func (i *Interpreter) FromFilePath(path string) error {
	// Make sure the file name starts with `python`
	filename := filepath.Base(path)
	if !strings.HasPrefix(filename, pythonExePrefix) {
		return fmt.Errorf("filepath is not a valid python interpreter: %s", path)
	}

	// Because we know the filename will always start with `python`
	// we can split off the version simply by indexing from the end of `python`
	// to the end of the string
	// This is a naive index that doesn't take UTF-8 runes into account but
	// since this is a filepath I think that's pretty safe
	// this will also catch things like python-config but we check later for version numbers
	// so this is fine too
	version := filename[len(pythonExePrefix):]

	parts := strings.Split(version, ".")

	// If we can't get a part either side of a ".", we have a bad version
	if len(parts) != 2 {
		return fmt.Errorf("malformed interpreter version: %s from filepath: %s", version, path)
	}

	major, minor := parts[0], parts[1]

	majorInt, err := strconv.Atoi(major)
	if err != nil {
		return fmt.Errorf("filepath %s major version component could not be parsed as an int: %v", path, major)
	}

	minorInt, err := strconv.Atoi(minor)
	if err != nil {
		return fmt.Errorf("filepath %s minor version component could not be parses as an int: %v", path, minor)
	}

	// If we get here, we know we have something like /usr/local/bin/python3.9

	// Ensure the path is absolute
	path, err = filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("could not resolve path %s to absolute: %w", path, err)
	}

	i.Major = majorInt
	i.Minor = minorInt
	i.Path = path

	return nil
}

// String satisfies the "stringer" interface and allows an `Interpreter`
// to be pretty printed using fmt.Println
func (i Interpreter) String() string {
	// Note, the vertical bar character below is not the U+007C "Vertical Line" pipe character
	// '|' but the U+2502 "Box Drawings Light Vertical" character '│'
	// this is so, when printed it looks like a proper table
	return fmt.Sprintf("%d.%d\t│ %s", i.Major, i.Minor, i.Path)
}
