// Package interpreter implements useful utilites for getting, searching,
// sorting and otherwise dealing with python interpreters on $PATH
package interpreter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	pythonExePrefix = "python"
	xYParts         = 2 // Number of parts in an X.Y version
)

// Interpreter represents a version of a python interpreter
// only major and minor are included because this is how the executables
// are stored on disk (e.g. /usr/local/bin/python3.9).
type Interpreter struct {
	Path  string // The absolute path to the interpreter executable
	Major int    // The intepreter major version e.g. 3
	Minor int    // The interpreter minor version e.g. 10
}

// FromFilePath extracts the version information from a python interpreter's filepath
// and loads the information into the calling `Interpreter`
// If the filename does not start with `python` or does not have a valid
// two digit version after it, an error will be returned
//
// A valid filepath will look like `/usr/local/bin/python3.9`
// things like `/usr/local/bin/python` will be rejected as these
// typically refer to the system version of python which should not be used.
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
	// since this is by definition a Unix filepath I think that's pretty safe
	// this will also catch things like `python-config` but we check later for version numbers
	// so this is fine too
	version := filename[len(pythonExePrefix):]

	parts := strings.Split(version, ".")

	// If we can't get a part either side of a ".", we have a bad version
	if len(parts) != xYParts {
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
// to be printed using fmt.Println, in this case showing the absolute path to the interpreter.
func (i Interpreter) String() string {
	return fmt.Sprint(i.Path)
}

// ToString is the pretty print representation of an `Interpreter`
//
// Example
//
//	i := Interpreter{Major: 3, Minor: 10, Path:"/usr/bin/python3.10"}
//	fmt.Println(i.ToString())
//
// Output: "3.10	│ /usr/bin/python3.10".
func (i Interpreter) ToString() string {
	// Note, the vertical bar character below is not the U+007C "Vertical Line" pipe character
	// '|' but the U+2502 "Box Drawings Light Vertical" character '│'
	// this is so, when printed it looks like a proper table
	return fmt.Sprintf("%d.%d\t│ %s", i.Major, i.Minor, i.Path)
}

// SatisfiesMajor tests whether the calling Interpreter satisfies the constraint
// of it's major version supporting the requested `version`.
func (i Interpreter) SatisfiesMajor(version int) bool {
	return i.Major == version
}

// SatisfiesExact tests whether the calling Interpreter satisfies
// the exact version contraint given by `major` and `minor`.
func (i Interpreter) SatisfiesExact(major, minor int) bool {
	return i.Major == major && i.Minor == minor
}

// byVersion represents a list of python interpreters
// and enables us to implement sorting which is how we tell which one is
// the latest python version without relying on filesystem lexical order
// which may not be deterministic.
type byVersion []Interpreter

// Len returns the number of interpreters in the list.
func (bv byVersion) Len() int {
	return len(bv)
}

// Less returns whether the element with index i should sort
// less than element with index j
// Note: we reverse it here and actually test for greater than
// because we want the latest interpreter to be at the front of the slice.
func (bv byVersion) Less(i, j int) bool {
	// Short circuit, if i.Major > j.Major, return true straight away
	if bv[i].Major > bv[j].Major {
		return true
	}

	// Only get here if majors are equal or i.Major < j.Major
	if bv[i].Major == bv[j].Major {
		// If majors are equal, compare minors
		return bv[i].Minor > bv[j].Minor
	}

	// Now only condition remaining is i.Major < j.Major
	// in which case the answer is false
	return false
}

// Swap swaps the position of two elements in the list.
func (bv byVersion) Swap(i, j int) {
	bv[i], bv[j] = bv[j], bv[i]
}

// GetAll looks under each path in `paths` for valid python
// interpreters and returns the ones it finds
//
// This is allowed in this context because in usage in this program, `paths` will
// be populated by searching through $PATH, meaning we don't have to bother checking
// if files are executable etc and $PATH is unlikely to be cluttered with random
// files called `python` unless they are the interpreter executables.
func GetAll(paths []string) ([]Interpreter, error) {
	var interpreters []Interpreter

	for _, path := range paths {
		found, err := getPythonInterpreters(path)
		if err != nil {
			return nil, fmt.Errorf("could not fetch interpreters under %s: %w", path, err)
		}
		interpreters = append(interpreters, found...)
	}

	return interpreters, nil
}

func Sort(interpreters []Interpreter) []Interpreter {
	pythons := interpreters
	sort.Sort(byVersion(pythons))

	return pythons
}

// getPythonInterpreters accepts an absolute path to a directory under which
// it will search for python interpreters, returning any it finds.
func getPythonInterpreters(dir string) ([]Interpreter, error) {
	contents, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read contents of %s: %w", dir, err)
	}

	var interpreters []Interpreter

	for _, item := range contents {
		var interpreter Interpreter
		itemPath := filepath.Join(dir, item.Name())
		if err := interpreter.FromFilePath(itemPath); err == nil {
			// Only add if the interpreter is valid and python3, the others we don't care about
			if interpreter.SatisfiesMajor(3) { //nolint: mnd
				interpreters = append(interpreters, interpreter)
			}
		}
	}

	return interpreters, nil
}
