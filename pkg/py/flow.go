package py

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

const venv = ".venv"

// GetVenvDir will walk up from cwd looking for a directory called ".venv"
// it will then ensure this directory contains a "pyvenv.cfg", the marker
// that this is indeed a python virtual environment, and then return the absolute
// path to the venv's interpreter
//
// If no .venv dir is found, will return an empty string
func GetVenvDir(cwd string) string {
	// First look in the cwd, I imagine most of the time when searching for venvs
	// we'll be in the root of a python project anyway so a lot of calls to this
	// will exit here
	if _, err := os.Stat(filepath.Join(cwd, venv)); errors.Is(err, fs.ErrNotExist) {
		// The .venv dir does not exist, this is not an error
		// but there is no interpreter path to return
		return ""
	}

	// TODO: Currently only looks in cwd which is fine for 90% cases
	// the real python-launcher will walk up the file tree looking for .venv
	// this is on the plan but let's just get this all working first

	return filepath.Join(cwd, venv, "bin", "python")
}
