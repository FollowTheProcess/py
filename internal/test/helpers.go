package test

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// GetProjectRoot is a convenience function for reliably getting the project root dir from anywhere
// so that tests can make use of root-relative paths
func GetProjectRoot() (string, error) {
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("could not find current filepath")
	}

	return filepath.Join(filepath.Dir(here), "../.."), nil
}
