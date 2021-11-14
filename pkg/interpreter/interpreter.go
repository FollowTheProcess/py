package interpreter

import (
	"fmt"
	"os"
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
	// since this is by definition a Unix filepath I think that's pretty safe
	// this will also catch things like `python-config` but we check later for version numbers
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

// SatisfiesMajor tests whether the calling Interpreter satisfies the constraint
// of it's major version supporting the requested `version`
func (i Interpreter) SatisfiesMajor(version int) bool {
	return i.Major == version
}

// SatisfiesExact tests whether the calling Interpreter satisfies
// the exact version contraint given by `major` and `minor`
func (i Interpreter) SatisfiesExact(major, minor int) bool {
	return i.Major == major && i.Minor == minor
}

// InterpreterList represents a list of python interpreters
// and enables us to implement sorting which is how we tell which one is
// the latest python version without relying on filesystem lexical order
// which may not be deterministic
type List []Interpreter

// Len returns the number of interpreters in the list
func (il List) Len() int {
	return len(il)
}

// Less returns whether the element with index i should sort
// less than element with index j
// Note: we reverse it here and actually test for greater than
// because we want the latest interpreter to be at the front of the slice
func (il List) Less(i, j int) bool {
	// Short circuit, if i.Major > j.Major, return true straight away
	if il[i].Major > il[j].Major {
		return true
	}

	// Only get here if majors are equal or i.Major < j.Major
	if il[i].Major == il[j].Major {
		// If majors are equal, compare minors
		return il[i].Minor > il[j].Minor
	}

	// Now only condition remaining is i.Major < j.Major
	// in which case the answer is false
	return false
}

// Swap swaps the position of two elements in the list
func (il List) Swap(i, j int) {
	il[i], il[j] = il[j], il[i]
}

// GetAll looks under each path in `paths` for valid python
// interpreters and returns the ones it finds
//
// A valid python interpreter in this context is any filepath with a base name
// that starts with `python`
// This is allowed in this context because in usage in this program, `paths` will
// be populated by searching through $PATH, meaning we don't have to bother checking
// if files are executable etc and $PATH is unlikely to be cluttered with random
// files called `python` unless they are the interpreter executables
func GetAll(paths []string) (List, error) {
	var interpreters List

	for _, path := range paths {
		found, err := getPythonInterpreters(path)
		if err != nil {
			return nil, fmt.Errorf("could not fetch interpreters under %s: %w", path, err)
		}
		interpreters = append(interpreters, found...)
	}

	return interpreters, nil
}

// getPythonInterpreters accepts an absolute path to a directory under which
// it will search for python interpreters, returning any it finds
func getPythonInterpreters(dir string) (List, error) {
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

// GetPath looks up the $PATH environment variable and will return
// each unique path in a string slice
func GetPath(key string) ([]string, error) {
	path, ok := os.LookupEnv(key)
	if !ok {
		// This should literally never happen on any Unix system
		return nil, fmt.Errorf("could not get $%s", key)
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
