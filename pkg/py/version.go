package py

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	pythonExePrefix = "python"
)

// Version represents a version of a python interpreter
// only major and minor are included because this is how the executables
// are stored on disk (e.g. /usr/local/bin/python3.9)
type Version struct {
	Major int
	Minor int
}

// FromFileName extracts the version information from a python interpreter's filename
// and loads the information into the calling `Version`
// If the filename does not start with `python` or does not have a valid
// two digit version after it, an error will be returned
//
// A valid filename will look like `python3.9` or `python3.10`
func (v *Version) FromFileName(filename string) error {
	// Make sure the filename starts with `python`
	if !strings.HasPrefix(filename, pythonExePrefix) {
		return fmt.Errorf("filename is not a valid python interpreter: %s", filename)
	}

	// Because we know the filename will always start with `python`
	// we can split off the version simply by indexing from the end of `python`
	// to the end of the string
	// This is a naive index that doesn't take UTF-8 runes into account but
	// since this is a filepath I think that's pretty safe
	version := filename[len(pythonExePrefix):]

	parts := strings.Split(version, ".")

	// If we can't a part either side of a ".", we have a bad version
	if len(parts) != 2 {
		return fmt.Errorf("malformed interpreter version: %s from filename: %s", version, filename)
	}

	major, minor := parts[0], parts[1]

	majorInt, err := strconv.Atoi(major)
	if err != nil {
		return fmt.Errorf("filename %s major version component could not be parsed as an int: %v", filename, major)
	}

	minorInt, err := strconv.Atoi(minor)
	if err != nil {
		return fmt.Errorf("filename %s minor version component could not be parses as an int: %v", filename, minor)
	}

	v.Major = majorInt
	v.Minor = minorInt

	return nil
}
