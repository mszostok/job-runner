package cgroup

import "github.com/spf13/afero"

// SetFS sets used filesystem to a given one. Returns a revert function.
// This function is accessible only in this package and only in test files.
// It won't be complied to a final code.
func SetFS(newFs afero.Fs) func() {
	oldFS := fs
	fs = newFs

	revert := func() {
		fs = oldFS
	}
	return revert
}
